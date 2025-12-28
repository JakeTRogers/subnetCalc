// Package formatter provides output formatting interfaces and implementations.
package formatter

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JakeTRogers/subnetCalc/logger"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

// OutputFormat represents the available output format types.
type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatTable OutputFormat = "table"
	FormatText  OutputFormat = "text"

	// DefaultTerminalWidth is the default width used for table formatting
	// when no terminal width is detected.
	DefaultTerminalWidth = 120
)

// Config holds configuration options for formatter creation.
type Config struct {
	Format      OutputFormat
	Width       int  // Terminal width for table formatting
	PrettyPrint bool // Pretty print JSON output
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Format:      FormatTable,
		Width:       DefaultTerminalWidth,
		PrettyPrint: true,
	}
}

// New creates a Formatter based on the provided configuration.
// This is the preferred way to create formatters.
func New(cfg Config) Formatter {
	log := logger.GetLogger()
	log.Trace().Str("format", string(cfg.Format)).Int("width", cfg.Width).Bool("pretty_print", cfg.PrettyPrint).Msg("creating formatter")

	switch cfg.Format {
	case FormatJSON:
		return NewJSONFormatter(cfg.PrettyPrint)
	case FormatText:
		return NewTextFormatter()
	case FormatTable:
		fallthrough
	default:
		return NewTableFormatter(cfg.Width)
	}
}

// Formatter defines the interface for formatting network information.
type Formatter interface {
	// FormatNetwork formats a single network's information.
	FormatNetwork(n subnet.Network) (string, error)

	// FormatSubnets formats a network with its subnets.
	FormatSubnets(n subnet.Network) (string, error)
}

// NetworkInfo holds formatted network information for display.
type NetworkInfo struct {
	CIDR          string
	NetworkAddr   string
	BroadcastAddr string
	FirstHostIP   string
	LastHostIP    string
	SubnetMask    string
	MaskBits      int
	MaxHosts      string
}

// SubnetInfo holds formatted subnet information for table display.
type SubnetInfo struct {
	CIDR       string
	SubnetMask string
	FirstIP    string
	LastIP     string
	Broadcast  string
	Hosts      string
}

// ToNetworkInfo converts a subnet.Network to formatted NetworkInfo for display.
func ToNetworkInfo(n subnet.Network) NetworkInfo {
	return NetworkInfo{
		CIDR:          n.CIDR.String(),
		NetworkAddr:   n.NetworkAddr.String(),
		BroadcastAddr: n.BroadcastAddr.String(),
		FirstHostIP:   n.FirstHostIP.String(),
		LastHostIP:    n.LastHostIP.String(),
		SubnetMask:    n.SubnetMask.String(),
		MaskBits:      n.MaskBits,
		MaxHosts:      FormatMaxHosts(n.MaxHosts),
	}
}

// ToSubnetInfo converts a subnet.Network to formatted SubnetInfo for table display.
func ToSubnetInfo(n subnet.Network) SubnetInfo {
	return SubnetInfo{
		CIDR:       n.CIDR.String(),
		SubnetMask: n.SubnetMask.String(),
		FirstIP:    n.FirstHostIP.String(),
		LastIP:     n.LastHostIP.String(),
		Broadcast:  n.BroadcastAddr.String(),
		Hosts:      FormatMaxHosts(n.MaxHosts),
	}
}

// ToSubnetInfoSlice converts a slice of subnet.Network to SubnetInfo.
func ToSubnetInfoSlice(networks []subnet.Network) []SubnetInfo {
	result := make([]SubnetInfo, len(networks))
	for i, n := range networks {
		result[i] = ToSubnetInfo(n)
	}
	return result
}

// FormatMaxHosts returns a human-readable string for max hosts.
// Caps display at a readable threshold for very large IPv6 networks.
func FormatMaxHosts(maxHosts *big.Int) string {
	if maxHosts == nil {
		return "0"
	}

	// Threshold: 2^64 (display as ">2^64" for anything larger)
	threshold := new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)

	if maxHosts.Cmp(threshold) >= 0 {
		// Find approximate power of 2
		bitLen := maxHosts.BitLen()
		return fmt.Sprintf(">2^%d", bitLen-1)
	}

	// For smaller numbers, format with commas
	return formatBigIntWithCommas(maxHosts)
}

// formatBigIntWithCommas formats a big.Int with thousand separators.
func formatBigIntWithCommas(n *big.Int) string {
	s := n.String()
	if len(s) <= 3 {
		return s
	}

	// Add commas from right to left
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)

}

// FormatNumber formats a non-negative integer with comma thousand separators (e.g., 1234 becomes "1,234").
func FormatNumber(n uint) string {
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

// Shared styles for consistent formatting across formatters.
var (
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
)

// FormatNetworkSummary formats a NetworkInfo as styled text.
// This is shared between TableFormatter and TextFormatter.
func FormatNetworkSummary(info NetworkInfo) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("               Network:") + " " + valueStyle.Render(info.CIDR) + "\n")
	b.WriteString(labelStyle.Render("    Host Address Range:") + " " + valueStyle.Render(info.FirstHostIP+" - "+info.LastHostIP) + "\n")
	b.WriteString(labelStyle.Render("     Broadcast Address:") + " " + valueStyle.Render(info.BroadcastAddr) + "\n")
	b.WriteString(labelStyle.Render("           Subnet Mask:") + " " + valueStyle.Render(info.SubnetMask) + "\n")
	b.WriteString(labelStyle.Render("         Maximum Hosts:") + " " + valueStyle.Render(info.MaxHosts) + "\n")
	return b.String()
}
