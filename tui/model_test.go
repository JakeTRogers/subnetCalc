package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel_variations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cidr       string
		targetBits int
		wantErr    bool
		wantRows   int
	}{
		{
			name:       "Valid /24 no split",
			cidr:       "192.168.1.0/24",
			targetBits: 0,
			wantErr:    false,
			wantRows:   1,
		},
		{
			name:       "Valid /24 split to /26",
			cidr:       "192.168.1.0/24",
			targetBits: 26,
			wantErr:    false,
			wantRows:   4,
		},
		{
			name:       "Invalid CIDR",
			cidr:       "invalid",
			targetBits: 0,
			wantErr:    true,
		},
		{
			name:       "Invalid target bits (too small)",
			cidr:       "192.168.1.0/24",
			targetBits: 20,
			wantErr:    true,
		},
		{
			name:       "Invalid target bits (too large)",
			cidr:       "192.168.1.0/24",
			targetBits: 31,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := NewModel(tt.cidr, tt.targetBits)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(model.rows) != tt.wantRows {
				t.Errorf("NewModel() rows = %d, want %d", len(model.rows), tt.wantRows)
			}
		})
	}
}

func TestModel_updateRows(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Initially 1 row
	if len(model.rows) != 1 {
		t.Errorf("Initial rows = %d, want 1", len(model.rows))
	}

	// Split the root
	model.root.Split()
	model.updateRows()

	// Now 2 rows
	if len(model.rows) != 2 {
		t.Errorf("After split rows = %d, want 2", len(model.rows))
	}

	// Join back
	model.root.Join()
	model.updateRows()

	// Back to 1 row
	if len(model.rows) != 1 {
		t.Errorf("After join rows = %d, want 1", len(model.rows))
	}
}

func TestModel_getMaxBits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Initially /24
	if model.getMaxBits() != 24 {
		t.Errorf("Initial maxBits = %d, want 24", model.getMaxBits())
	}

	// Split to create /25s
	model.root.Split()
	model.updateRows()

	if model.getMaxBits() != 25 {
		t.Errorf("After split maxBits = %d, want 25", model.getMaxBits())
	}
}

func TestModel_hasSplits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	if model.hasSplits() {
		t.Error("Expected hasSplits() = false initially")
	}

	model.root.Split()
	model.updateRows()

	if !model.hasSplits() {
		t.Error("Expected hasSplits() = true after split")
	}
}

func TestModel_exportJSON(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.1.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	json := model.exportJSON()

	expectedFields := []string{
		`"cidr"`,
		`"192.168.1.0/24"`,
		`"firstIP"`,
		`"lastIP"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(json, field) {
			t.Errorf("JSON missing expected content: %s", field)
		}
	}
}

func TestModel_cursorBounds(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Should have 4 rows
	if len(model.rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(model.rows))
	}

	// Initial cursor should be 0
	if model.cursor != 0 {
		t.Errorf("Initial cursor = %d, want 0", model.cursor)
	}

	// Set cursor beyond bounds
	model.cursor = 100
	model.updateRows()

	// Should be clamped to max valid index
	if model.cursor != 3 {
		t.Errorf("Clamped cursor = %d, want 3", model.cursor)
	}

	// Set negative
	model.cursor = -10
	model.updateRows()

	if model.cursor != 0 {
		t.Errorf("Clamped negative cursor = %d, want 0", model.cursor)
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_View_loading(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Width 0 should show loading
	model.width = 0
	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View with width=0 should show Loading, got %q", view)
	}
}

func TestModel_View_normal(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	model.height = 40
	view := model.View()

	if !strings.Contains(view, "192.168.0.0/24") {
		t.Errorf("View should contain CIDR, got %q", view)
	}
	if !strings.Contains(view, "Subnet Calculator") {
		t.Errorf("View should contain title, got %q", view)
	}
}

func TestModel_View_withStatus(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	model.height = 40
	model.statusMsg = "Test status message"
	view := model.View()

	if !strings.Contains(view, "Test status message") {
		t.Errorf("View should contain status message, got %q", view)
	}
}

func TestModel_calculateColumnWidths_IPv4(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	widths := model.calculateColumnWidths()

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
}

func TestModel_calculateColumnWidths_IPv6(t *testing.T) {
	t.Parallel()
	model, err := NewModel("2001:db8::/64", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 160
	widths := model.calculateColumnWidths()

	// IPv6 should have wider columns
	if widths.subnet < 15 {
		t.Errorf("IPv6 subnet width should be at least 15, got %d", widths.subnet)
	}
}

func TestModel_calculateColumnWidths_narrowTerminal(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 60 // Very narrow
	widths := model.calculateColumnWidths()

	// Should still return valid widths (minimums)
	minWidths := minColumnWidths()
	if widths.subnet < minWidths.subnet {
		t.Errorf("narrow terminal subnet width = %d, should be at least %d", widths.subnet, minWidths.subnet)
	}
}

func TestModel_calculateColumnWidths_wideSplits(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 28)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 200
	widths := model.calculateColumnWidths()

	// With many splits, should have reasonable widths
	if widths.splitCol < 5 {
		t.Errorf("splitCol width should be at least 5, got %d", widths.splitCol)
	}
}

func TestModel_Update_windowResize(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Send window resize message
	msg := tea.WindowSizeMsg{Width: 150, Height: 50}
	newModel, _ := model.Update(msg)
	updated := newModel.(Model)

	if updated.width != 150 {
		t.Errorf("width after resize = %d, want 150", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height after resize = %d, want 50", updated.height)
	}
}

func TestModel_Update_clearStatus(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.statusMsg = "Test message"

	// Send clear status message
	newModel, _ := model.Update(clearStatusMsg{})
	updated := newModel.(Model)

	if updated.statusMsg != "" {
		t.Errorf("statusMsg after clear = %q, want empty", updated.statusMsg)
	}
}

func TestModel_handleKeyPress_quit(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Test quit key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.handleKeyPress(msg)

	// Should return tea.Quit command
	if cmd == nil {
		t.Error("quit key should return a command")
	}
}

func TestModel_handleKeyPress_navigation(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Test down navigation
	model.cursor = 0
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.handleKeyPress(downMsg)
	updated := newModel.(Model)
	if updated.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", updated.cursor)
	}

	// Test up navigation
	upMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = updated.handleKeyPress(upMsg)
	updated = newModel.(Model)
	if updated.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", updated.cursor)
	}

	// Test up at top (should stay at 0)
	newModel, _ = updated.handleKeyPress(upMsg)
	updated = newModel.(Model)
	if updated.cursor != 0 {
		t.Errorf("cursor at top after up = %d, want 0", updated.cursor)
	}
}

func TestModel_handleKeyPress_downAtBottom(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Move to bottom
	model.cursor = len(model.rows) - 1

	// Try to go down
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.handleKeyPress(downMsg)
	updated := newModel.(Model)

	// Should stay at bottom
	if updated.cursor != len(model.rows)-1 {
		t.Errorf("cursor at bottom after down = %d, want %d", updated.cursor, len(model.rows)-1)
	}
}

func TestModel_handleKeyPress_split(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	model.height = 40

	// Initially 1 row
	if len(model.rows) != 1 {
		t.Fatalf("initial rows = %d, want 1", len(model.rows))
	}

	// Split the subnet
	splitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, _ := model.handleKeyPress(splitMsg)
	updated := newModel.(Model)

	// Should now have 2 rows
	if len(updated.rows) != 2 {
		t.Errorf("rows after split = %d, want 2", len(updated.rows))
	}
}

func TestModel_handleKeyPress_join(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 26)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.width = 120
	model.height = 40

	// Move to first child
	model.cursor = 0

	// Join should collapse to parent
	joinMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	newModel, _ := model.handleKeyPress(joinMsg)
	updated := newModel.(Model)

	// After join, should have fewer rows
	if len(updated.rows) >= 4 {
		t.Errorf("rows after join = %d, should be less than 4", len(updated.rows))
	}
}

func TestModel_handleKeyPress_export(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	// Export key
	exportMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, cmd := model.handleKeyPress(exportMsg)
	updated := newModel.(Model)

	if updated.statusMsg == "" {
		t.Error("export should set status message")
	}
	if cmd == nil {
		t.Error("export should return clearStatusAfter command")
	}
}

func TestModel_handleKeyPress_help(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 0)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	initialShowAll := model.help.ShowAll

	// Toggle help
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.handleKeyPress(helpMsg)
	updated := newModel.(Model)

	if updated.help.ShowAll == initialShowAll {
		t.Error("help toggle should change ShowAll state")
	}
}

func TestModel_handleKeyPress_horizontalScroll(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 28)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.scrollOffset = 1

	// Scroll left
	leftMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ := model.handleKeyPress(leftMsg)
	updated := newModel.(Model)
	if updated.scrollOffset != 0 {
		t.Errorf("scrollOffset after left = %d, want 0", updated.scrollOffset)
	}

	// Try to scroll left at 0 (should stay at 0)
	newModel, _ = updated.handleKeyPress(leftMsg)
	updated = newModel.(Model)
	if updated.scrollOffset != 0 {
		t.Errorf("scrollOffset at 0 after left = %d, want 0", updated.scrollOffset)
	}

	// Scroll right
	rightMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ = updated.handleKeyPress(rightMsg)
	updated = newModel.(Model)
	if updated.scrollOffset != 1 {
		t.Errorf("scrollOffset after right = %d, want 1", updated.scrollOffset)
	}
}

func TestModel_handleKeyPress_pageUpDown(t *testing.T) {
	t.Parallel()
	model, err := NewModel("192.168.0.0/24", 28)
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}

	model.height = 20 // Set height for viewport calculation
	initialCursor := model.cursor

	// Page down
	pageDownMsg := tea.KeyMsg{Type: tea.KeyPgDown}
	newModel, _ := model.handleKeyPress(pageDownMsg)
	updated := newModel.(Model)
	if updated.cursor <= initialCursor {
		t.Errorf("cursor after page down = %d, should be > %d", updated.cursor, initialCursor)
	}

	// Page up
	pageUpMsg := tea.KeyMsg{Type: tea.KeyPgUp}
	newModel, _ = updated.handleKeyPress(pageUpMsg)
	updated = newModel.(Model)
	// Should move cursor up
	if updated.cursor >= len(model.rows)-1 && len(model.rows) > 10 {
		t.Errorf("cursor after page up = %d, should have moved up", updated.cursor)
	}
}
