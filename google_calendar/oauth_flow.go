package googlecalendar

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// OAuthConfig returns an oauth2.Config for Google Calendar OAuth.
func OAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarReadonlyScope},
	}
}

// AuthCodeURL returns the Google consent URL with offline access, prompt=consent and PKCE S256.
func AuthCodeURL(cfg *oauth2.Config, state, verifier string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"), oauth2.S256ChallengeOption(verifier))
}

// ExchangeCode exchanges an authorization code for a token using the PKCE verifier.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
	return cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}
