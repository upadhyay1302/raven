package raven

import "testing"

func Test_Level_StringsAreNonEmpty(t *testing.T) {
	usedStrings := make(map[string]Level)

	for l := levelMin; l < levelMax; l++ {
		str := l.String()

		if len(str) == 0 {
			t.Errorf("Level %d has an empty String()", int(l))
		}

		if existing, seen := usedStrings[str]; seen {
			t.Errorf("Level %d and Level %d share the same string %q — all levels must be unique",
				int(existing), int(l), str)
		}
		usedStrings[str] = l
	}
}

func Test_Level_StringValues(t *testing.T) {
	cases := []struct {
		level    Level
		expected string
	}{
		{Transient, "transient"},
		{Verbose, "verbose"},
		{Info, "info"},
		{Warning, "warning"},
		{Error, "error"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			if got := tc.level.String(); got != tc.expected {
				t.Errorf("expected %q got %q", tc.expected, got)
			}
		})
	}
}

func Test_Level_UnknownString(t *testing.T) {
	unknown := levelMax // one past the valid range
	if got := unknown.String(); got != "unknown" {
		t.Errorf("expected out-of-range level to return \"unknown\", got %q", got)
	}
}

func Test_Level_IsValid(t *testing.T) {
	cases := []struct {
		level    Level
		expected bool
	}{
		{Transient, true},
		{Verbose, true},
		{Info, true},
		{Warning, true},
		{Error, true},
		{levelMax, false},           // one past the end
		{Level(255), false},         // far out of range
	}

	for _, tc := range cases {
		t.Run(tc.level.String(), func(t *testing.T) {
			if got := tc.level.IsValid(); got != tc.expected {
				t.Errorf("Level(%d).IsValid() = %v, expected %v", int(tc.level), got, tc.expected)
			}
		})
	}
}

func Test_Level_OrderIsAscending(t *testing.T) {
	// severity must increase from Transient to Error
	// this is relied upon by threshold comparisons throughout the codebase
	levels := []Level{Transient, Verbose, Info, Warning, Error}

	for i := 1; i < len(levels); i++ {
		if levels[i] <= levels[i-1] {
			t.Errorf("expected %v (%d) to be greater than %v (%d)",
				levels[i], int(levels[i]),
				levels[i-1], int(levels[i-1]),
			)
		}
	}
}

func Test_Level_MinAndMaxBounds(t *testing.T) {
	if levelMin != 0 {
		t.Errorf("expected levelMin to be 0, got %d", int(levelMin))
	}
	if levelMax <= levelMin {
		t.Errorf("expected levelMax (%d) to be greater than levelMin (%d)",
			int(levelMax), int(levelMin))
	}
}

func Test_Level_AllLevelsAreBetweenMinAndMax(t *testing.T) {
	levels := []Level{Transient, Verbose, Info, Warning, Error}

	for _, l := range levels {
		if l < levelMin || l >= levelMax {
			t.Errorf("Level %v (%d) is outside [levelMin, levelMax) range [%d, %d)",
				l, int(l), int(levelMin), int(levelMax))
		}
	}
}