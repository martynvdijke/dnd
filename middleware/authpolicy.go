package middleware

import (
	"os"
	"strings"
)

// PasswordFallbackEnvVar is the break-glass switch for password login.
//
// OIDC (Authelia) is the default login, so password login is disabled unless an
// operator opts back in with this variable set to `true` or `1`. It exists for
// recovery only: if Authelia is unreachable or misconfigured, setting it and
// restarting restores password login without a code deploy.
const PasswordFallbackEnvVar = "ALLOW_PASSWORD_FALLBACK"

// PasswordFallbackEnabled reports whether password login may be used.
//
// It is read per call rather than cached so a running process reflects its
// environment and tests can flip the value with t.Setenv.
func PasswordFallbackEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(PasswordFallbackEnvVar))) {
	case "true", "1":
		return true
	}
	return false
}
