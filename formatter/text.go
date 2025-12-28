package formatter

import (
	"fmt"
	"strings"

	"github.com/JakeTRogers/subnetCalc/logger"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

// TextFormatter formats network information as plain styled text.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// FormatNetwork formats a single network's information as styled text.
func (f *TextFormatter) FormatNetwork(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Msg("formatting network as text")
	info := ToNetworkInfo(n)
	return FormatNetworkSummary(info), nil
}

// FormatSubnets formats a network with its subnets as styled text (simple list).
func (f *TextFormatter) FormatSubnets(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Int("subnet_count", len(n.Subnets)).Msg("formatting subnets as text")
	if len(n.Subnets) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nSubnets (%d total):\n", len(n.Subnets)))

	for i, sn := range n.Subnets {
		b.WriteString(fmt.Sprintf("  %d. %s (hosts: %s)\n", i+1, sn.CIDR.String(), FormatMaxHosts(sn.MaxHosts)))
	}

	return b.String(), nil
}

// Ensure TextFormatter implements Formatter.
var _ Formatter = (*TextFormatter)(nil)
