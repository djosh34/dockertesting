package simple

import "testing"

func TestAdd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
		{10, 20, 30},
	}

	for _, tt := range tests {
		got := Add(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSubtract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{5, 3, 2},
		{0, 0, 0},
		{-1, -1, 0},
		{10, 20, -10},
	}

	for _, tt := range tests {
		got := Subtract(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Subtract(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMultiply(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{2, 3, 6},
		{0, 5, 0},
		{-2, 3, -6},
	}

	for _, tt := range tests {
		got := Multiply(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Multiply(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// Test Divide but intentionally don't test the divide-by-zero case to achieve 70-99% coverage
func TestDivide(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, want int
	}{
		{10, 2, 5},
		{9, 3, 3},
		{-10, 2, -5},
	}

	for _, tt := range tests {
		got := Divide(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Divide(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
	// Note: We intentionally don't test Divide(x, 0) to achieve <100% coverage
}
