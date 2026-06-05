package middleware

import (
	"context"
	"errors"
	"testing"
)

func TestTraceDBQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		called := false
		err := TraceDBQuery(ctx, "test-op", func(ctx context.Context) error {
			called = true
			return nil
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if !called {
			t.Error("function was not called")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		err := TraceDBQuery(ctx, "test-op-err", func(ctx context.Context) error {
			return errors.New("db error")
		})
		if err == nil || err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err)
		}
	})
}
