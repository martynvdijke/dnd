package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"
)

func TestFetchFromICalURL_ClientCancellation(t *testing.T) {
	started := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		closeStarted := func() {
			select {
			case <-started:
			default:
				close(started)
			}
		}
		closeStarted()
		// Block until client cancels or timeout.
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}))
	defer srv.Close()

	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type result struct {
		events []struct{}
		err    error
	}
	done := make(chan result, 1)
	go func() {
		_, err := fetchFromICalURL(ctx, settings)
		done <- result{err: err}
	}()

	// Wait for the outbound request to be observed.
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for outbound request to start")
	}

	// Cancel the parent context — fetch should abort promptly.
	cancel()

	select {
	case r := <-done:
		if r.err == nil {
			t.Fatal("expected error after cancellation, got nil")
		}
		if !errors.Is(r.err, context.Canceled) {
			t.Fatalf("expected context.Canceled in error chain, got %v", r.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("fetchFromICalURL did not return promptly after context cancellation")
	}
}

func TestIterateRows_NoLeakOnEarlyReturn(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	if db.DB == nil {
		t.Skip("db not initialised")
	}

	// Seed a small table so Query returns rows.
	if _, err := db.DB.Exec(`CREATE TABLE IF NOT EXISTS _ctx_test (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	for i := 1; i <= 3; i++ {
		if _, err := db.DB.Exec(`INSERT INTO _ctx_test(v) VALUES(?)`, "x"); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	prev := db.DB.Stats().MaxOpenConnections
	db.DB.SetMaxOpenConns(1)
	defer db.DB.SetMaxOpenConns(prev)

	errCh := make(chan error, 1)
	go func() {
		var lastErr error
		// Run enough iterations that a leak would exhaust the pool (MaxOpenConns=1).
		for i := 0; i < 6; i++ {
			rows, err := db.DB.Query(`SELECT id, v FROM _ctx_test`)
			if err != nil {
				lastErr = err
				break
			}
			// Simulate early-return: fail on first row scan.
			err = iterateRows(rows, func() error {
				return errors.New("early abort")
			})
			if err == nil {
				lastErr = errors.New("expected early abort error")
				break
			}
		}
		if lastErr != nil {
			errCh <- lastErr
			return
		}
		// Final query must still succeed if no connection was leaked.
		var n int
		if err := db.DB.QueryRow(`SELECT 1`).Scan(&n); err != nil {
			errCh <- err
			return
		}
		if n != 1 {
			errCh <- errors.New("unexpected SELECT 1 result")
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("leak check failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timed out - likely leaked *sql.Rows holds the only pooled connection")
	}
}
