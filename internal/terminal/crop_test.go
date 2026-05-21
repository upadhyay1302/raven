package terminal

import "testing"

func Test_CropPreservingANSI(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		maxCols  int
		expected string
	}{
		{"plain text fits", "Hello", 10, "Hello"},
		{"plain text crops", "Hello World", 5, "Hello"},
		{"ansi preserved", "\x1b[91mHello\x1b[0m", 3, "\x1b[91mHel\x1b[0m"},
		{"empty string", "", 10, ""},
		{"zero max", "Hello", 0, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CropPreservingANSI(tc.input, tc.maxCols)
			if got != tc.expected {
				t.Errorf("expected %q got %q", tc.expected, got)
			}
		})
	}
}