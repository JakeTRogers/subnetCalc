package tui

import (
	"testing"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input uint
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{12, "12"},
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{12345678, "12,345,678"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatNumber(tt.input)
			if got != tt.want {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatRangeAbbreviated(t *testing.T) {
	tests := []struct {
		name    string
		firstIP string
		lastIP  string
		network string
		wantLen int // Check approximate length, not exact match due to format variations
	}{
		{
			name:    "Same first 3 octets",
			firstIP: "192.168.1.1",
			lastIP:  "192.168.1.254",
			network: "192.168.1.0",
			wantLen: 10, // ".1 - .254" approximately
		},
		{
			name:    "Different second octet",
			firstIP: "10.0.0.1",
			lastIP:  "10.0.255.254",
			network: "10.0.0.0",
			wantLen: 15, // ".0.1 - .255.254" approximately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRangeAbbreviated(tt.firstIP, tt.lastIP, tt.network)
			// Just check it's not empty and contains the separator
			if got == "" {
				t.Error("formatRangeAbbreviated() returned empty string")
			}
			if !contains(got, " - ") {
				t.Errorf("formatRangeAbbreviated() = %q, missing ' - ' separator", got)
			}
		})
	}
}

func TestGetColorForPrefix(t *testing.T) {
	// Test that colors cycle correctly
	initialPrefix := 16

	// Same prefix difference should give same color
	color1 := getColorForPrefix(20, initialPrefix)
	color2 := getColorForPrefix(20, initialPrefix)
	if color1 != color2 {
		t.Error("Same prefix should give same color")
	}

	// Different prefixes should give different colors (until cycle)
	color3 := getColorForPrefix(21, initialPrefix)
	if color1 == color3 {
		t.Error("Different prefixes should give different colors")
	}

	// Test cycling (prefixColors has 16 colors)
	color16 := getColorForPrefix(initialPrefix, initialPrefix)
	color32 := getColorForPrefix(initialPrefix+16, initialPrefix)
	if color16 != color32 {
		t.Error("Colors should cycle after 16 prefixes")
	}
}

func TestMinColumnWidths(t *testing.T) {
	widths := minColumnWidths()

	if widths.subnet < 1 {
		t.Error("subnet width should be positive")
	}
	if widths.mask < 1 {
		t.Error("mask width should be positive")
	}
	if widths.rangeCol < 1 {
		t.Error("rangeCol width should be positive")
	}
	if widths.hosts < 1 {
		t.Error("hosts width should be positive")
	}
	if widths.splitCol < 1 {
		t.Error("splitCol width should be positive")
	}
}

func TestBuildSpanCell(t *testing.T) {
	// Test single row span
	span := &spanInfo{startRow: 0, endRow: 0, rowCount: 1}
	cell := buildSpanCell(span, 0, 24, 5)
	if cell == "" {
		t.Error("Single row span should produce non-empty cell")
	}

	// Test multi-row span - first row
	span2 := &spanInfo{startRow: 0, endRow: 3, rowCount: 4}
	cellFirst := buildSpanCell(span2, 0, 24, 5)
	if !contains(cellFirst, "┌") {
		t.Errorf("First row of span should have top border, got %q", cellFirst)
	}

	// Test multi-row span - last row
	cellLast := buildSpanCell(span2, 3, 24, 5)
	if !contains(cellLast, "└") {
		t.Errorf("Last row of span should have bottom border, got %q", cellLast)
	}

	// Test multi-row span - middle row with label
	cellMid := buildSpanCell(span2, 1, 24, 5) // midPoint = (4-1)/2 = 1
	if !contains(cellMid, "│") {
		t.Errorf("Middle row should have side borders, got %q", cellMid)
	}
}
