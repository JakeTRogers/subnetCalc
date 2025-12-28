// Package subnet provides core domain types and calculations for IP subnet operations.
package subnet

import (
	"fmt"
	"math/big"
	"net/netip"

	"github.com/JakeTRogers/subnetCalc/logger"
)

// MaxGeneratedSubnets is the safety limit for subnet splitting operations.
// This prevents accidental memory exhaustion from very large splits.
const MaxGeneratedSubnets = 1_000_000

// Network represents an IP network with calculated properties.
type Network struct {
	CIDR          netip.Prefix `json:"cidr"`
	NetworkAddr   netip.Addr   `json:"networkAddr"`
	BroadcastAddr netip.Addr   `json:"broadcastAddr"`
	FirstHostIP   netip.Addr   `json:"firstIP"`
	LastHostIP    netip.Addr   `json:"lastIP"`
	SubnetMask    netip.Addr   `json:"subnetMask"`
	MaskBits      int          `json:"maskBits"`
	MaxHosts      *big.Int     `json:"maxHosts"`
	Subnets       []Network    `json:"subnets,omitempty"`
}

// NewNetwork creates a Network from a CIDR string.
func NewNetwork(cidr string) (Network, error) {
	log := logger.GetLogger()
	log.Debug().Str("cidr", cidr).Msg("creating network")

	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return Network{}, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	n := NewNetworkFromPrefix(prefix)
	log.Debug().Str("cidr", n.CIDR.String()).Int("mask_bits", n.MaskBits).Msg("network created successfully")
	return n, nil
}

// NewNetworkFromPrefix creates a Network from a netip.Prefix.
func NewNetworkFromPrefix(prefix netip.Prefix) Network {
	// Normalize to network address
	prefix = netip.PrefixFrom(prefix.Masked().Addr(), prefix.Bits())

	n := Network{
		CIDR:        prefix,
		NetworkAddr: prefix.Addr(),
		MaskBits:    prefix.Bits(),
	}

	// Calculate subnet mask
	n.SubnetMask = CalculateSubnetMask(prefix.Bits(), prefix.Addr().BitLen())

	// Calculate broadcast address
	n.BroadcastAddr = CalculateBroadcastAddr(n.NetworkAddr, n.SubnetMask)

	// Calculate first and last usable host IPs
	n.FirstHostIP = n.NetworkAddr.Next()
	n.LastHostIP = n.BroadcastAddr.Prev()

	// Calculate max hosts using big.Int for IPv6 support
	n.MaxHosts = CalculateMaxHosts(prefix.Addr().BitLen(), prefix.Bits())

	return n
}

// Split divides this network into subnets of the specified prefix length.
func (n *Network) Split(targetBits int) error {
	log := logger.GetLogger()
	log.Debug().Str("cidr", n.CIDR.String()).Int("current_bits", n.MaskBits).Int("target_bits", targetBits).Msg("splitting network")

	if targetBits <= n.MaskBits {
		return fmt.Errorf("target prefix /%d must be larger than network prefix /%d", targetBits, n.MaskBits)
	}

	maxBits := n.CIDR.Addr().BitLen()
	if targetBits > maxBits {
		return fmt.Errorf("target prefix /%d exceeds maximum /%d for this address family", targetBits, maxBits)
	}

	// Guard against generating an enormous number of subnets (and against int overflow).
	diff := targetBits - n.MaskBits
	numSubnets := new(big.Int).Lsh(big.NewInt(1), uint(diff))
	if !numSubnets.IsInt64() {
		return fmt.Errorf("requested split would generate too many subnets")
	}
	if numSubnets.Int64() > MaxGeneratedSubnets {
		return fmt.Errorf("requested split would generate %d subnets (limit %d)", numSubnets.Int64(), int64(MaxGeneratedSubnets))
	}

	log.Trace().Int64("subnet_count", numSubnets.Int64()).Msg("generating subnets")
	n.Subnets = GenerateSubnets(n.CIDR, targetBits)
	log.Debug().Int("subnet_count", len(n.Subnets)).Msg("network split completed")
	return nil
}

// GenerateSubnets creates all subnets of targetBits size within the given prefix.
func GenerateSubnets(prefix netip.Prefix, targetBits int) []Network {
	diff := targetBits - prefix.Bits()
	// Calculate number of subnets: 2^(targetBits - currentBits)
	numSubnets := 1 << diff
	subnets := make([]Network, 0, numSubnets)

	currentAddr := prefix.Addr()
	for i := 0; i < numSubnets; i++ {
		subnetPrefix := netip.PrefixFrom(currentAddr, targetBits)
		subnet := NewNetworkFromPrefix(subnetPrefix)
		subnets = append(subnets, subnet)

		// Move to next subnet's network address
		currentAddr = AddToAddr(subnet.BroadcastAddr, 1)
	}

	return subnets
}

// CalculateSubnetMask generates a subnet mask for the given mask bits and address size.
func CalculateSubnetMask(maskBits, addrBits int) netip.Addr {
	maskBytes := make([]byte, addrBits/8)
	remainingBits := maskBits

	for i := range maskBytes {
		if remainingBits >= 8 {
			maskBytes[i] = 0xFF
			remainingBits -= 8
		} else if remainingBits > 0 {
			maskBytes[i] = byte(0xFF << (8 - remainingBits))
			remainingBits = 0
		}
	}

	// Error is safe to ignore: maskBytes length is addrBits/8, which is always
	// 4 (IPv4) or 16 (IPv6) bytes when derived from a valid netip.Addr.
	addr, _ := netip.AddrFromSlice(maskBytes)
	return addr
}

// CalculateBroadcastAddr computes the broadcast address from network address and subnet mask.
func CalculateBroadcastAddr(networkAddr, subnetMask netip.Addr) netip.Addr {
	netBytes := networkAddr.AsSlice()
	maskBytes := subnetMask.AsSlice()
	broadcastBytes := make([]byte, len(netBytes))

	for i := range netBytes {
		broadcastBytes[i] = netBytes[i] | ^maskBytes[i]
	}

	// Error is safe to ignore: broadcastBytes has same length as netBytes,
	// which is always 4 (IPv4) or 16 (IPv6) bytes from a valid netip.Addr.
	addr, _ := netip.AddrFromSlice(broadcastBytes)
	return addr
}

// CalculateMaxHosts returns the number of usable host addresses in a subnet.
// Uses big.Int to handle IPv6 networks without overflow.
func CalculateMaxHosts(addrBits, maskBits int) *big.Int {
	hostBits := addrBits - maskBits
	if hostBits <= 0 {
		return big.NewInt(0)
	}

	// 2^hostBits - 2 (subtract network and broadcast addresses)
	maxHosts := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(hostBits)), nil)
	maxHosts.Sub(maxHosts, big.NewInt(2))

	// Don't return negative for /31 or /32 networks
	if maxHosts.Sign() < 0 {
		return big.NewInt(0)
	}

	return maxHosts
}

// AddToAddr adds n to an IP address, returning the resulting address.
func AddToAddr(addr netip.Addr, n int) netip.Addr {
	bytes := addr.AsSlice()

	// Add n to the address bytes (big-endian)
	carry := n
	for i := len(bytes) - 1; i >= 0 && carry > 0; i-- {
		sum := int(bytes[i]) + carry
		bytes[i] = byte(sum & 0xFF)
		carry = sum >> 8
	}

	// Error is safe to ignore: bytes slice comes from addr.AsSlice(),
	// which is always 4 (IPv4) or 16 (IPv6) bytes from a valid netip.Addr.
	result, _ := netip.AddrFromSlice(bytes)
	return result
}
