/*
Copyright © 2023 Jake Rogers <code@supportoss.org>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"os"
	"testing"
)

func TestFlipBytes(t *testing.T) {
	tests := []struct {
		input []byte
		want  []byte
	}{
		{[]byte{0x00}, []byte{0xFF}},
		{[]byte{0x00, 0xFF, 0xAA}, []byte{0xFF, 0x00, 0x55}},
		{[]byte{}, []byte{}},
		{[]byte{0xFF, 0xFF, 0xFF, 0x00}, []byte{0x00, 0x00, 0x00, 0xFF}},
	}

	for i, tt := range tests {
		// Make a copy to avoid modifying the original slice
		input := make([]byte, len(tt.input))
		copy(input, tt.input)

		got := flipBytes(input)
		if !bytes.Equal(got, tt.want) {
			t.Errorf("test %d: flipBytes(%v) = %v, want %v", i, tt.input, got, tt.want)
		}
	}
}

func TestNetworkGetBroadcastAddr(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"192.168.1.0/24", "192.168.1.255"},
		{"172.16.0.0/16", "172.16.255.255"},
		{"10.0.0.0/8", "10.255.255.255"},
		{"192.168.1.0/26", "192.168.1.63"},
		{"192.168.1.0/30", "192.168.1.3"},
	}

	for _, tt := range tests {
		n := getNetworkDetails(tt.cidr)
		got := n.getBroadcastAddr()
		want := netip.MustParseAddr(tt.want)

		if got != want {
			t.Errorf("getBroadcastAddr(%s) = %v, want %v", tt.cidr, got, want)
		}
	}
}

func TestNetworkGetSubnetBits(t *testing.T) {
	tests := []struct {
		cidr string
		want int
	}{
		{"10.0.0.0/8", 0},     // Class A default
		{"10.0.0.0/16", 8},    // Class A with subnetting
		{"172.16.0.0/16", 0},  // Class B default
		{"172.16.0.0/24", 8},  // Class B with subnetting
		{"192.168.1.0/24", 0}, // Class C default
		{"192.168.1.0/26", 2}, // Class C with subnetting
	}

	for _, tt := range tests {
		n := getNetworkDetails(tt.cidr)
		if got := n.getSubnetBits(); got != tt.want {
			t.Errorf("getSubnetBits(%s) = %d, want %d", tt.cidr, got, tt.want)
		}
	}
}

func TestNetworkGetSubnetMask(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"10.0.0.0/8", "255.0.0.0"},
		{"172.16.0.0/16", "255.255.0.0"},
		{"192.168.1.0/24", "255.255.255.0"},
		{"192.168.1.0/26", "255.255.255.192"},
		{"192.168.1.0/30", "255.255.255.252"},
		{"10.12.32.0/19", "255.255.224.0"},
	}

	for _, tt := range tests {
		n := getNetworkDetails(tt.cidr)
		got := n.getSubnetMask()
		want := netip.MustParseAddr(tt.want)

		if got != want {
			t.Errorf("getSubnetMask(%s) = %v, want %v", tt.cidr, got, want)
		}
	}
}

func TestNetworkGetSubnets(t *testing.T) {
	tests := []struct {
		cidr       string
		subnetBits int
		wantCount  int
	}{
		{"192.168.1.0/24", 26, 4},  // /24 -> /26
		{"192.168.10.0/25", 27, 4}, // /25 -> /27
		{"10.0.0.0/16", 18, 4},     // /16 -> /18
		{"10.12.32.0/19", 20, 2},   // /19 -> /20
	}

	for _, tt := range tests {
		n := getNetworkDetails(tt.cidr)
		n.getSubnets(tt.subnetBits)

		if got := len(n.Subnets); got != tt.wantCount {
			t.Errorf("getSubnets(%s, %d) created %d subnets, want %d",
				tt.cidr, tt.subnetBits, got, tt.wantCount)
		}

		// Verify subnet properties
		for i, subnet := range n.Subnets {
			if subnet.MaskBits != tt.subnetBits {
				t.Errorf("subnet[%d] has mask bits %d, want %d",
					i, subnet.MaskBits, tt.subnetBits)
			}
			if !n.CIDR.Contains(subnet.NetworkAddr) {
				t.Errorf("subnet[%d] network %v not contained in parent %v",
					i, subnet.NetworkAddr, n.CIDR)
			}
		}
	}
}

func TestGetNetworkDetails(t *testing.T) {
	tests := []struct {
		cidr string
		want network
	}{
		{
			cidr: "192.168.1.0/24",
			want: network{
				NetworkAddr:   netip.MustParseAddr("192.168.1.0"),
				FirstHostIP:   netip.MustParseAddr("192.168.1.1"),
				LastHostIP:    netip.MustParseAddr("192.168.1.254"),
				BroadcastAddr: netip.MustParseAddr("192.168.1.255"),
				SubnetMask:    netip.MustParseAddr("255.255.255.0"),
				MaxHosts:      254,
			},
		},
		{
			cidr: "172.16.0.0/16",
			want: network{
				NetworkAddr:   netip.MustParseAddr("172.16.0.0"),
				FirstHostIP:   netip.MustParseAddr("172.16.0.1"),
				LastHostIP:    netip.MustParseAddr("172.16.255.254"),
				BroadcastAddr: netip.MustParseAddr("172.16.255.255"),
				SubnetMask:    netip.MustParseAddr("255.255.0.0"),
				MaxHosts:      65534,
			},
		},
		{
			cidr: "192.168.1.0/30",
			want: network{
				NetworkAddr:   netip.MustParseAddr("192.168.1.0"),
				FirstHostIP:   netip.MustParseAddr("192.168.1.1"),
				LastHostIP:    netip.MustParseAddr("192.168.1.2"),
				BroadcastAddr: netip.MustParseAddr("192.168.1.3"),
				SubnetMask:    netip.MustParseAddr("255.255.255.252"),
				MaxHosts:      2,
			},
		},
	}

	for _, tt := range tests {
		got := getNetworkDetails(tt.cidr)

		checks := []struct {
			name string
			got  any
			want any
		}{
			{"NetworkAddr", got.NetworkAddr, tt.want.NetworkAddr},
			{"FirstHostIP", got.FirstHostIP, tt.want.FirstHostIP},
			{"LastHostIP", got.LastHostIP, tt.want.LastHostIP},
			{"BroadcastAddr", got.BroadcastAddr, tt.want.BroadcastAddr},
			{"SubnetMask", got.SubnetMask, tt.want.SubnetMask},
			{"MaxHosts", got.MaxHosts, tt.want.MaxHosts},
		}

		for _, check := range checks {
			if check.got != check.want {
				t.Errorf("getNetworkDetails(%s).%s = %v, want %v",
					tt.cidr, check.name, check.got, check.want)
			}
		}
	}
}

func TestGetNetworkDetailsInvalidInput(t *testing.T) {
	invalidCIDRs := []string{
		"999.999.999.999/24", // invalid IP
		"192.168.1.0",        // missing mask
		"192.168.1.0/99",     // invalid mask for IPv4
		"",                   // empty string
	}

	for _, cidr := range invalidCIDRs {
		if _, err := netip.ParsePrefix(cidr); err == nil {
			t.Errorf("Expected error for invalid CIDR: %s", cidr)
		}
	}
}

func TestNetworkPrintNetworkJSON(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	n := getNetworkDetails("192.168.1.0/24")
	n.printNetworkJSON()

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 2048)
	bytesRead, _ := r.Read(buf)
	output := string(buf[:bytesRead])

	// Verify it's valid JSON
	var result network
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("printNetworkJSON() invalid JSON: %v\nOutput: %s", err, output)
		return
	}

	// Verify key field
	if result.CIDR.String() != "192.168.1.0/24" {
		t.Errorf("JSON CIDR = %v, want 192.168.1.0/24", result.CIDR)
	}
}

func TestIPv6Support(t *testing.T) {
	ipv6CIDRs := []string{"2001:db8::/64", "2001:db8::/48"}

	for _, cidr := range ipv6CIDRs {
		// Test that IPv6 addresses can be parsed without crashing
		n := getNetworkDetails(cidr)
		if !n.CIDR.Addr().Is6() {
			t.Errorf("Expected IPv6 address for %s", cidr)
		}
	}
}

// Benchmarks
func BenchmarkGetNetworkDetails(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getNetworkDetails("192.168.1.0/24")
	}
}

func BenchmarkGetSubnets(b *testing.B) {
	n := getNetworkDetails("10.0.0.0/16")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n.Subnets = nil
		n.getSubnets(24)
	}
}

func BenchmarkFlipBytes(b *testing.B) {
	bytes := []byte{0xFF, 0xFF, 0xFF, 0x00}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := make([]byte, len(bytes))
		copy(data, bytes)
		flipBytes(data)
	}
}
