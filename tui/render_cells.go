package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JakeTRogers/subnetCalc/internal/ui"
)

// spanInfo tracks merged cell spans for hierarchy visualization.
type spanInfo struct {
	node     *SubnetNode
	startRow int
	endRow   int
	rowCount int
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
			if node.CIDR().Bits() < bits {
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

// renderSplitCell renders a single split hierarchy cell.
func (m Model) renderSplitCell(node *SubnetNode, rowIdx, bits, splitColWidth int, columnSpans map[int][]spanInfo, isSelected bool) string {
	color := ui.GetColorForPrefix(bits, m.initialPrefix)
	spans := columnSpans[bits]

	// Check if this row's subnet is large enough to show in this column
	if node.CIDR().Bits() < bits {
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
