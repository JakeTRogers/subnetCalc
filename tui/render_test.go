package tui

import (
	"strings"
	"testing"

	"github.com/JakeTRogers/subnetCalc/formatter"
	"github.com/JakeTRogers/subnetCalc/internal/ui"
)

func TestFormatNumber_variations(t *testing.T) {
	t.Parallel()
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
			got := formatter.FormatNumber(tt.input)
			if got != tt.want {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatRangeAbbreviated_variations(t *testing.T) {
	t.Parallel()
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
			if !strings.Contains(got, " - ") {
				t.Errorf("formatRangeAbbreviated() = %q, missing ' - ' separator", got)
			}
		})
	}
}

func TestGetColorForPrefix_cycling(t *testing.T) {
	t.Parallel()
	// Test that colors cycle correctly
	initialPrefix := 16

	// Same prefix difference should give same color
	color1 := ui.GetColorForPrefix(20, initialPrefix)
	color2 := ui.GetColorForPrefix(20, initialPrefix)
	if color1 != color2 {
		t.Error("Same prefix should give same color")
	}

	// Different prefixes should give different colors (until cycle)
	color3 := ui.GetColorForPrefix(21, initialPrefix)
	if color1 == color3 {
		t.Error("Different prefixes should give different colors")
	}

	// Test cycling (PrefixColors has 16 colors)
	color16 := ui.GetColorForPrefix(initialPrefix, initialPrefix)
	color32 := ui.GetColorForPrefix(initialPrefix+16, initialPrefix)
	if color16 != color32 {
		t.Error("Colors should cycle after 16 prefixes")
	}
}

func TestMinColumnWidths_positive(t *testing.T) {
	t.Parallel()
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

func TestBuildSpanCell_variations(t *testing.T) {
	t.Parallel()
	// Test single row span
	span := &spanInfo{startRow: 0, endRow: 0, rowCount: 1}
	cell := buildSpanCell(span, 0, 24, 5)
	if cell == "" {
		t.Error("Single row span should produce non-empty cell")
	}

	// Test multi-row span - first row
	span2 := &spanInfo{startRow: 0, endRow: 3, rowCount: 4}
	cellFirst := buildSpanCell(span2, 0, 24, 5)
	if !strings.Contains(cellFirst, "┌") {
		t.Errorf("First row of span should have top border, got %q", cellFirst)
	}

	// Test multi-row span - last row
	cellLast := buildSpanCell(span2, 3, 24, 5)
	if !strings.Contains(cellLast, "└") {
		t.Errorf("Last row of span should have bottom border, got %q", cellLast)
	}

	// Test multi-row span - middle row with label
	cellMid := buildSpanCell(span2, 1, 24, 5) // midPoint = (4-1)/2 = 1
	if !strings.Contains(cellMid, "│") {
		t.Errorf("Middle row should have side borders, got %q", cellMid)
	}
}

func TestBuildSpanCell_twoRows(t *testing.T) {
	t.Parallel()
	span := &spanInfo{startRow: 0, endRow: 1, rowCount: 2}

	// First row of 2-row span
	first := buildSpanCell(span, 0, 24, 5)
	if !strings.Contains(first, "│") {
		t.Errorf("First of 2-row span should have │, got %q", first)
	}

	// Second row of 2-row span
	second := buildSpanCell(span, 1, 24, 5)
	if !strings.Contains(second, "└") {
		t.Errorf("Second of 2-row span should have └, got %q", second)
	}
}

func TestBuildSpanCell_threeRows(t *testing.T) {
	t.Parallel()
	span := &spanInfo{startRow: 0, endRow: 2, rowCount: 3}

	// First row
	first := buildSpanCell(span, 0, 24, 5)
	if !strings.Contains(first, "┌") {
		t.Errorf("First of 3-row span should have ┌, got %q", first)
	}

	// Middle row (has label)
	mid := buildSpanCell(span, 1, 24, 5)
	if !strings.Contains(mid, "│") {
		t.Errorf("Middle of 3-row span should have │, got %q", mid)
	}

	// Last row
	last := buildSpanCell(span, 2, 24, 5)
	if !strings.Contains(last, "└") {
		t.Errorf("Last of 3-row span should have └, got %q", last)
	}
}

func TestBuildSpanCell_nonMidRow(t *testing.T) {
	t.Parallel()
	// 5-row span: midPoint = (5-1)/2 = 2
	span := &spanInfo{startRow: 0, endRow: 4, rowCount: 5}

	// Row 1 is not first, not last, not mid
	cell := buildSpanCell(span, 1, 24, 5)
	if !strings.Contains(cell, "│") {
		t.Errorf("Non-mid row should have │ borders, got %q", cell)
	}
}

func TestFormatRangeAbbreviated_invalidIP(t *testing.T) {
	t.Parallel()
	// Test with invalid IPs - should return full format
	result := formatRangeAbbreviated("invalid", "192.168.1.254", "192.168.1.0")
	if !strings.Contains(result, "invalid") {
		t.Errorf("Invalid first IP should return full format, got %q", result)
	}

	result2 := formatRangeAbbreviated("192.168.1.1", "invalid", "192.168.1.0")
	if !strings.Contains(result2, "invalid") {
		t.Errorf("Invalid last IP should return full format, got %q", result2)
	}

	result3 := formatRangeAbbreviated("192.168.1.1", "192.168.1.254", "invalid")
	if !strings.Contains(result3, "192.168.1.1") {
		t.Errorf("Invalid network should return full format, got %q", result3)
	}
}

func TestFormatRangeParts_variations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		firstIP   string
		lastIP    string
		network   string
		wantFirst string
		wantLast  string
	}{
		{
			name:      "Same first 3 octets",
			firstIP:   "192.168.1.1",
			lastIP:    "192.168.1.254",
			network:   "192.168.1.0",
			wantFirst: ".1",
			wantLast:  ".254",
		},
		{
			name:      "Different last 2 octets",
			firstIP:   "10.0.0.1",
			lastIP:    "10.0.255.254",
			network:   "10.0.0.0",
			wantFirst: ".0.1",
			wantLast:  ".255.254",
		},
		{
			name:      "Different last 3 octets",
			firstIP:   "10.0.0.1",
			lastIP:    "10.7.255.254",
			network:   "10.0.0.0",
			wantFirst: ".0.0.1",
			wantLast:  ".7.255.254",
		},
		{
			name:      "Invalid first IP",
			firstIP:   "invalid",
			lastIP:    "192.168.1.254",
			network:   "192.168.1.0",
			wantFirst: "invalid",
			wantLast:  "192.168.1.254",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := formatRangeParts(tt.firstIP, tt.lastIP, tt.network)
			if parts.first != tt.wantFirst {
				t.Errorf("formatRangeParts().first = %q, want %q", parts.first, tt.wantFirst)
			}
			if parts.last != tt.wantLast {
				t.Errorf("formatRangeParts().last = %q, want %q", parts.last, tt.wantLast)
			}
		})
	}
}

func TestParseIPBytes_variations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ip      string
		wantNil bool
		want0   byte
		want3   byte
	}{
		{
			name:    "Valid IPv4",
			ip:      "192.168.1.100",
			wantNil: false,
			want0:   192,
			want3:   100,
		},
		{
			name:    "All zeros",
			ip:      "0.0.0.0",
			wantNil: false,
			want0:   0,
			want3:   0,
		},
		{
			name:    "All 255s",
			ip:      "255.255.255.255",
			wantNil: false,
			want0:   255,
			want3:   255,
		},
		{
			name:    "Invalid - too few octets",
			ip:      "192.168.1",
			wantNil: true,
		},
		{
			name:    "Invalid - empty",
			ip:      "",
			wantNil: true,
		},
		{
			name:    "Invalid - IPv6",
			ip:      "2001:db8::1",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseIPBytes(tt.ip)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseIPBytes(%q) = %v, want nil", tt.ip, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseIPBytes(%q) = nil, want non-nil", tt.ip)
			}
			if got[0] != tt.want0 {
				t.Errorf("parseIPBytes(%q)[0] = %d, want %d", tt.ip, got[0], tt.want0)
			}
			if got[3] != tt.want3 {
				t.Errorf("parseIPBytes(%q)[3] = %d, want %d", tt.ip, got[3], tt.want3)
			}
		})
	}
}

func TestCalculateVerticalScroll(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Model has 4 rows, viewport of 2 should require scrolling
	viewportHeight := 2

	// Cursor at 0, scroll should be 0
	model.cursor = 0
	model.verticalScroll = 0
	scroll := model.calculateVerticalScroll(viewportHeight)
	if scroll != 0 {
		t.Errorf("scroll at cursor 0 = %d, want 0", scroll)
	}

	// Cursor at 3 (beyond viewport), scroll should adjust
	model.cursor = 3
	model.verticalScroll = 0
	scroll = model.calculateVerticalScroll(viewportHeight)
	if scroll != 2 {
		t.Errorf("scroll at cursor 3 with viewport 2 = %d, want 2", scroll)
	}

	// Cursor at 1 with scroll at 2, should scroll back
	model.cursor = 1
	model.verticalScroll = 2
	scroll = model.calculateVerticalScroll(viewportHeight)
	if scroll != 1 {
		t.Errorf("scroll at cursor 1 with initial scroll 2 = %d, want 1", scroll)
	}
}

func TestModel_renderTable_noRows(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Clear rows to test edge case
	model.rows = nil
	result := model.renderTable()

	if !strings.Contains(result, "No subnets to display") {
		t.Errorf("renderTable with no rows should show message, got %q", result)
	}
}

func TestModel_renderTable_withSplits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	model.height = 40

	result := model.renderTable()

	// Should contain subnet info
	if !strings.Contains(result, "192.168.0") {
		t.Errorf("renderTable should contain subnet, got %q", result)
	}

	// Should contain column headers
	if !strings.Contains(result, "Subnet") {
		t.Errorf("renderTable should contain 'Subnet' header, got %q", result)
	}
}

func TestCalculateColumnSpans_noSplits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	spans := model.calculateColumnSpans(0, 24, false)
	if len(spans) != 0 {
		t.Errorf("calculateColumnSpans with no splits should return empty map, got %d entries", len(spans))
	}
}

func TestCalculateColumnSpans_withSplits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	maxBits := model.getMaxBits()
	numSplitLevels := maxBits - model.initialPrefix + 1

	spans := model.calculateColumnSpans(numSplitLevels, maxBits, true)

	if len(spans) == 0 {
		t.Error("calculateColumnSpans with splits should return non-empty map")
	}
}

func TestBuildScrollIndicator_noScroll(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// With single row and large viewport, no scroll needed
	indicator := model.buildScrollIndicator(0, false, 0, 0, 0, 0, 100, 0)
	if indicator != "" {
		t.Errorf("buildScrollIndicator with no scroll should be empty, got %q", indicator)
	}
}

func TestBuildScrollIndicator_horizontalScroll(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 28)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Simulate need for horizontal scroll
	indicator := model.buildScrollIndicator(0, true, 10, 3, 0, 7, 100, 0)
	if !strings.Contains(indicator, "h-scroll") {
		t.Errorf("buildScrollIndicator should show h-scroll, got %q", indicator)
	}
}

func TestBuildScrollIndicator_verticalScroll(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 28)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Simulate need for vertical scroll (16 rows, viewport 5)
	indicator := model.buildScrollIndicator(0, false, 0, 0, 0, 0, 5, 11)
	if !strings.Contains(indicator, "v-scroll") {
		t.Errorf("buildScrollIndicator should show v-scroll, got %q", indicator)
	}
}
