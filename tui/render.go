package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for table rendering.
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("24")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
)

// prefixColors is a color palette for different prefix lengths.
var prefixColors = []lipgloss.Color{
	lipgloss.Color("212"), // Pink
	lipgloss.Color("141"), // Purple
	lipgloss.Color("75"),  // Light blue
	lipgloss.Color("81"),  // Cyan
	lipgloss.Color("120"), // Light green
	lipgloss.Color("228"), // Yellow
	lipgloss.Color("216"), // Orange
	lipgloss.Color("210"), // Salmon
	lipgloss.Color("177"), // Light purple
	lipgloss.Color("69"),  // Blue
	lipgloss.Color("87"),  // Light cyan
	lipgloss.Color("156"), // Pale green
	lipgloss.Color("222"), // Gold
	lipgloss.Color("213"), // Light pink
	lipgloss.Color("105"), // Blue-purple
	lipgloss.Color("192"), // Yellow-green
}

// getColorForPrefix returns a color based on the prefix length.
func getColorForPrefix(bits, initialPrefix int) lipgloss.Color {
	idx := (bits - initialPrefix) % len(prefixColors)
	if idx < 0 {
		idx = 0
	}
	return prefixColors[idx]
}

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
		hosts:    6,  // "Hosts"
		splitCol: 5,  // "/xx"
	}
}

// spanInfo tracks merged cell spans for hierarchy visualization.
type spanInfo struct {
	node     *SubnetNode
	startRow int
	endRow   int
	rowCount int
}

// formatRangeAbbreviated formats the IP range, showing only differing octets.
func formatRangeAbbreviated(firstIP, lastIP, networkAddr string) string {
	firstBytes := parseIPBytes(firstIP)
	lastBytes := parseIPBytes(lastIP)
	netBytes := parseIPBytes(networkAddr)

	if firstBytes == nil || lastBytes == nil || netBytes == nil {
		return fmt.Sprintf("%s - %s", firstIP, lastIP)
	}

	// Find first differing octet between lastIP and network address
	firstDiffOctet := 3 // Default to last octet
	for i := 0; i < 4; i++ {
		if lastBytes[i] != netBytes[i] {
			firstDiffOctet = i
			break
		}
	}

	// Format first IP - show from first differing octet onward
	var firstParts []string
	for i := firstDiffOctet; i < 4; i++ {
		firstParts = append(firstParts, fmt.Sprintf("%d", firstBytes[i]))
	}
	firstStr := "." + strings.Join(firstParts, ".")

	// Format last IP - show from first differing octet onward
	var lastParts []string
	for i := firstDiffOctet; i < 4; i++ {
		lastParts = append(lastParts, fmt.Sprintf("%d", lastBytes[i]))
	}
	lastStr := "." + strings.Join(lastParts, ".")

	return fmt.Sprintf("%s - %s", firstStr, lastStr)
}

// parseIPBytes parses an IP string into bytes (IPv4 only for abbreviation).
func parseIPBytes(ip string) []byte {
	var bytes [4]byte
	n, _ := fmt.Sscanf(ip, "%d.%d.%d.%d", &bytes[0], &bytes[1], &bytes[2], &bytes[3])
	if n != 4 {
		return nil
	}
	return bytes[:]
}

// formatNumber formats a number with comma separators.
func formatNumber(n uint) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
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

	// Calculate viewport and scroll
	viewportHeight := max(3, m.height-10)
	m.adjustScrollForCursor(viewportHeight)
	maxVerticalScroll := max(0, len(m.rows)-viewportHeight)

	// Build rows
	rowStrings := m.buildRows(viewportHeight, subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits, hasSplits, maxVisibleSplitCols, scrollOffset, columnSpans)

	// Scroll indicator
	scrollIndicator := m.buildScrollIndicator(hasSplits, numSplitLevels, maxVisibleSplitCols, scrollOffset, maxScroll, viewportHeight, maxVerticalScroll)

	// Combine header and rows
	allRows := append([]string{header}, rowStrings...)
	table := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	result := borderStyle.Render(table)
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

	headerParts = append(headerParts, headerStyle.Width(subnetWidth).Render(subnetLabel))
	headerParts = append(headerParts, headerStyle.Width(maskWidth).Render(maskLabel))
	headerParts = append(headerParts, headerStyle.Width(rangeWidth).Render(rangeLabel))
	headerParts = append(headerParts, headerStyle.Width(hostsWidth).Render(hostsLabel))

	// Add split column headers
	if hasSplits {
		for i := 0; i < maxVisibleSplitCols; i++ {
			colIdx := i + scrollOffset
			bits := maxBits - colIdx
			if bits < m.initialPrefix {
				bits = m.initialPrefix
			}
			color := getColorForPrefix(bits, m.initialPrefix)
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

// calculateColumnSpans pre-calculates span information for each column.
func (m Model) calculateColumnSpans(numSplitLevels, maxBits int, hasSplits bool) map[int][]spanInfo {
	columnSpans := make(map[int][]spanInfo)

	if !hasSplits {
		return columnSpans
	}

	for colIdx := 0; colIdx < numSplitLevels; colIdx++ {
		bits := maxBits - colIdx
		if bits < m.initialPrefix {
			bits = m.initialPrefix
		}

		var spans []spanInfo
		var currentSpan *spanInfo

		for rowIdx, node := range m.rows {
			if node.CIDR.Bits() < bits {
				if currentSpan != nil {
					currentSpan.endRow = rowIdx - 1
					currentSpan.rowCount = currentSpan.endRow - currentSpan.startRow + 1
					spans = append(spans, *currentSpan)
					currentSpan = nil
				}
				continue
			}

			ancestor := getAncestorAtDepth(node, bits)
			if ancestor == nil {
				ancestor = m.root
			}

			if currentSpan == nil || currentSpan.node != ancestor {
				if currentSpan != nil {
					currentSpan.endRow = rowIdx - 1
					currentSpan.rowCount = currentSpan.endRow - currentSpan.startRow + 1
					spans = append(spans, *currentSpan)
				}
				currentSpan = &spanInfo{
					node:     ancestor,
					startRow: rowIdx,
				}
			}
		}

		if currentSpan != nil {
			currentSpan.endRow = len(m.rows) - 1
			currentSpan.rowCount = currentSpan.endRow - currentSpan.startRow + 1
			spans = append(spans, *currentSpan)
		}

		columnSpans[bits] = spans
	}

	return columnSpans
}

// adjustScrollForCursor ensures the cursor is within the viewport.
func (m *Model) adjustScrollForCursor(viewportHeight int) {
	if m.cursor < m.verticalScroll {
		m.verticalScroll = m.cursor
	} else if m.cursor >= m.verticalScroll+viewportHeight {
		m.verticalScroll = m.cursor - viewportHeight + 1
	}

	maxVerticalScroll := max(0, len(m.rows)-viewportHeight)
	m.verticalScroll = max(0, min(m.verticalScroll, maxVerticalScroll))
}

// buildRows constructs the visible row strings.
func (m Model) buildRows(viewportHeight, subnetWidth, maskWidth, rangeWidth, hostsWidth, splitColWidth, maxBits int, hasSplits bool, maxVisibleSplitCols, scrollOffset int, columnSpans map[int][]spanInfo) []string {
	var rowStrings []string

	for rowIdx := m.verticalScroll; rowIdx < len(m.rows) && rowIdx < m.verticalScroll+viewportHeight; rowIdx++ {
		node := m.rows[rowIdx]
		isSelected := rowIdx == m.cursor

		style := normalStyle
		if isSelected {
			style = selectedStyle
		}

		// Format the main columns
		subnet := style.Width(subnetWidth).Render(node.CIDR.String())
		mask := style.Width(maskWidth).Render(node.SubnetMask.String())

		networkAddr := node.CIDR.Masked().Addr()
		rangeStr := formatRangeAbbreviated(node.FirstIP.String(), node.LastIP.String(), networkAddr.String())
		rangeCell := style.Width(rangeWidth).Render(rangeStr)

		hosts := style.Width(hostsWidth).Render(formatNumber(node.Hosts))

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

// renderSplitCell renders a single split hierarchy cell.
func (m Model) renderSplitCell(node *SubnetNode, rowIdx, bits, splitColWidth int, columnSpans map[int][]spanInfo, isSelected bool) string {
	color := getColorForPrefix(bits, m.initialPrefix)
	spans := columnSpans[bits]

	// Check if this row's subnet is large enough to show in this column
	if node.CIDR.Bits() < bits {
		return lipgloss.NewStyle().Width(splitColWidth).Render("")
	}

	// Find which span this row belongs to
	var currentSpanInfo *spanInfo
	for idx := range spans {
		if rowIdx >= spans[idx].startRow && rowIdx <= spans[idx].endRow {
			currentSpanInfo = &spans[idx]
			break
		}
	}

	if currentSpanInfo == nil {
		return lipgloss.NewStyle().Width(splitColWidth).Render("")
	}

	cellContent := buildSpanCell(currentSpanInfo, rowIdx, bits, splitColWidth)

	finalStyle := lipgloss.NewStyle().Foreground(color).Width(splitColWidth)
	if isSelected {
		finalStyle = finalStyle.Background(lipgloss.Color("57"))
	}
	return finalStyle.Render(cellContent)
}

// buildSpanCell creates the visual content for a span cell.
func buildSpanCell(spanInfo *spanInfo, rowIdx, bits, splitColWidth int) string {
	posInSpan := rowIdx - spanInfo.startRow
	spanSize := spanInfo.rowCount
	isFirst := posInSpan == 0
	isLast := posInSpan == spanSize-1
	midPoint := (spanSize - 1) / 2
	isMid := posInSpan == midPoint

	label := fmt.Sprintf("/%d", bits)
	innerWidth := splitColWidth - 2

	// Pad label to center it within inner width
	paddedLabel := fmt.Sprintf("%*s", (innerWidth+len(label))/2, label)
	paddedLabel = fmt.Sprintf("%-*s", innerWidth, paddedLabel)

	leftBorder := "│"
	rightBorder := "│"

	switch {
	case spanSize == 1:
		return leftBorder + paddedLabel + rightBorder
	case spanSize == 2:
		if isFirst {
			return leftBorder + paddedLabel + rightBorder
		}
		return "└" + strings.Repeat("─", innerWidth) + "┘"
	case spanSize == 3:
		if isFirst {
			return "┌" + strings.Repeat("─", innerWidth) + "┐"
		} else if isLast {
			return "└" + strings.Repeat("─", innerWidth) + "┘"
		}
		return leftBorder + paddedLabel + rightBorder
	case isFirst:
		return "┌" + strings.Repeat("─", innerWidth) + "┐"
	case isLast:
		return "└" + strings.Repeat("─", innerWidth) + "┘"
	case isMid:
		return leftBorder + paddedLabel + rightBorder
	default:
		return leftBorder + strings.Repeat(" ", innerWidth) + rightBorder
	}
}

// buildScrollIndicator creates the scroll indicator text.
func (m Model) buildScrollIndicator(hasSplits bool, numSplitLevels, maxVisibleSplitCols, scrollOffset, maxScroll, viewportHeight, maxVerticalScroll int) string {
	var scrollIndicators []string

	if hasSplits && numSplitLevels > maxVisibleSplitCols {
		scrollIndicators = append(scrollIndicators, fmt.Sprintf("h-scroll: %d/%d", scrollOffset+1, maxScroll+1))
	}
	if len(m.rows) > viewportHeight {
		scrollIndicators = append(scrollIndicators, fmt.Sprintf("v-scroll: %d/%d", m.verticalScroll+1, maxVerticalScroll+1))
	}

	if len(scrollIndicators) > 0 {
		return fmt.Sprintf(" [%s, use ↑↓/PgUp/PgDn to navigate]", strings.Join(scrollIndicators, ", "))
	}
	return ""
}
