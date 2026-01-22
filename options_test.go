package dockertesting

import (
	"testing"
	"time"
)

func TestNewOptions_RequiresPackagePath(t *testing.T) {
	t.Parallel()
	_, err := NewOptions("")
	if err == nil {
		t.Error("expected error for empty package path, got nil")
	}
}

func TestNewOptions_DefaultValues(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PackagePath != "/path/to/package" {
		t.Errorf("expected PackagePath '/path/to/package', got %q", opts.PackagePath)
	}
	if opts.Pattern != DefaultPattern {
		t.Errorf("expected Pattern %q, got %q", DefaultPattern, opts.Pattern)
	}
	if opts.SockPath != DefaultSockPath {
		t.Errorf("expected SockPath %q, got %q", DefaultSockPath, opts.SockPath)
	}
	if opts.EnableVarSock {
		t.Error("expected EnableVarSock to be false by default")
	}
	if len(opts.Args) != 0 {
		t.Errorf("expected empty Args, got %v", opts.Args)
	}
	if len(opts.Aliases) != 0 {
		t.Errorf("expected empty Aliases, got %v", opts.Aliases)
	}
}

func TestWithPattern(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithPattern("./pkg/..."))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Pattern != "./pkg/..." {
		t.Errorf("expected Pattern './pkg/...', got %q", opts.Pattern)
	}
}

func TestWithArgs(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithArgs("-v", "-race"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.Args) != 2 {
		t.Fatalf("expected 2 Args, got %d", len(opts.Args))
	}
	if opts.Args[0] != "-v" {
		t.Errorf("expected Args[0] '-v', got %q", opts.Args[0])
	}
	if opts.Args[1] != "-race" {
		t.Errorf("expected Args[1] '-race', got %q", opts.Args[1])
	}
}

func TestWithArgs_Multiple(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithArgs("-v"), WithArgs("-race"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.Args) != 2 {
		t.Fatalf("expected 2 Args, got %d", len(opts.Args))
	}
	if opts.Args[0] != "-v" {
		t.Errorf("expected Args[0] '-v', got %q", opts.Args[0])
	}
	if opts.Args[1] != "-race" {
		t.Errorf("expected Args[1] '-race', got %q", opts.Args[1])
	}
}

func TestWithAliases(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithAliases("myapp.test", "api.test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.Aliases) != 2 {
		t.Fatalf("expected 2 Aliases, got %d", len(opts.Aliases))
	}
	if opts.Aliases[0] != "myapp.test" {
		t.Errorf("expected Aliases[0] 'myapp.test', got %q", opts.Aliases[0])
	}
	if opts.Aliases[1] != "api.test" {
		t.Errorf("expected Aliases[1] 'api.test', got %q", opts.Aliases[1])
	}
}

func TestWithAliases_Multiple(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithAliases("myapp.test"), WithAliases("api.test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts.Aliases) != 2 {
		t.Fatalf("expected 2 Aliases, got %d", len(opts.Aliases))
	}
}

func TestWithVarSock(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithVarSock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !opts.EnableVarSock {
		t.Error("expected EnableVarSock to be true")
	}
}

func TestWithSockPath(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithSockPath("/custom/docker.sock"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.SockPath != "/custom/docker.sock" {
		t.Errorf("expected SockPath '/custom/docker.sock', got %q", opts.SockPath)
	}
}

func TestWithMultipleOptions(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions(
		"/path/to/package",
		WithPattern("./cmd/..."),
		WithArgs("-v", "-count=1"),
		WithAliases("myapp.test"),
		WithVarSock(),
		WithSockPath("/custom/docker.sock"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PackagePath != "/path/to/package" {
		t.Errorf("expected PackagePath '/path/to/package', got %q", opts.PackagePath)
	}
	if opts.Pattern != "./cmd/..." {
		t.Errorf("expected Pattern './cmd/...', got %q", opts.Pattern)
	}
	if len(opts.Args) != 2 {
		t.Fatalf("expected 2 Args, got %d", len(opts.Args))
	}
	if len(opts.Aliases) != 1 {
		t.Fatalf("expected 1 Alias, got %d", len(opts.Aliases))
	}
	if !opts.EnableVarSock {
		t.Error("expected EnableVarSock to be true")
	}
	if opts.SockPath != "/custom/docker.sock" {
		t.Errorf("expected SockPath '/custom/docker.sock', got %q", opts.SockPath)
	}
}

func TestNewOptions_DefaultTimeout(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, opts.Timeout)
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()
	customTimeout := 5 * time.Minute
	opts, err := NewOptions("/path/to/package", WithTimeout(customTimeout))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Timeout != customTimeout {
		t.Errorf("expected Timeout %v, got %v", customTimeout, opts.Timeout)
	}
}

func TestWithTimeout_ZeroDisablesTimeout(t *testing.T) {
	t.Parallel()
	opts, err := NewOptions("/path/to/package", WithTimeout(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Timeout != 0 {
		t.Errorf("expected Timeout 0, got %v", opts.Timeout)
	}
}

func TestWithTimeout_ShortDuration(t *testing.T) {
	t.Parallel()
	shortTimeout := 30 * time.Second
	opts, err := NewOptions("/path/to/package", WithTimeout(shortTimeout))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Timeout != shortTimeout {
		t.Errorf("expected Timeout %v, got %v", shortTimeout, opts.Timeout)
	}
}
