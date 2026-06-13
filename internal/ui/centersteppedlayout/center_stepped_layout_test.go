package centersteppedlayout

import "testing"

func TestAbsNegative(t *testing.T) {
	var x, got, want float32
	x = -64.2465
	got = abs(x)
	want = 64.2465
	if got != want {
		t.Errorf("abs(%f) = %f, want %f", x, got, want)
	}
}

func TestAbsPositive(t *testing.T) {
	var x, got, want float32
	x = 123.543
	got = abs(x)
	want = 123.543
	if got != want {
		t.Errorf("abs(%f) = %f, want %f", x, got, want)
	}
}
