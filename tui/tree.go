package tui

import (
	"encoding/json"
	"net/netip"

	"github.com/JakeTRogers/subnetCalc/subnet"
)

// SubnetNode represents a node in the subnet tree.
type SubnetNode struct {
	CIDR          netip.Prefix
	FirstIP       netip.Addr
	LastIP        netip.Addr
	BroadcastAddr netip.Addr
	SubnetMask    netip.Addr
	Hosts         uint
	Parent        *SubnetNode
	Children      []*SubnetNode
	IsSplit       bool
}

// ExportNode is a JSON-serializable representation of a subnet.
type ExportNode struct {
	CIDR          string        `json:"cidr"`
	FirstIP       string        `json:"firstIP"`
	LastIP        string        `json:"lastIP"`
	BroadcastAddr string        `json:"broadcastAddr"`
	SubnetMask    string        `json:"subnetMask"`
	Hosts         uint          `json:"hosts"`
	Children      []*ExportNode `json:"children,omitempty"`
}

// createSubnetNode creates a new subnet node from a prefix.
func createSubnetNode(prefix netip.Prefix, parent *SubnetNode) *SubnetNode {
	node := &SubnetNode{
		CIDR:   prefix,
		Parent: parent,
	}

	// Calculate network details using subnet package
	networkAddr := prefix.Masked().Addr()
	maskBits := prefix.Bits()
	addrBits := prefix.Addr().BitLen()

	// Calculate subnet mask using subnet package
	node.SubnetMask = subnet.CalculateSubnetMask(maskBits, addrBits)

	// Calculate broadcast address using subnet package
	node.BroadcastAddr = subnet.CalculateBroadcastAddr(networkAddr, node.SubnetMask)

	// First and last usable IPs
	node.FirstIP = networkAddr.Next()
	node.LastIP = node.BroadcastAddr.Prev()

	// Calculate hosts - use uint for TUI display (capped)
	hostBits := addrBits - maskBits
	if hostBits >= 2 && hostBits <= 32 {
		node.Hosts = (1 << hostBits) - 2
	} else if hostBits > 32 {
		// For IPv6, cap at max uint32 for display
		node.Hosts = ^uint(0) // Max uint value
	} else {
		node.Hosts = 0
	}

	return node
}

// Split splits a subnet node into two children.
func (n *SubnetNode) Split() bool {
	if n.IsSplit || n.CIDR.Bits() >= 30 {
		return false
	}

	newBits := n.CIDR.Bits() + 1
	networkAddr := n.CIDR.Masked().Addr()

	// First child: same network address, smaller prefix
	child1Prefix := netip.PrefixFrom(networkAddr, newBits)
	n.Children = append(n.Children, createSubnetNode(child1Prefix, n))

	// Second child: next network address
	child1Broadcast := n.Children[0].BroadcastAddr
	child2Addr := child1Broadcast.Next()
	child2Prefix := netip.PrefixFrom(child2Addr, newBits)
	n.Children = append(n.Children, createSubnetNode(child2Prefix, n))

	n.IsSplit = true
	return true
}

// SplitToDepth recursively splits the subnet until reaching the target bit depth.
func (n *SubnetNode) SplitToDepth(targetBits int) {
	if n.CIDR.Bits() >= targetBits {
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

// Join merges the children back into the parent.
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
		CIDR:          n.CIDR.String(),
		FirstIP:       n.FirstIP.String(),
		LastIP:        n.LastIP.String(),
		BroadcastAddr: n.BroadcastAddr.String(),
		SubnetMask:    n.SubnetMask.String(),
		Hosts:         n.Hosts,
	}

	for _, child := range n.Children {
		export.Children = append(export.Children, child.toExportNode())
	}

	return export
}

// ExportJSON returns the tree as a JSON string.
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
	if node.CIDR.Bits() == targetBits {
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
