package subnet

import (
	"math/big"
	"net/netip"
	"testing"
)

func TestNewNetwork(t *testing.T) {
	tests := []struct {
		name            string
		cidr            string
		wantErr         bool
		wantNetworkAddr string
		wantBroadcast   string
		wantFirstIP     string
		wantLastIP      string
		wantMaskBits    int
	}{
		{
			name:            "IPv4 /24",
			cidr:            "192.168.1.0/24",
			wantNetworkAddr: "192.168.1.0",
			wantBroadcast:   "192.168.1.255",
			wantFirstIP:     "192.168.1.1",
			wantLastIP:      "192.168.1.254",
			wantMaskBits:    24,
		},
		{
			name:            "IPv4 /16",
			cidr:            "10.0.0.0/16",
			wantNetworkAddr: "10.0.0.0",
			wantBroadcast:   "10.0.255.255",
			wantFirstIP:     "10.0.0.1",
			wantLastIP:      "10.0.255.254",
			wantMaskBits:    16,
		},
		{
			name:            "IPv4 /8",
			cidr:            "10.0.0.0/8",
			wantNetworkAddr: "10.0.0.0",
			wantBroadcast:   "10.255.255.255",
			wantFirstIP:     "10.0.0.1",
			wantLastIP:      "10.255.255.254",
			wantMaskBits:    8,
		},
		{
			name:            "IPv4 normalized input",
			cidr:            "192.168.1.100/24",
			wantNetworkAddr: "192.168.1.0",
			wantBroadcast:   "192.168.1.255",
			wantFirstIP:     "192.168.1.1",
			wantLastIP:      "192.168.1.254",
			wantMaskBits:    24,
		},
		{
			name:            "IPv6 /64",
			cidr:            "2001:db8::/64",
			wantNetworkAddr: "2001:db8::",
			wantBroadcast:   "2001:db8::ffff:ffff:ffff:ffff",
			wantFirstIP:     "2001:db8::1",
			wantLastIP:      "2001:db8::ffff:ffff:ffff:fffe",
			wantMaskBits:    64,
		},
		{
			name:    "Invalid CIDR",
			cidr:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := NewNetwork(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if n.NetworkAddr.String() != tt.wantNetworkAddr {
				t.Errorf("NetworkAddr = %v, want %v", n.NetworkAddr, tt.wantNetworkAddr)
			}
			if n.BroadcastAddr.String() != tt.wantBroadcast {
				t.Errorf("BroadcastAddr = %v, want %v", n.BroadcastAddr, tt.wantBroadcast)
			}
			if n.FirstHostIP.String() != tt.wantFirstIP {
				t.Errorf("FirstHostIP = %v, want %v", n.FirstHostIP, tt.wantFirstIP)
			}
			if n.LastHostIP.String() != tt.wantLastIP {
				t.Errorf("LastHostIP = %v, want %v", n.LastHostIP, tt.wantLastIP)
			}
			if n.MaskBits != tt.wantMaskBits {
				t.Errorf("MaskBits = %v, want %v", n.MaskBits, tt.wantMaskBits)
			}
		})
	}
}

func TestNetworkSplit(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		targetBits  int
		wantErr     bool
		wantCount   int
		wantFirstIP string
		wantLastIP  string
	}{
		{
			name:        "Split /24 into /26",
			cidr:        "192.168.1.0/24",
			targetBits:  26,
			wantCount:   4,
			wantFirstIP: "192.168.1.0",
			wantLastIP:  "192.168.1.192",
		},
		{
			name:        "Split /16 into /24",
			cidr:        "10.0.0.0/16",
			targetBits:  24,
			wantCount:   256,
			wantFirstIP: "10.0.0.0",
			wantLastIP:  "10.0.255.0",
		},
		{
			name:       "Invalid: target smaller than network",
			cidr:       "192.168.1.0/24",
			targetBits: 16,
			wantErr:    true,
		},
		{
			name:       "Invalid: target equals network",
			cidr:       "192.168.1.0/24",
			targetBits: 24,
			wantErr:    true,
		},
		{
			name:       "Invalid: would generate too many subnets",
			cidr:       "0.0.0.0/0",
			targetBits: 31,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := NewNetwork(tt.cidr)
			if err != nil {
				t.Fatalf("NewNetwork() error = %v", err)
			}

			err = n.Split(tt.targetBits)
			if (err != nil) != tt.wantErr {
				t.Errorf("Split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(n.Subnets) != tt.wantCount {
				t.Errorf("Split() created %d subnets, want %d", len(n.Subnets), tt.wantCount)
			}

			if n.Subnets[0].NetworkAddr.String() != tt.wantFirstIP {
				t.Errorf("First subnet NetworkAddr = %v, want %v", n.Subnets[0].NetworkAddr, tt.wantFirstIP)
			}

			if n.Subnets[len(n.Subnets)-1].NetworkAddr.String() != tt.wantLastIP {
				t.Errorf("Last subnet NetworkAddr = %v, want %v", n.Subnets[len(n.Subnets)-1].NetworkAddr, tt.wantLastIP)
			}

			// Verify each subnet has correct mask bits
			for i, subnet := range n.Subnets {
				if subnet.MaskBits != tt.targetBits {
					t.Errorf("Subnet %d MaskBits = %d, want %d", i, subnet.MaskBits, tt.targetBits)
				}
			}
		})
	}
}

func TestCalculateMaxHosts(t *testing.T) {
	tests := []struct {
		name     string
		addrBits int
		maskBits int
		want     *big.Int
	}{
		{
			name:     "/24 network",
			addrBits: 32,
			maskBits: 24,
			want:     big.NewInt(254),
		},
		{
			name:     "/16 network",
			addrBits: 32,
			maskBits: 16,
			want:     big.NewInt(65534),
		},
		{
			name:     "/8 network",
			addrBits: 32,
			maskBits: 8,
			want:     big.NewInt(16777214),
		},
		{
			name:     "/32 network",
			addrBits: 32,
			maskBits: 32,
			want:     big.NewInt(0),
		},
		{
			name:     "/31 network",
			addrBits: 32,
			maskBits: 31,
			want:     big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMaxHosts(tt.addrBits, tt.maskBits)
			if got.Cmp(tt.want) != 0 {
				t.Errorf("CalculateMaxHosts(%d, %d) = %v, want %v", tt.addrBits, tt.maskBits, got, tt.want)
			}
		})
	}
}

func TestCalculateSubnetMask(t *testing.T) {
	tests := []struct {
		name     string
		maskBits int
		addrBits int
		want     string
	}{
		{
			name:     "/24 IPv4",
			maskBits: 24,
			addrBits: 32,
			want:     "255.255.255.0",
		},
		{
			name:     "/16 IPv4",
			maskBits: 16,
			addrBits: 32,
			want:     "255.255.0.0",
		},
		{
			name:     "/8 IPv4",
			maskBits: 8,
			addrBits: 32,
			want:     "255.0.0.0",
		},
		{
			name:     "/26 IPv4",
			maskBits: 26,
			addrBits: 32,
			want:     "255.255.255.192",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSubnetMask(tt.maskBits, tt.addrBits)
			if got.String() != tt.want {
				t.Errorf("CalculateSubnetMask(%d, %d) = %v, want %v", tt.maskBits, tt.addrBits, got, tt.want)
			}
		})
	}
}

func TestCalculateBroadcastAddr(t *testing.T) {
	tests := []struct {
		name        string
		networkAddr string
		subnetMask  string
		want        string
	}{
		{
			name:        "/24 network",
			networkAddr: "192.168.1.0",
			subnetMask:  "255.255.255.0",
			want:        "192.168.1.255",
		},
		{
			name:        "/16 network",
			networkAddr: "10.0.0.0",
			subnetMask:  "255.255.0.0",
			want:        "10.0.255.255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkAddr := netip.MustParseAddr(tt.networkAddr)
			subnetMask := netip.MustParseAddr(tt.subnetMask)
			got := CalculateBroadcastAddr(networkAddr, subnetMask)
			if got.String() != tt.want {
				t.Errorf("CalculateBroadcastAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculator(t *testing.T) {
	calc := NewCalculator()

	t.Run("Calculate", func(t *testing.T) {
		n, err := calc.Calculate("192.168.1.0/24")
		if err != nil {
			t.Fatalf("Calculate() error = %v", err)
		}
		if n.MaskBits != 24 {
			t.Errorf("MaskBits = %v, want 24", n.MaskBits)
		}
	})

	t.Run("Split", func(t *testing.T) {
		n, _ := calc.Calculate("192.168.1.0/24")
		err := calc.Split(&n, 26)
		if err != nil {
			t.Fatalf("Split() error = %v", err)
		}
		if len(n.Subnets) != 4 {
			t.Errorf("Split() created %d subnets, want 4", len(n.Subnets))
		}
	})
}

func BenchmarkNewNetwork(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewNetwork("192.168.1.0/24")
	}
}

func BenchmarkSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n, _ := NewNetwork("10.0.0.0/8")
		_ = n.Split(16)
	}
}
