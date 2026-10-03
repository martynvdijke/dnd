package search

import "testing"

func TestEntityTypeName(t *testing.T) {
	if EntityTypeName("character") != "Character" {
		t.Fatal("name")
	}
	if EntityTypeName("unknown") != "unknown" {
		t.Fatal("unknown")
	}
	_ = EmptyLegacy()
	_ = EntityURL("character", 1)
	_ = EntityURL("recap", 1)
	_ = EntityURL("unknown", 1)
	_ = LegacyBucketPtr(&SearchResults{}, "character")
	_ = LegacyBucketPtr(&SearchResults{}, "unknown")
	if LegacyBucketPtr(&SearchResults{}, "character") == nil {
		t.Fatal("bucket")
	}
	if LegacyBucketPtr(&SearchResults{}, "nope") != nil {
		t.Fatal("nil")
	}
}

func TestSpellPickerLevelRange(t *testing.T) {
	if _, _, ok := spellPickerLevelRange("0"); !ok {
		t.Fatal("0")
	}
	if _, _, ok := spellPickerLevelRange("1-3"); !ok {
		t.Fatal("1-3")
	}
	if _, _, ok := spellPickerLevelRange("bad"); ok {
		t.Fatal("bad should fail")
	}
}

func TestJsonHelpers(t *testing.T) {
	m := map[string]any{"name": "Bob", "weight": 1.5, "flag": true}
	if jsonEntryString(m, "name") != "Bob" {
		t.Fatal("string")
	}
	if jsonEntryString(map[string]any{"x": 1.0}, "x") != "1" {
		t.Fatal("float string")
	}
	if jsonEntryString(map[string]any{"b": true}, "b") != "true" {
		t.Fatal("bool")
	}
	if jsonEntryString(m, "missing") != "" {
		t.Fatal("missing")
	}
	if jsonEntryNumber(m, "weight") == nil {
		t.Fatal("number")
	}
	if jsonEntryNumber(map[string]any{"weight": "2.5"}, "weight") == nil {
		t.Fatal("string number")
	}
	if jsonEntryNumber(m, "missing") != nil {
		t.Fatal("missing number")
	}
}

func TestEntityURLImpl(t *testing.T) {
	_ = entityURLImpl("character", 1)
	_ = entityURL("character", 1)
}
