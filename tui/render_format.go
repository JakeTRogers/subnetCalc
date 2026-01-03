package tui

import (
	"fmt"
	"strings"
)

// rangeParts holds the separated first and last IP strings for aligned rendering.
type rangeParts struct {
	first string
	last  string
}

// formatRangeParts returns the abbreviated first and last IP strings separately.
// This allows for aligned rendering where the "-" separator is at a consistent column.
func formatRangeParts(firstIP, lastIP, networkAddr string) rangeParts {
	firstBytes := parseIPBytes(firstIP)
	lastBytes := parseIPBytes(lastIP)
	netBytes := parseIPBytes(networkAddr)

	if firstBytes == nil || lastBytes == nil || netBytes == nil {
		return rangeParts{first: firstIP, last: lastIP}
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

	return rangeParts{first: firstStr, last: lastStr}
}

// formatRangeAbbreviated formats the IP range, showing only differing octets.
func formatRangeAbbreviated(firstIP, lastIP, networkAddr string) string {
	parts := formatRangeParts(firstIP, lastIP, networkAddr)
	return fmt.Sprintf("%s - %s", parts.first, parts.last)
}

// parseIPBytes parses an IP string into bytes (IPv4 only for abbreviation).
func parseIPBytes(ip string) []byte {
	var bytes [4]byte
	// Error safely ignored: we validate via return count n instead.
	n, _ := fmt.Sscanf(ip, "%d.%d.%d.%d", &bytes[0], &bytes[1], &bytes[2], &bytes[3])
	if n != 4 {
		return nil
	}
	return bytes[:]
}
