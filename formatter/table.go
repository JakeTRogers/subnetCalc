package formatter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JakeTRogers/subnetCalc/internal/ui"
	"github.com/JakeTRogers/subnetCalc/logger"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

// Table column width constants for consistent formatting.
const (
	colIndexWidth     = 4  // Width for row number column
	colSubnetWidth    = 20 // Width for subnet CIDR column
	colMaskWidth      = 16 // Width for subnet mask column
	colRangeWidth     = 30 // Width for assignable IP range column
	colBroadcastWidth = 16 // Width for broadcast address column
	colHostsWidth     = 12 // Width for host count column
)

// TableFormatter formats network information as styled tables.
type TableFormatter struct {
	terminalWidth int
}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter(terminalWidth int) *TableFormatter {
	return &TableFormatter{terminalWidth: terminalWidth}
}

// FormatNetwork formats a single network's information as a styled table.
func (f *TableFormatter) FormatNetwork(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Msg("formatting network as table")
	info := ToNetworkInfo(n)
	return FormatNetworkSummary(info), nil
}

// FormatSubnets formats a network with its subnets as a styled table.
func (f *TableFormatter) FormatSubnets(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Int("subnet_count", len(n.Subnets)).Msg("formatting subnets as table")
	if len(n.Subnets) == 0 {
		return "", nil
	}

	subnets := ToSubnetInfoSlice(n.Subnets)
	return f.renderTable(n.CIDR.String(), subnets), nil
}

// renderTable renders a subnet table.
func (f *TableFormatter) renderTable(parentCIDR string, subnets []SubnetInfo) string {
	if len(subnets) == 0 {
		return "No subnets to display"
	}

	// Build header
	var headerParts []string
	headerParts = append(headerParts, ui.HeaderStyle.Width(colIndexWidth).Render("#"))
	headerParts = append(headerParts, ui.HeaderStyle.Width(colSubnetWidth).Render("Subnet"))
	headerParts = append(headerParts, ui.HeaderStyle.Width(colMaskWidth).Render("Subnet Mask"))
	headerParts = append(headerParts, ui.HeaderStyle.Width(colRangeWidth).Render("Assignable Range"))
	headerParts = append(headerParts, ui.HeaderStyle.Width(colBroadcastWidth).Render("Broadcast"))
	headerParts = append(headerParts, ui.HeaderStyle.Width(colHostsWidth).Render("Hosts"))

	header := lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)

	// Build rows
	var rowStrings []string

	for i, sn := range subnets {
		// Alternate row colors for readability
		var style lipgloss.Style
		if i%2 == 0 {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
		}

		num := style.Width(colIndexWidth).Render(fmt.Sprintf("%d", i+1))
		cidr := style.Width(colSubnetWidth).Render(sn.CIDR)
		mask := style.Width(colMaskWidth).Render(sn.SubnetMask)
		rangeStr := fmt.Sprintf("%s - %s", sn.FirstIP, sn.LastIP)
		rangeCell := style.Width(colRangeWidth).Render(rangeStr)
		broadcastCell := style.Width(colBroadcastWidth).Render(sn.Broadcast)
		hosts := style.Width(colHostsWidth).Render(sn.Hosts)

		var rowParts []string
		rowParts = append(rowParts, num, cidr, mask, rangeCell, broadcastCell, hosts)
		rowStrings = append(rowStrings, lipgloss.JoinHorizontal(lipgloss.Top, rowParts...))
	}

	// Title
	title := ui.TitleStyle.Render(fmt.Sprintf("  %s contains %d /%s subnets:",
		parentCIDR, len(subnets), extractPrefix(subnets[0].CIDR)))

	// Combine header and rows
	allRows := append([]string{header}, rowStrings...)
	table := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	return title + "\n" + ui.BorderStyle.Render(table)
}

// extractPrefix extracts the prefix number from a CIDR string.
func extractPrefix(cidr string) string {
	parts := strings.Split(cidr, "/")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// Ensure TableFormatter implements Formatter.
var _ Formatter = (*TableFormatter)(nil)
