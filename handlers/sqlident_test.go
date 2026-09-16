package handlers

import (
	"errors"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestValidIdentifier(t *testing.T) {
	valid := []string{"name", "a1", "_x", "CamelCase", "some_field_2", "ID"}
	invalid := []string{"", "a b", "a.b", `a"b`, "a'b", "a;b", "a--b", "a(b", "a)b", "$.name", "a/b", "a\\b", "naïve"}

	for _, s := range valid {
		if !validIdentifier(s) {
			t.Errorf("validIdentifier(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		if validIdentifier(s) {
			t.Errorf("validIdentifier(%q) = true, want false", s)
		}
	}
}

func TestJSONPathForField(t *testing.T) {
	tests := []struct {
		field   string
		want    string
		wantErr error
	}{
		{field: "name", want: `$."name"`},
		{field: "display_name", want: `$."display_name"`},
		{field: `na"me`, wantErr: ErrInvalidIdentifier},
		{field: "na'me", wantErr: ErrInvalidIdentifier},
		{field: "name;DROP TABLE x", wantErr: ErrInvalidIdentifier},
		{field: "name--", wantErr: ErrInvalidIdentifier},
		{field: "name(", wantErr: ErrInvalidIdentifier},
		{field: "", wantErr: ErrInvalidIdentifier},
	}
	for _, tt := range tests {
		got, err := jsonPathForField(tt.field)
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("jsonPathForField(%q) err = %v, want %v", tt.field, err, tt.wantErr)
			}
			if got != "" {
				t.Errorf("jsonPathForField(%q) = %q on error, want empty", tt.field, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("jsonPathForField(%q) unexpected error: %v", tt.field, err)
			continue
		}
		if got != tt.want {
			t.Errorf("jsonPathForField(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func TestOrderDirection(t *testing.T) {
	tests := map[string]string{
		"desc": "DESC", "DESC": "DESC", " Desc ": "DESC",
		"asc": "ASC", "ASC": "ASC", "": "ASC", "bogus": "ASC", "1;DROP": "ASC",
	}
	for in, want := range tests {
		if got := orderDirection(in); got != want {
			t.Errorf("orderDirection(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestListCompendiumEntriesInvalidSortReturns400 proves an invalid sort field is
// rejected with 400 before any SQL runs (jsonPathForField rejects it).
func TestListCompendiumEntriesInvalidSortReturns400(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/schemas/:id/entries", ListCompendiumEntries)
	})

	for _, bad := range []string{`name"`, "name;DROP TABLE compendium_entries", "name--", "name(", "name'", "a b"} {
		w := testutil.Get(t, r, "/api/compendium/schemas/1/entries?sort="+url.QueryEscape(bad))
		testutil.AssertStatus(t, w, 400)
	}

	// A well-formed field name is accepted (empty result set, 200).
	w := testutil.Get(t, r, "/api/compendium/schemas/1/entries?sort=name&order=desc")
	testutil.AssertStatus(t, w, 200)
}
