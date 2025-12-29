package tui

import (
	"encoding/json"
	"net/netip"

	"github.com/JakeTRogers/subnetCalc/subnet"
)

// SubnetNode represents a node in the subnet tree hierarchy.
// Each node embeds network information and can be split into two child subnets
// or joined back with its sibling into the parent subnet.
type SubnetNode struct {
	Network  subnet.Network // Network holds the subnet calculation details
	Parent   *SubnetNode    // Parent node, nil for root
	Children []*SubnetNode  // Child subnets when split
	IsSplit  bool           // Whether this node has been split
}

// CIDR returns the network prefix (e.g., 192.168.1.0/24).
func (n *SubnetNode) CIDR() netip.Prefix {
	return n.Network.CIDR
}

// FirstIP returns the first usable host IP address.
func (n *SubnetNode) FirstIP() netip.Addr {
	return n.Network.FirstHostIP
}

// LastIP returns the last usable host IP address.
func (n *SubnetNode) LastIP() netip.Addr {
	return n.Network.LastHostIP
}

// BroadcastAddr returns the broadcast address (last IP in range for IPv6).
func (n *SubnetNode) BroadcastAddr() netip.Addr {
	return n.Network.BroadcastAddr
}

// SubnetMask returns the subnet mask as an address.
func (n *SubnetNode) SubnetMask() netip.Addr {
	return n.Network.SubnetMask
}

// Hosts returns the number of usable host addresses (capped at max uint for large IPv6 networks).
func (n *SubnetNode) Hosts() uint {
	// For display in TUI, cap at max uint
	if !n.Network.MaxHosts.IsUint64() {
		return ^uint(0) // Max uint value
	}
	maxHosts := n.Network.MaxHosts.Uint64()
	if maxHosts > uint64(^uint(0)) {
		return ^uint(0)
	}
	return uint(maxHosts)
}

// ExportNode is a JSON-serializable representation of a subnet node.
// It mirrors SubnetNode but uses string representations suitable for JSON output.
type ExportNode struct {
	CIDR          string        `json:"cidr"`               // Network in CIDR notation
	FirstIP       string        `json:"firstIP"`            // First usable host IP
	LastIP        string        `json:"lastIP"`             // Last usable host IP
	BroadcastAddr string        `json:"broadcastAddr"`      // Broadcast address
	SubnetMask    string        `json:"subnetMask"`         // Subnet mask
	Hosts         uint          `json:"hosts"`              // Number of usable hosts
	Children      []*ExportNode `json:"children,omitempty"` // Child subnets if split
}

// createSubnetNode creates a new subnet node from a prefix.
func createSubnetNode(prefix netip.Prefix, parent *SubnetNode) *SubnetNode {
	return &SubnetNode{
		Network: subnet.NewNetworkFromPrefix(prefix),
		Parent:  parent,
	}
}

// Split divides a subnet node into two child subnets by incrementing the prefix length.
// Returns true if the split was successful, false if already split or at max depth.
func (n *SubnetNode) Split() bool {
	cidr := n.Network.CIDR
	if n.IsSplit || cidr.Bits() >= MaxSplitDepth {
		return false
	}

	newBits := cidr.Bits() + 1
	networkAddr := cidr.Masked().Addr()

	// First child: same network address, smaller prefix
	child1Prefix := netip.PrefixFrom(networkAddr, newBits)
	n.Children = append(n.Children, createSubnetNode(child1Prefix, n))

	// Second child: next network address
	child1Broadcast := n.Children[0].Network.BroadcastAddr
	child2Addr := child1Broadcast.Next()
	child2Prefix := netip.PrefixFrom(child2Addr, newBits)
	n.Children = append(n.Children, createSubnetNode(child2Prefix, n))

	n.IsSplit = true
	return true
}

// SplitToDepth recursively splits the subnet until all leaves reach the target prefix length.
// For example, SplitToDepth(26) on a /24 creates four /26 subnets.
func (n *SubnetNode) SplitToDepth(targetBits int) {
	if n.Network.CIDR.Bits() >= targetBits {
		return
	}

	// Split this node
	if !n.Split() {
		return
	}

	// Recursively split children
	for _, child := range n.Children {
		child.SplitToDepth(targetBits)
	}
}

// Join merges all children back into this node, undoing any splits.
// Returns true if joined, false if the node was not split.
func (n *SubnetNode) Join() bool {
	if !n.IsSplit {
		return false
	}

	// Recursively join children first
	for _, child := range n.Children {
		child.Join()
	}

	n.Children = nil
	n.IsSplit = false
	return true
}

// toExportNode converts SubnetNode tree to ExportNode tree.
func (n *SubnetNode) toExportNode() *ExportNode {
	export := &ExportNode{
		CIDR:          n.Network.CIDR.String(),
		FirstIP:       n.Network.FirstHostIP.String(),
		LastIP:        n.Network.LastHostIP.String(),
		BroadcastAddr: n.Network.BroadcastAddr.String(),
		SubnetMask:    n.Network.SubnetMask.String(),
		Hosts:         n.Hosts(),
	}

	for _, child := range n.Children {
		export.Children = append(export.Children, child.toExportNode())
	}

	return export
}

// ExportJSON returns the subnet tree as a formatted JSON string.
// The output includes the full tree structure with all split children.
func (n *SubnetNode) ExportJSON() (string, error) {
	export := n.toExportNode()
	jsonBytes, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// getAncestorAtDepth returns the ancestor at a specific prefix length.
func getAncestorAtDepth(node *SubnetNode, targetBits int) *SubnetNode {
	if node == nil {
		return nil
	}
	if node.Network.CIDR.Bits() == targetBits {
		return node
	}
	return getAncestorAtDepth(node.Parent, targetBits)
}

// collectLeaves recursively collects all leaf nodes into the provided slice.
func collectLeaves(node *SubnetNode, leaves *[]*SubnetNode) {
	if node == nil {
		return
	}

	if !node.IsSplit {
		*leaves = append(*leaves, node)
	} else {
		for _, child := range node.Children {
			collectLeaves(child, leaves)
		}
	}
}
