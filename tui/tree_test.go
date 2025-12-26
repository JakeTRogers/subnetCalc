package tui

import (
	"encoding/json"
	"net/netip"
	"testing"
)

func TestCreateSubnetNode_IPv4(t *testing.T) {
	tests := []struct {
		name          string
		cidr          string
		wantFirstIP   string
		wantLastIP    string
		wantBroadcast string
		wantMask      string
		wantHosts     uint
	}{
		{
			name:          "IPv4 /24",
			cidr:          "192.168.1.0/24",
			wantFirstIP:   "192.168.1.1",
			wantLastIP:    "192.168.1.254",
			wantBroadcast: "192.168.1.255",
			wantMask:      "255.255.255.0",
			wantHosts:     254,
		},
		{
			name:          "IPv4 /25",
			cidr:          "10.0.0.0/25",
			wantFirstIP:   "10.0.0.1",
			wantLastIP:    "10.0.0.126",
			wantBroadcast: "10.0.0.127",
			wantMask:      "255.255.255.128",
			wantHosts:     126,
		},
		{
			name:          "IPv4 /30",
			cidr:          "172.16.0.0/30",
			wantFirstIP:   "172.16.0.1",
			wantLastIP:    "172.16.0.2",
			wantBroadcast: "172.16.0.3",
			wantMask:      "255.255.255.252",
			wantHosts:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := createSubnetNode(netip.MustParsePrefix(tt.cidr), nil)
			if got := node.FirstIP.String(); got != tt.wantFirstIP {
				t.Fatalf("FirstIP = %s, want %s", got, tt.wantFirstIP)
			}
			if got := node.LastIP.String(); got != tt.wantLastIP {
				t.Fatalf("LastIP = %s, want %s", got, tt.wantLastIP)
			}
			if got := node.BroadcastAddr.String(); got != tt.wantBroadcast {
				t.Fatalf("BroadcastAddr = %s, want %s", got, tt.wantBroadcast)
			}
			if got := node.SubnetMask.String(); got != tt.wantMask {
				t.Fatalf("SubnetMask = %s, want %s", got, tt.wantMask)
			}
			if got := node.Hosts; got != tt.wantHosts {
				t.Fatalf("Hosts = %d, want %d", got, tt.wantHosts)
			}
		})
	}
}

func TestCreateSubnetNode_IPv6_HostsCapped(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("2001:db8::/64"), nil)
	if node.CIDR.Addr().Is4() {
		t.Fatalf("expected IPv6 node")
	}
	if node.Hosts != ^uint(0) {
		t.Fatalf("Hosts = %d, want max uint", node.Hosts)
	}
}

func TestSubnetNodeSplit(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("192.168.1.0/24"), nil)
	if ok := node.Split(); !ok {
		t.Fatalf("Split() = false, want true")
	}
	if !node.IsSplit {
		t.Fatalf("IsSplit = false, want true")
	}
	if len(node.Children) != 2 {
		t.Fatalf("Children = %d, want 2", len(node.Children))
	}
	if got, want := node.Children[0].CIDR.String(), "192.168.1.0/25"; got != want {
		t.Fatalf("child1 CIDR = %s, want %s", got, want)
	}
	if got, want := node.Children[1].CIDR.String(), "192.168.1.128/25"; got != want {
		t.Fatalf("child2 CIDR = %s, want %s", got, want)
	}

	// Split again should be a no-op.
	if ok := node.Split(); ok {
		t.Fatalf("second Split() = true, want false")
	}
}

func TestSubnetNodeSplitLimit(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("192.168.1.0/30"), nil)
	if ok := node.Split(); ok {
		t.Fatalf("Split(/30) = true, want false")
	}
}

func TestSubnetNodeJoin(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("192.168.1.0/24"), nil)
	_ = node.Split()
	if ok := node.Join(); !ok {
		t.Fatalf("Join() = false, want true")
	}
	if node.IsSplit {
		t.Fatalf("IsSplit = true, want false")
	}
	if len(node.Children) != 0 {
		t.Fatalf("Children = %d, want 0", len(node.Children))
	}
}

func TestSubnetNodeSplitToDepthAndCollectLeaves(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("192.168.0.0/24"), nil)
	node.SplitToDepth(26)

	var leaves []*SubnetNode
	collectLeaves(node, &leaves)
	if len(leaves) != 4 {
		t.Fatalf("leaves = %d, want 4", len(leaves))
	}

	want := []string{
		"192.168.0.0/26",
		"192.168.0.64/26",
		"192.168.0.128/26",
		"192.168.0.192/26",
	}
	for i := range want {
		if got := leaves[i].CIDR.String(); got != want[i] {
			t.Fatalf("leaf[%d] CIDR = %s, want %s", i, got, want[i])
		}
	}
}

func TestGetAncestorAtDepth(t *testing.T) {
	root := createSubnetNode(netip.MustParsePrefix("192.168.0.0/24"), nil)
	root.Split()                 // /25
	_ = root.Children[0].Split() // /26 under first /25

	leaf := root.Children[0].Children[0]
	if leaf.CIDR.Bits() != 26 {
		t.Fatalf("expected /26 leaf, got /%d", leaf.CIDR.Bits())
	}

	tests := []struct {
		name       string
		targetBits int
		wantBits   int
	}{
		{name: "self", targetBits: 26, wantBits: 26},
		{name: "parent", targetBits: 25, wantBits: 25},
		{name: "root", targetBits: 24, wantBits: 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anc := getAncestorAtDepth(leaf, tt.targetBits)
			if anc == nil {
				t.Fatalf("ancestor is nil")
			}
			if got := anc.CIDR.Bits(); got != tt.wantBits {
				t.Fatalf("ancestor bits = %d, want %d", got, tt.wantBits)
			}
		})
	}
}

func TestExportJSON(t *testing.T) {
	node := createSubnetNode(netip.MustParsePrefix("192.168.1.0/24"), nil)
	jsonStr, err := node.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	var exported ExportNode
	if err := json.Unmarshal([]byte(jsonStr), &exported); err != nil {
		t.Fatalf("unmarshal export json: %v", err)
	}
	if exported.CIDR != "192.168.1.0/24" {
		t.Fatalf("exported CIDR = %q, want %q", exported.CIDR, "192.168.1.0/24")
	}
	if exported.FirstIP == "" || exported.LastIP == "" {
		t.Fatalf("exported first/last IP should not be empty")
	}
}
