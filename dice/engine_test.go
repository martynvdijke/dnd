package dice

import (
	"fmt"
	"testing"
	"testing/quick"
)

func TestPoolBasic(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("2d6+3")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
	if result.Notation != "2d6+3" {
		t.Errorf("expected notation '2d6+3', got %q", result.Notation)
	}
	if result.Total == "" {
		t.Error("expected non-empty total")
	}
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestKeepHighest(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("4d6kh3")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
	if result.Notation != "4d6kh3" {
		t.Errorf("expected notation '4d6kh3', got %q", result.Notation)
	}
}

func TestExplodingDice(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("3d6!")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
}

func TestReroll(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("1d20r=1")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
}

func TestSuccessCounting(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("8d10>7")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
}

func TestInvalidExpression(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	_, err = p.Roll("invalid!!")
	if err == nil {
		t.Fatal("expected error for invalid expression, got nil")
	}
}

func TestMultipleRolls(t *testing.T) {
	p, err := NewPool(2)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	// Run multiple rolls
	for i := range 5 {
		result, err := p.Roll("1d20")
		if err != nil {
			t.Fatalf("roll %d failed: %v", i, err)
		}
		if result.Error != "" {
			t.Fatalf("roll %d error: %s", i, result.Error)
		}
	}
}

func TestPercentile(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("d100")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
}

func TestBasicD20(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	result, err := p.Roll("1d20")
	if err != nil {
		t.Fatalf("Roll failed: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error: %s", result.Error)
	}
	// Total should be 1-20
	t.Logf("d20 roll: notation=%s total=%s output=%s", result.Notation, result.Total, result.Output)
}

func TestPropertyDiceRollBounds(t *testing.T) {
	p, err := NewPool(2)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	f := func(n, sides int) bool {
		if n < 1 || n > 10 || sides < 2 || sides > 100 {
			return true
		}
		expr := fmt.Sprintf("%dd%d", n, sides)
		result, err := p.Roll(expr)
		if err != nil {
			return false
		}
		total, _ := result.Total.Int64()
		return total >= int64(n) && total <= int64(n*sides)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func FuzzDiceExpression(f *testing.F) {
	f.Add("2d6+3")
	f.Add("1d20")
	f.Add("invalid!!")
	f.Add("")
	f.Add("3d6kh2")
	f.Add("1d20r=1")
	f.Fuzz(func(t *testing.T, expr string) {
		p, err := NewPool(1)
		if err != nil {
			t.Skipf("pool init: %v", err)
		}
		result, err := p.Roll(expr)
		if err == nil && result.Error != "" {
			_ = result.Error
		}
	})
}

func TestRoll_2d6(t *testing.T) {
	pool, err := NewPool(1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	result, err := pool.Roll("2d6")
	if err != nil {
		t.Fatalf("Roll 2d6: %v", err)
	}
	n, err := result.Total.Int64()
	if err != nil {
		t.Fatalf("Total not int: %v (%s)", err, result.Total)
	}
	if n < 2 || n > 12 {
		t.Fatalf("Total %d out of range [2,12] for 2d6", n)
	}
}
