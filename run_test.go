package dockertesting

import (
	"context"
	"errors"
	"testing"
)

func TestTimeoutError_Error(t *testing.T) {
	t.Parallel()
	err := &TimeoutError{
		Operation: "create container",
		Err:       errors.New("context deadline exceeded"),
	}

	expected := "timeout during create container: context deadline exceeded"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestTimeoutError_Unwrap(t *testing.T) {
	t.Parallel()
	innerErr := errors.New("context deadline exceeded")
	err := &TimeoutError{
		Operation: "create container",
		Err:       innerErr,
	}

	if !errors.Is(err, innerErr) {
		t.Error("expected TimeoutError to unwrap to inner error")
	}
}

func TestWrapTimeoutError_NilError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	result := wrapTimeoutError(ctx, nil, "some operation")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestWrapTimeoutError_NonTimeoutError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	innerErr := errors.New("some other error")
	result := wrapTimeoutError(ctx, innerErr, "create network")

	// Should wrap with fmt.Errorf, not TimeoutError
	var timeoutErr *TimeoutError
	if errors.As(result, &timeoutErr) {
		t.Error("expected non-timeout error not to be wrapped as TimeoutError")
	}

	if result.Error() != "failed to create network: some other error" {
		t.Errorf("unexpected error message: %v", result.Error())
	}
}

func TestWrapTimeoutError_DeadlineExceeded(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	// Wait for context to be cancelled
	<-ctx.Done()

	innerErr := errors.New("operation cancelled")
	result := wrapTimeoutError(ctx, innerErr, "execute tests")

	var timeoutErr *TimeoutError
	if !errors.As(result, &timeoutErr) {
		t.Error("expected error to be wrapped as TimeoutError")
	}

	if timeoutErr.Operation != "execute tests" {
		t.Errorf("expected operation 'execute tests', got %q", timeoutErr.Operation)
	}

	if !errors.Is(result, innerErr) {
		t.Error("expected TimeoutError to unwrap to inner error")
	}
}
