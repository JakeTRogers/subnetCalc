package tui

import (
	"strings"
	"testing"
)

// contains is a helper for tests to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestNewModel(t *testing.T) {
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

func TestModelUpdateRows(t *testing.T) {
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

func TestModelGetMaxBits(t *testing.T) {
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

func TestModelHasSplits(t *testing.T) {
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

func TestModelExportJSON(t *testing.T) {
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
		if !contains(json, field) {
			t.Errorf("JSON missing expected content: %s", field)
		}
	}
}

func TestModelCursorBounds(t *testing.T) {
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
