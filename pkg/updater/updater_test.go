package updater

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{"newer major", "2.0.0", "1.0.0", true},
		{"newer minor", "1.1.0", "1.0.0", true},
		{"newer patch", "1.0.1", "1.0.0", true},
		{"same version", "1.0.0", "1.0.0", false},
		{"older major", "1.0.0", "2.0.0", false},
		{"older minor", "1.0.0", "1.1.0", false},
		{"older patch", "1.0.0", "1.0.1", false},
		{"two digit patch newer", "1.0.10", "1.0.5", true},
		{"two digit minor newer", "1.10.0", "1.5.0", true},
		{"complex newer", "2.1.3", "1.9.9", true},
		{"short version", "1.1", "1.0.5", true},
		{"very short version", "2", "1.9.9", true},
		{"pre-release tag", "1.0.1-beta", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNewerVersion(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
					tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}

func TestParseVersionPart(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"10", 10},
		{"123", 123},
		{"5-beta", 5},
		{"3-rc1", 3},
		{"", 0},
		{"abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseVersionPart(tt.input)
			if result != tt.expected {
				t.Errorf("parseVersionPart(%q) = %d, want %d",
					tt.input, result, tt.expected)
			}
		})
	}
}
