package t2

import "testing"

func Test_Sum(t *testing.T) {
	want := 42
	a := 2
	b := 40
	got := a + b
	if got != want {
		t.Error(got, " should be ", want)
	}

}

func Test_Mul(t *testing.T) {
	want := 42
	a := 7
	b := 6
	got := a * b
	if got != want {
		t.Error(got, " should be ", want)
	}
}
