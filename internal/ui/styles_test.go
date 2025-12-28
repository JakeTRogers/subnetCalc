package ui

import (
	"testing"
)

func TestGetColorForPrefix_variations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		bits          int
		initialPrefix int
		wantIdx       int // expected color index in PrefixColors
	}{
		{
			name:          "Same as initial prefix",
			bits:          24,
			initialPrefix: 24,
			wantIdx:       0,
		},
		{
			name:          "One more than initial",
			bits:          25,
			initialPrefix: 24,
			wantIdx:       1,
		},
		{
			name:          "Five more than initial",
			bits:          29,
			initialPrefix: 24,
			wantIdx:       5,
		},
		{
			name:          "Cycles after 16",
			bits:          40,
			initialPrefix: 24,
			wantIdx:       0, // (40-24) % 16 = 0
		},
		{
			name:          "Cycles partial",
			bits:          42,
			initialPrefix: 24,
			wantIdx:       2, // (42-24) % 16 = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetColorForPrefix(tt.bits, tt.initialPrefix)
			want := PrefixColors[tt.wantIdx]
			if got != want {
				t.Errorf("GetColorForPrefix(%d, %d) = %v, want %v", tt.bits, tt.initialPrefix, got, want)
			}
		})
	}
}

func TestGetColorForPrefix_negativeIndex(t *testing.T) {
	t.Parallel()
	// When bits < initialPrefix, idx would be negative, but we clamp to 0
	got := GetColorForPrefix(20, 24)
	want := PrefixColors[0]
	if got != want {
		t.Errorf("GetColorForPrefix(20, 24) with negative index = %v, want %v (index 0)", got, want)
	}
}

func TestGetColorForPrefix_sameColor(t *testing.T) {
	t.Parallel()
	// Same input should always give same output
	color1 := GetColorForPrefix(26, 24)
	color2 := GetColorForPrefix(26, 24)
	if color1 != color2 {
		t.Error("Same inputs should produce same color")
	}
}

func TestGetColorForPrefix_differentColors(t *testing.T) {
	t.Parallel()
	// Different prefix depths should give different colors (until cycle)
	color1 := GetColorForPrefix(24, 24)
	color2 := GetColorForPrefix(25, 24)
	if color1 == color2 {
		t.Error("Different prefix depths should give different colors")
	}
}

func TestPrefixColors_length(t *testing.T) {
	t.Parallel()
	if len(PrefixColors) != 16 {
		t.Errorf("PrefixColors should have 16 colors, got %d", len(PrefixColors))
	}
}

func TestStyles_notNil(t *testing.T) {
	t.Parallel()
	// Verify all exported styles are initialized
	tests := []struct {
		name  string
		style interface{}
	}{
		{"HeaderStyle", HeaderStyle},
		{"SelectedStyle", SelectedStyle},
		{"NormalStyle", NormalStyle},
		{"BorderStyle", BorderStyle},
		{"TitleStyle", TitleStyle},
		{"StatusStyle", StatusStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.style == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
		})
	}
}
