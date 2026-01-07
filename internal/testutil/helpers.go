package testutil

import (
	"context"
	"testing"
	"time"
)

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("assertion failed: %s", message)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("assertion failed: %s", message)
	}
}

// AssertPanics fails if the function f does not panic
func AssertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, but function did not panic")
		}
	}()
	f()
}

// WithTimeout creates a context with timeout for testing
func WithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// WithCancel creates a cancellable context for testing
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// AssertContextCancelled fails if context is not cancelled
func AssertContextCancelled(t *testing.T, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
		// Expected - context is cancelled
	default:
		t.Fatal("expected context to be cancelled, but it is not")
	}
}

// AssertContextNotCancelled fails if context is cancelled
func AssertContextNotCancelled(t *testing.T, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
		t.Fatal("expected context to not be cancelled, but it is")
	default:
		// Expected - context is not cancelled
	}
}

// WaitOrTimeout waits for a channel or times out
func WaitOrTimeout(t *testing.T, ch <-chan struct{}, timeout time.Duration, message string) {
	t.Helper()
	select {
	case <-ch:
		// Success
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for %s", message)
	}
}

// AssertSliceLength checks slice length
func AssertSliceLength(t *testing.T, slice interface{}, expectedLen int) {
	t.Helper()
	var actualLen int
	switch s := slice.(type) {
	case []interface{}:
		actualLen = len(s)
	case []string:
		actualLen = len(s)
	default:
		t.Fatalf("unsupported slice type: %T", slice)
	}
	if actualLen != expectedLen {
		t.Fatalf("expected slice length %d, got %d", expectedLen, actualLen)
	}
}
