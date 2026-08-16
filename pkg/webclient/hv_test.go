package webclient

import (
	"context"
	"os"
	"testing"
	"time"

	"rtlabs.tech/protonsession/pkg/errors"
)

func TestWaitForEnterTimeout(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	c := &WebApiClient{HVInputTimeout: 300 * time.Millisecond}

	start := time.Now()
	err = c.WaitForEnter(context.Background())
	elapsed := time.Since(start)

	if !errors.Is(err, errors.ErrHVInputTimeoutError) {
		t.Fatalf("expected ErrHVInputTimeoutError, got %v", err)
	}
	t.Logf("timed out after %s with %v", elapsed, err)
}

func TestWaitForEnterReceivesInput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	go func() {
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("\n"))
	}()

	c := &WebApiClient{HVInputTimeout: 2 * time.Second}
	if err := c.WaitForEnter(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWaitForEnterContextCancel(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c := &WebApiClient{HVInputTimeout: 10 * time.Second}
	if err := c.WaitForEnter(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
