package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JakeTRogers/subnetCalc/formatter"
	"github.com/JakeTRogers/subnetCalc/internal/ui"
)

// columnWidths holds calculated column widths for the table.
type columnWidths struct {
	subnet   int
	mask     int
	rangeCol int
	hosts    int
	splitCol int
}

// minColumnWidths returns minimum widths for readability.
func minColumnWidths() columnWidths {
	return columnWidths{
		subnet:   12, // "/xx" + some room
		mask:     12, // "255.255.x.x"
		rangeCol: 15, // abbreviated range
		hosts:    7,  // "Hosts" (5 chars) + padding for lipgloss rendering
		splitCol: 5,  // "/xx"
	}
}

// renderTable renders the subnet table with merged cells for hierarchy.
func (m Model) renderTable() string {
	if len(m.rows) == 0 {
		return "No subnets to display"
	}

	maxBits := m.getMaxBits()
	hasSplits := m.hasSplits()

	// Calculate dynamic column widths based on content
	widths := m.calculateColumnWidths()
	subnetWidth := widths.subnet
	maskWidth := widths.mask
	rangeWidth := widths.rangeCol
	hostsWidth := widths.hosts
	splitColWidth := widths.splitCol

	// Calculate available width for split columns
	mainWidth := subnetWidth + maskWidth + rangeWidth + hostsWidth + 8
	availableWidth := m.width - mainWidth - 4

	// Number of split columns: from initial prefix to deepest split
	numSplitLevels := 0
	if hasSplits {
		numSplitLevels = maxBits - m.initialPrefix + 1
	}

	maxVisibleSplitCols := max(1, min(availableWidth/splitColWidth, numSplitLevels))

	// Adjust scroll offset bounds
	scrollOffset := m.scrollOffset
	maxScroll := max(0, numSplitLevels-maxVisibleSplitCols)
	scrollOffset = max(0, min(scrollOffset, maxScroll))

	// Build header
	header := m.buildHeader(subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits, hasSplits, maxVisibleSplitCols, scrollOffset)

	// Pre-calculate span information
	columnSpans := m.calculateColumnSpans(numSplitLevels, maxBits, hasSplits)

	// Calculate viewport and scroll (using local copy for this render pass)
	viewportHeight := max(3, m.height-10)
	verticalScroll := m.calculateVerticalScroll(viewportHeight)
	maxVerticalScroll := max(0, len(m.rows)-viewportHeight)

	// Build rows
	rowStrings := m.buildRows(verticalScroll, viewportHeight, subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits, hasSplits, maxVisibleSplitCols, scrollOffset, columnSpans)

	// Scroll indicator
	scrollIndicator := m.buildScrollIndicator(verticalScroll, hasSplits, numSplitLevels, maxVisibleSplitCols, scrollOffset, maxScroll, viewportHeight, maxVerticalScroll)

	// Combine header and rows
	allRows := append([]string{header}, rowStrings...)
	table := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	result := ui.BorderStyle.Render(table)
	if scrollIndicator != "" {
		result += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(scrollIndicator)
	}

	return result
}

// buildHeader constructs the table header row.
func (m Model) buildHeader(subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits int, hasSplits bool, maxVisibleSplitCols, scrollOffset int) string {
	var headerParts []string

	subnetLabel := "Subnet"
	maskLabel := "Subnet Mask"
	rangeLabel := "Assignable Range"
	hostsLabel := "Hosts"

	// Abbreviate headers if columns are too narrow
	if maskWidth < 12 {
		maskLabel = "Mask"
	}
	if rangeWidth < 17 {
		rangeLabel = "Range"
	}

	headerParts = append(headerParts, ui.HeaderStyle.Width(subnetWidth).Render(subnetLabel))
	headerParts = append(headerParts, ui.HeaderStyle.Width(maskWidth).Render(maskLabel))
	headerParts = append(headerParts, ui.HeaderStyle.Width(rangeWidth).Render(rangeLabel))
	headerParts = append(headerParts, ui.HeaderStyle.Width(hostsWidth).Render(hostsLabel))

	// Add split column headers
	if hasSplits {
		for i := 0; i < maxVisibleSplitCols; i++ {
			colIdx := i + scrollOffset
			bits := maxBits - colIdx
			if bits < m.initialPrefix {
				bits = m.initialPrefix
			}
			color := ui.GetColorForPrefix(bits, m.initialPrefix)
			colHeader := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(color).
				Width(splitColWidth).
				Align(lipgloss.Center).
				Render(fmt.Sprintf("/%d", bits))
			headerParts = append(headerParts, colHeader)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)
}

// calculateVerticalScroll computes the vertical scroll offset to keep cursor visible.
// It returns the adjusted scroll offset based on cursor position and viewport height.
// This function computes the scroll offset based on the model state without mutating it.
func (m Model) calculateVerticalScroll(viewportHeight int) int {
	scroll := m.verticalScroll

	if m.cursor < scroll {
		scroll = m.cursor
	} else if m.cursor >= scroll+viewportHeight {
		scroll = m.cursor - viewportHeight + 1
	}

	maxVerticalScroll := max(0, len(m.rows)-viewportHeight)
	return max(0, min(scroll, maxVerticalScroll))
}

// buildRows constructs the visible row strings.
func (m Model) buildRows(verticalScroll, viewportHeight, subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits int, hasSplits bool, maxVisibleSplitCols, scrollOffset int, columnSpans map[int][]spanInfo) []string {
	var rowStrings []string

	for rowIdx := verticalScroll; rowIdx < len(m.rows) && rowIdx < verticalScroll+viewportHeight; rowIdx++ {
		node := m.rows[rowIdx]
		isSelected := rowIdx == m.cursor

		style := ui.NormalStyle
		if isSelected {
			style = ui.SelectedStyle
		}

		// Format the main columns
		subnet := style.Width(subnetWidth).Render(node.CIDR().String())
		mask := style.Width(maskWidth).Render(node.SubnetMask().String())

		networkAddr := node.CIDR().Masked().Addr()
		rangeStr := formatRangeAbbreviated(node.FirstIP().String(), node.LastIP().String(), networkAddr.String())
		rangeCell := style.Width(rangeWidth).Render(rangeStr)

		hosts := style.Width(hostsWidth).Render(formatter.FormatNumber(node.Hosts()))

		var rowParts []string
		rowParts = append(rowParts, subnet, mask, rangeCell, hosts)

		// Add split hierarchy columns
		if hasSplits {
			for i := 0; i < maxVisibleSplitCols; i++ {
				colIdx := i + scrollOffset
				bits := maxBits - colIdx
				if bits < m.initialPrefix {
					bits = m.initialPrefix
				}

				cellContent := m.renderSplitCell(node, rowIdx, bits, splitColWidth, columnSpans, isSelected)
				rowParts = append(rowParts, cellContent)
			}
		}

		rowStrings = append(rowStrings, lipgloss.JoinHorizontal(lipgloss.Top, rowParts...))
	}

	return rowStrings
}

// buildScrollIndicator creates the scroll indicator text.
func (m Model) buildScrollIndicator(verticalScroll int, hasSplits bool, numSplitLevels, maxVisibleSplitCols, scrollOffset, maxScroll, viewportHeight, maxVerticalScroll int) string {
	var scrollIndicators []string

	if hasSplits && numSplitLevels > maxVisibleSplitCols {
		scrollIndicators = append(scrollIndicators, fmt.Sprintf("h-scroll: %d/%d", scrollOffset+1, maxScroll+1))
	}
	if len(m.rows) > viewportHeight {
		scrollIndicators = append(scrollIndicators, fmt.Sprintf("v-scroll: %d/%d", verticalScroll+1, maxVerticalScroll+1))
	}

	if len(scrollIndicators) > 0 {
		return fmt.Sprintf(" [%s, use ↑↓/PgUp/PgDn to navigate]", strings.Join(scrollIndicators, ", "))
	}
	return ""
}
