package dice

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

// TestDiceContract asserts the Go<->JS contract documented in
// dice/roller-entry.js: both globals exist and are callable, and __diceRoll
// returns the expected JSON shape.
func TestDiceContract(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	eng := p.Acquire()
	if eng == nil {
		t.Fatal("Acquire returned nil engine")
	}
	defer p.Release(eng)

	rollVal := eng.vm.Get("__diceRoll")
	if rollVal == nil || goja.IsUndefined(rollVal) {
		t.Fatal("contract broken: globalThis.__diceRoll is missing")
	}
	if _, ok := goja.AssertFunction(rollVal); !ok {
		t.Fatal("contract broken: globalThis.__diceRoll is not callable")
	}

	rollerVal := eng.vm.Get("__diceRoller")
	if rollerVal == nil || goja.IsUndefined(rollerVal) {
		t.Fatal("contract broken: globalThis.__diceRoller is missing")
	}

	raw, err := eng.rollFn(goja.Undefined(), eng.vm.ToValue("2d6"))
	if err != nil {
		t.Fatalf(`__diceRoll("2d6") error: %v`, err)
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw.String()), &shape); err != nil {
		t.Fatalf("__diceRoll did not return JSON: %v (%q)", err, raw.String())
	}
	if _, ok := shape["error"]; ok {
		t.Fatalf("__diceRoll returned an error: %s", raw.String())
	}
	for _, field := range []string{"notation", "total", "rolls", "output"} {
		if _, ok := shape[field]; !ok {
			t.Errorf("contract broken: __diceRoll JSON missing %q field: %s", field, raw.String())
		}
	}
}

// TestBundleStampMatchesPin keeps the Go expected version, the package.json pin,
// and the embedded bundle stamp in sync. A dependency bump that updates only one
// of the three fails here.
func TestBundleStampMatchesPin(t *testing.T) {
	data, err := os.ReadFile("../package.json")
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatalf("parse package.json: %v", err)
	}
	pin := pkg.Dependencies["@dice-roller/rpg-dice-roller"]
	if pin != expectedDiceVersion {
		t.Fatalf("package.json pins @dice-roller/rpg-dice-roller %q but dice/engine.go expects %q — update expectedDiceVersion to match", pin, expectedDiceVersion)
	}
	if strings.ContainsAny(pin, "^~") {
		t.Fatalf("dependency must be pinned to an exact version, got %q", pin)
	}
	stamp, err := bundleBuildStamp()
	if err != nil {
		t.Fatalf("read bundle stamp: %v", err)
	}
	if stamp != expectedDiceVersion {
		t.Fatalf("embedded bundle stamp %q != expected %q — rebuild with `npm run build:dice`", stamp, expectedDiceVersion)
	}
}

// TestStaleBundleDetected simulates the bump-without-rebuild case: the expected
// version no longer matches the bundle stamp, so engine init must fail fast
// rather than silently using the stale bundle.
func TestStaleBundleDetected(t *testing.T) {
	old := expectedDiceVersion
	expectedDiceVersion = "0.0.0-stale"
	defer func() { expectedDiceVersion = old }()

	_, err := NewPool(1)
	if err == nil {
		t.Fatal("expected NewPool to fail on a build-stamp mismatch")
	}
	if !strings.Contains(err.Error(), "build stamp mismatch") {
		t.Fatalf("expected a build stamp mismatch error, got: %v", err)
	}
}
