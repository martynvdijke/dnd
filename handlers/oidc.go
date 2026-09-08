package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"villum/db"
	"villum/middleware"
)

// OIDC login via Authelia (Authorization Code + PKCE S256). Sessions reuse the
// same `session` cookie/store as password auth; only auth_method differs.
// Disabled by default (OIDC_ENABLED=true to enable); password login stays as fallback.

// OIDCConfig is the OIDC relying-party config, read from env. The client
// secret comes from OIDC_CLIENT_SECRET_FILE (preferred) or OIDC_CLIENT_SECRET
// and is never logged or committed.
type OIDCConfig struct {
	Enabled     bool
	Issuer      string
	ClientID    string
	Secret      string
	RedirectURL string
	Scopes      []string
	LogoutURL   string
}

func loadOIDCConfig() OIDCConfig {
	cfg := OIDCConfig{Enabled: os.Getenv("OIDC_ENABLED") == "true"}
	if f := os.Getenv("OIDC_CLIENT_SECRET_FILE"); f != "" {
		if b, err := os.ReadFile(f); err == nil {
			cfg.Secret = strings.TrimSpace(string(b))
		}
	}
	if cfg.Secret == "" {
		cfg.Secret = os.Getenv("OIDC_CLIENT_SECRET")
	}
	cfg.Issuer = strings.TrimSuffix(os.Getenv("OIDC_ISSUER_URL"), "/")
	cfg.ClientID = os.Getenv("OIDC_CLIENT_ID")
	cfg.RedirectURL = os.Getenv("OIDC_REDIRECT_URL")
	if s := os.Getenv("OIDC_SCOPES"); s != "" {
		cfg.Scopes = strings.Fields(s)
	} else {
		cfg.Scopes = []string{"openid", "email", "profile", "groups"}
	}
	if u := os.Getenv("OIDC_LOGOUT_URL"); u != "" {
		cfg.LogoutURL = u
	} else {
		cfg.LogoutURL = cfg.Issuer + "/logout"
	}
	return cfg
}

func (c OIDCConfig) valid() bool {
	return c.Enabled && c.Issuer != "" && c.ClientID != "" && c.Secret != "" && c.RedirectURL != ""
}

// oidcProvider caches discovery + verifier per issuer/client.
type oidcProvider struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
	issuer   string
	clientID string
}

var (
	oidcMu    sync.Mutex
	oidcCache *oidcProvider
)

func getOIDCProvider(ctx context.Context, cfg OIDCConfig) (*oidcProvider, error) {
	oidcMu.Lock()
	defer oidcMu.Unlock()
	if oidcCache != nil && oidcCache.issuer == cfg.Issuer && oidcCache.clientID == cfg.ClientID {
		return oidcCache, nil
	}
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oidcCache = &oidcProvider{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.Secret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
		},
		issuer:   cfg.Issuer,
		clientID: cfg.ClientID,
	}
	return oidcCache, nil
}

func oidcRandHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func setOIDCCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", false, true)
}

// OIDCStatus reports whether OIDC login is enabled (drives the login page button).
func OIDCStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": loadOIDCConfig().Enabled})
}

// OIDCLogin starts the Authorization Code + PKCE flow (state/nonce/verifier cookies).
func OIDCLogin(c *gin.Context) {
	cfg := loadOIDCConfig()
	if !cfg.Enabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "oidc disabled"})
		return
	}
	if !cfg.valid() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oidc misconfigured"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	p, err := getOIDCProvider(ctx, cfg)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oidc discovery failed"})
		return
	}
	state, err := oidcRandHex(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}
	nonce, err := oidcRandHex(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}
	verifier := oauth2.GenerateVerifier()
	setOIDCCookie(c, "oidc_state", state, 300)
	setOIDCCookie(c, "oidc_nonce", nonce, 300)
	setOIDCCookie(c, "oidc_verifier", verifier, 300)
	url := p.oauth.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
	)
	c.Redirect(http.StatusFound, url)
}

// oidcClaims is the subset of ID token claims Villum uses.
type oidcClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_username"`
	Nonce         string `json:"nonce"`
}

func clearOIDCCookies(c *gin.Context) {
	for _, n := range []string{"oidc_state", "oidc_nonce", "oidc_verifier"} {
		c.SetCookie(n, "", -1, "/", "", false, true)
	}
}

func oidcFail(c *gin.Context, msg string) {
	clearOIDCCookies(c)
	c.Redirect(http.StatusFound, "/login?error="+msg)
	c.Abort()
}

// OIDCCallback verifies state/nonce/PKCE + ID token, links/provisions the user,
// creates a session with the same cookie shape as password auth, and redirects to /.
func OIDCCallback(c *gin.Context) {
	cfg := loadOIDCConfig()
	if !cfg.Enabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "oidc disabled"})
		return
	}
	if !cfg.valid() {
		oidcFail(c, "oidc_config")
		return
	}
	state, err1 := c.Cookie("oidc_state")
	nonce, err2 := c.Cookie("oidc_nonce")
	verifier, err3 := c.Cookie("oidc_verifier")
	if err1 != nil || err2 != nil || err3 != nil || state == "" || nonce == "" || verifier == "" {
		oidcFail(c, "oidc_expired")
		return
	}
	if q := c.Query("state"); q == "" || subtle.ConstantTimeCompare([]byte(q), []byte(state)) != 1 {
		oidcFail(c, "oidc_state")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	p, err := getOIDCProvider(ctx, cfg)
	if err != nil {
		oidcFail(c, "oidc_provider")
		return
	}
	token, err := p.oauth.Exchange(ctx, c.Query("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		oidcFail(c, "oidc_exchange")
		return
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		oidcFail(c, "oidc_token")
		return
	}
	idToken, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		oidcFail(c, "oidc_verify")
		return
	}
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		oidcFail(c, "oidc_claims")
		return
	}
	if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(nonce)) != 1 {
		oidcFail(c, "oidc_nonce")
		return
	}
	if claims.Email == "" || !claims.EmailVerified {
		oidcFail(c, "oidc_email")
		return
	}
	// Groups claim presence (vs value) decides whether the admin flag syncs:
	// absent claim leaves the role untouched so a group-less token can't demote admins.
	var rawClaims map[string]any
	groups := []string{}
	groupsPresent := false
	if err := idToken.Claims(&rawClaims); err == nil {
		if g, ok := rawClaims["groups"]; ok && g != nil {
			groupsPresent = true
			if arr, ok := g.([]any); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						groups = append(groups, s)
					}
				}
			}
		}
	}
	sub := cfg.Issuer + "|" + claims.Sub
	u, err := linkOrProvisionOIDCUser(c.Request.Context(), sub, claims, groups, groupsPresent)
	if err != nil {
		oidcFail(c, "oidc_user")
		return
	}
	clearOIDCCookies(c)
	sessionID := middleware.Store.CreateWithAuth(u.id, u.username, u.role, c.ClientIP(), "oidc", sub)
	c.SetCookie("session", sessionID, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// OIDCLogout clears the local session (same cookie semantics as password
// logout) and redirects to the IdP logout so SSO ends there too.
func OIDCLogout(c *gin.Context) {
	if sessionID, _ := c.Cookie("session"); sessionID != "" {
		middleware.Store.Delete(sessionID)
	}
	c.SetCookie("session", "", -1, "/", "", false, true)
	if cfg := loadOIDCConfig(); cfg.Enabled && cfg.Issuer != "" {
		c.Redirect(http.StatusFound, cfg.LogoutURL)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type oidcUser struct {
	id       int64
	username string
	role     string
}

// oidcRoleForGroups maps the IdP groups claim to a Villum role.
func oidcRoleForGroups(groups []string) string {
	for _, g := range groups {
		if g == "admins" {
			return "admin"
		}
	}
	return "user"
}

func scanOIDCUser(row *sql.Row) (*oidcUser, error) {
	u := &oidcUser{}
	var sub sql.NullString
	if err := row.Scan(&u.id, &u.username, &u.role, &sub); err != nil {
		return nil, err
	}
	return u, nil
}

// linkOrProvisionOIDCUser links a verified OIDC identity to an existing user by
// oidc_sub, then by verified email, else provisions a new user. The admin flag
// syncs from the groups claim only when the claim is present.
func linkOrProvisionOIDCUser(ctx context.Context, sub string, claims oidcClaims, groups []string, groupsPresent bool) (*oidcUser, error) {
	if u, err := scanOIDCUser(db.DB.QueryRowContext(ctx,
		`SELECT id, username, role, oidc_sub FROM users WHERE oidc_sub = ?`, sub)); err == nil {
		return maybeSyncOIDCRole(ctx, u, groups, groupsPresent)
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	if u, err := scanOIDCUser(db.DB.QueryRowContext(ctx,
		`SELECT id, username, role, oidc_sub FROM users WHERE email = ?`, claims.Email)); err == nil {
		if _, err := db.DB.ExecContext(ctx, `UPDATE users SET oidc_sub = ? WHERE id = ?`, sub, u.id); err != nil {
			return nil, err
		}
		return maybeSyncOIDCRole(ctx, u, groups, groupsPresent)
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	base := claims.PreferredName
	if base == "" {
		base = claims.Name
	}
	if base == "" {
		if i := strings.Index(claims.Email, "@"); i > 0 {
			base = claims.Email[:i]
		} else {
			base = claims.Email
		}
	}
	username := uniquifyUsername(ctx, base)
	display := claims.Name
	if display == "" {
		display = username
	}
	role := "user"
	if groupsPresent {
		role = oidcRoleForGroups(groups)
	}
	pw, err := oidcRandHex(16)
	if err != nil {
		return nil, err
	}
	// Unusable password marker: bcrypt comparison always fails, so OIDC-only
	// accounts can never log in via the password flow.
	res, err := db.DB.ExecContext(ctx,
		`INSERT INTO users(username, password, display_name, role, email, created_at, oidc_sub) VALUES(?, ?, ?, ?, ?, datetime('now'), ?)`,
		username, "oidc$"+pw, display, role, claims.Email, sub)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &oidcUser{id: id, username: username, role: role}, nil
}

func maybeSyncOIDCRole(ctx context.Context, u *oidcUser, groups []string, groupsPresent bool) (*oidcUser, error) {
	if !groupsPresent {
		return u, nil
	}
	role := oidcRoleForGroups(groups)
	if role == u.role {
		return u, nil
	}
	if _, err := db.DB.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, u.id); err != nil {
		return nil, err
	}
	u.role = role
	return u, nil
}

func uniquifyUsername(ctx context.Context, base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "user"
	}
	candidate := base
	for i := 2; ; i++ {
		var one int
		if err := db.DB.QueryRowContext(ctx, `SELECT 1 FROM users WHERE username = ?`, candidate).Scan(&one); err == sql.ErrNoRows {
			return candidate
		} else if err != nil {
			return candidate
		}
		candidate = base + "-" + strconv.Itoa(i)
	}
}
