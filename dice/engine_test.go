package dice

import (
	"testing"
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
	for i := 0; i < 5; i++ {
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
