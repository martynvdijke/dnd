package handlers

import (
	"errors"
	"regexp"
	"strings"
)

// validIdentifierRe matches a bare identifier: ASCII letters, digits and
// underscores. It is deliberately stricter than SQLite needs — these helpers
// gate the few values that cannot be passed as bound parameters (sort field
// names becomes the final JSON-path key, plus the ASC/DESC keyword), so the
// accepted set stays small and predictable.
var validIdentifierRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ErrInvalidIdentifier is returned when a client-supplied value is not a simple
// identifier. Callers should surface it as a 400 and execute no SQL.
var ErrInvalidIdentifier = errors.New("invalid identifier")

// validIdentifier reports whether s is safe to use as a SQL identifier
// component or as the final key of a json_extract path.
func validIdentifier(s string) bool {
	return validIdentifierRe.MatchString(s)
}

// jsonPathForField converts a client-supplied field name into a json_extract
// path (e.g. `$."name"`). The returned path is meant to be passed as a BOUND
// PARAMETER, never concatenated into SQL text — SQLite accepts a bind value for
// the json_extract path argument.
func jsonPathForField(field string) (string, error) {
	if !validIdentifier(field) {
		return "", ErrInvalidIdentifier
	}
	return `$."` + field + `"`, nil
}

// orderDirection maps a client-supplied order parameter to the only two SQL
// keywords that may ever be concatenated into a query. Anything that is not
// "desc" (case-insensitive) yields "ASC".
func orderDirection(dir string) string {
	if strings.EqualFold(strings.TrimSpace(dir), "desc") {
		return "DESC"
	}
	return "ASC"
}
