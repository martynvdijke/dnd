package dice

import "testing"

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
