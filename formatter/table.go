package formatter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JakeTRogers/subnetCalc/subnet"
)

// TableFormatter formats network information as styled tables.
type TableFormatter struct {
	terminalWidth int
}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter(terminalWidth int) *TableFormatter {
	return &TableFormatter{terminalWidth: terminalWidth}
}

// Styles for table rendering
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("24")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)
)

// FormatNetwork formats a single network's information as a styled table.
func (f *TableFormatter) FormatNetwork(n subnet.Network) (string, error) {
	info := ToNetworkInfo(n)
	return FormatNetworkSummary(info), nil
}

// FormatSubnets formats a network with its subnets as a styled table.
func (f *TableFormatter) FormatSubnets(n subnet.Network) (string, error) {
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

	// Column widths
	numWidth := 4
	subnetWidth := 20
	maskWidth := 16
	rangeWidth := 30
	broadcastWidth := 16
	hostsWidth := 12

	// Build header
	var headerParts []string
	headerParts = append(headerParts, headerStyle.Width(numWidth).Render("#"))
	headerParts = append(headerParts, headerStyle.Width(subnetWidth).Render("Subnet"))
	headerParts = append(headerParts, headerStyle.Width(maskWidth).Render("Subnet Mask"))
	headerParts = append(headerParts, headerStyle.Width(rangeWidth).Render("Assignable Range"))
	headerParts = append(headerParts, headerStyle.Width(broadcastWidth).Render("Broadcast"))
	headerParts = append(headerParts, headerStyle.Width(hostsWidth).Render("Hosts"))

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

		num := style.Width(numWidth).Render(fmt.Sprintf("%d", i+1))
		cidr := style.Width(subnetWidth).Render(sn.CIDR)
		mask := style.Width(maskWidth).Render(sn.SubnetMask)
		rangeStr := fmt.Sprintf("%s - %s", sn.FirstIP, sn.LastIP)
		rangeCell := style.Width(rangeWidth).Render(rangeStr)
		broadcastCell := style.Width(broadcastWidth).Render(sn.Broadcast)
		hosts := style.Width(hostsWidth).Render(sn.Hosts)

		var rowParts []string
		rowParts = append(rowParts, num, cidr, mask, rangeCell, broadcastCell, hosts)
		rowStrings = append(rowStrings, lipgloss.JoinHorizontal(lipgloss.Top, rowParts...))
	}

	// Title
	title := titleStyle.Render(fmt.Sprintf("  %s contains %d /%s subnets:",
		parentCIDR, len(subnets), extractPrefix(subnets[0].CIDR)))

	// Combine header and rows
	allRows := append([]string{header}, rowStrings...)
	table := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	return title + "\n" + borderStyle.Render(table)
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
