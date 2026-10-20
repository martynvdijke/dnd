package handlers

import (
	"os"
	"testing"

	"villum/middleware"
)

// TestMain runs the handler suite with the password break-glass flag enabled.
//
// The handler tests exercise the break-glass path: fixtures create password
// sessions with middleware.Store.Create and drive real routes through
// middleware.AuthRequired (e.g. tokens_test.go), and the shared login helpers
// authenticate with a password. OIDC is the default login, so password login
// and password sessions are refused unless ALLOW_PASSWORD_FALLBACK is on; it is
// enabled here once, for the whole package, rather than sprinkled through every
// test. Tests that assert the retired behaviour (password login or password
// sessions rejected while the flag is off) must turn it back off themselves
// with t.Setenv(middleware.PasswordFallbackEnvVar, "false").
func TestMain(m *testing.M) {
	os.Setenv(middleware.PasswordFallbackEnvVar, "true")
	os.Exit(m.Run())
}
