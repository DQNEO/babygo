package strings

import "testing"

func TestContains(t *testing.T) {
	type args struct {
		s      string
		substr string
	}
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"simple-contain-mid", "abcde", "bcd", true},
		{"simple-contain-head", "abcde", "ab", true},
		{"simple-contain-tail", "abcde", "de", true},
		{"simple-contain-mind-1", "abcde", "c", true},
		{"simple-contain-head-1", "abcde", "a", true},
		{"simple-contain-tail-1", "abcde", "e", true},
		{"simple-not-contain", "abcde", "xyz", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
