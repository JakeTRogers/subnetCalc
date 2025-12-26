package formatter

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/JakeTRogers/subnetCalc/subnet"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		wantType string
	}{
		{
			name:     "JSON format",
			cfg:      Config{Format: FormatJSON, PrettyPrint: true},
			wantType: "*formatter.JSONFormatter",
		},
		{
			name:     "Table format",
			cfg:      Config{Format: FormatTable, Width: 120},
			wantType: "*formatter.TableFormatter",
		},
		{
			name:     "Text format",
			cfg:      Config{Format: FormatText},
			wantType: "*formatter.TextFormatter",
		},
		{
			name:     "Default config",
			cfg:      DefaultConfig(),
			wantType: "*formatter.TableFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New(tt.cfg)
			if f == nil {
				t.Fatal("New() returned nil")
			}

			// Verify formatter type by checking interface implementation
			switch tt.cfg.Format {
			case FormatJSON:
				if _, ok := f.(*JSONFormatter); !ok {
					t.Errorf("Expected JSONFormatter for JSON format")
				}
			case FormatTable:
				if _, ok := f.(*TableFormatter); !ok {
					t.Errorf("Expected TableFormatter for Table format")
				}
			case FormatText:
				if _, ok := f.(*TextFormatter); !ok {
					t.Errorf("Expected TextFormatter for Text format")
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Format != FormatTable {
		t.Errorf("Default format should be Table, got %v", cfg.Format)
	}
	if cfg.Width != 120 {
		t.Errorf("Default width should be 120, got %d", cfg.Width)
	}
	if !cfg.PrettyPrint {
		t.Error("Default PrettyPrint should be true")
	}
}

func TestJSONFormatter_FormatNetwork(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	tests := []struct {
		name       string
		indent     bool
		wantFields []string
	}{
		{
			name:       "With indentation",
			indent:     true,
			wantFields: []string{"cidr", "firstIP", "lastIP", "networkAddr", "broadcastAddr", "subnetMask", "maskBits", "maxHosts"},
		},
		{
			name:       "Without indentation",
			indent:     false,
			wantFields: []string{"cidr", "firstIP", "lastIP"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewJSONFormatter(tt.indent)
			output, err := f.FormatNetwork(n)
			if err != nil {
				t.Fatalf("FormatNetwork() error = %v", err)
			}

			var result map[string]any
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
			}

			for _, field := range tt.wantFields {
				if _, exists := result[field]; !exists {
					t.Errorf("Missing field: %s", field)
				}
			}

			if tt.indent && !strings.Contains(output, "\n") {
				t.Error("Expected indented output to contain newlines")
			}
		})
	}
}

func TestJSONFormatter_FormatSubnets(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	if err := n.Split(26); err != nil {
		t.Fatalf("Failed to split network: %v", err)
	}

	f := NewJSONFormatter(true)
	output, err := f.FormatSubnets(n)
	if err != nil {
		t.Fatalf("FormatSubnets() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	subnets, exists := result["subnets"]
	if !exists {
		t.Error("Missing 'subnets' field")
	}

	subnetSlice, ok := subnets.([]any)
	if !ok {
		t.Error("'subnets' should be an array")
	}

	if len(subnetSlice) != 4 {
		t.Errorf("Expected 4 subnets, got %d", len(subnetSlice))
	}
}

func TestTableFormatter_FormatNetwork(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	f := NewTableFormatter(120)
	output, err := f.FormatNetwork(n)
	if err != nil {
		t.Fatalf("FormatNetwork() error = %v", err)
	}

	expectedStrings := []string{
		"Network:",
		"192.168.1.0/24",
		"Host Address Range:",
		"Broadcast Address:",
		"Subnet Mask:",
		"Maximum Hosts:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestTableFormatter_FormatSubnets(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	if err := n.Split(26); err != nil {
		t.Fatalf("Failed to split network: %v", err)
	}

	f := NewTableFormatter(120)
	output, err := f.FormatSubnets(n)
	if err != nil {
		t.Fatalf("FormatSubnets() error = %v", err)
	}

	expectedStrings := []string{
		"192.168.1.0/24",
		"contains",
		"/26",
		"subnets",
		"Subnet",
		"Subnet Mask",
		"Hosts",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestTableFormatter_FormatSubnets_Empty(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	f := NewTableFormatter(120)
	output, err := f.FormatSubnets(n)
	if err != nil {
		t.Fatalf("FormatSubnets() error = %v", err)
	}

	if output != "" {
		t.Errorf("Expected empty output for network without subnets, got: %s", output)
	}
}

func TestTextFormatter_FormatNetwork(t *testing.T) {
	n, err := subnet.NewNetwork("10.0.0.0/8")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	f := NewTextFormatter()
	output, err := f.FormatNetwork(n)
	if err != nil {
		t.Fatalf("FormatNetwork() error = %v", err)
	}

	expectedStrings := []string{
		"Network:",
		"10.0.0.0/8",
		"Host Address Range:",
		"Maximum Hosts:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestTextFormatter_FormatSubnets(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	if err := n.Split(26); err != nil {
		t.Fatalf("Failed to split network: %v", err)
	}

	f := NewTextFormatter()
	output, err := f.FormatSubnets(n)
	if err != nil {
		t.Fatalf("FormatSubnets() error = %v", err)
	}

	if !strings.Contains(output, "Subnets (4 total)") {
		t.Error("Expected output to contain subnet count")
	}

	if !strings.Contains(output, "192.168.1.") {
		t.Error("Expected output to contain subnet CIDRs")
	}
}

func TestToNetworkInfo(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	info := ToNetworkInfo(n)

	if info.CIDR != "192.168.1.0/24" {
		t.Errorf("CIDR = %v, want 192.168.1.0/24", info.CIDR)
	}
	if info.NetworkAddr != "192.168.1.0" {
		t.Errorf("NetworkAddr = %v, want 192.168.1.0", info.NetworkAddr)
	}
	if info.BroadcastAddr != "192.168.1.255" {
		t.Errorf("BroadcastAddr = %v, want 192.168.1.255", info.BroadcastAddr)
	}
	if info.MaskBits != 24 {
		t.Errorf("MaskBits = %v, want 24", info.MaskBits)
	}
	if info.MaxHosts != "254" {
		t.Errorf("MaxHosts = %v, want 254", info.MaxHosts)
	}
}

func TestToSubnetInfo(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/26")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	info := ToSubnetInfo(n)

	if info.CIDR != "192.168.1.0/26" {
		t.Errorf("CIDR = %v, want 192.168.1.0/26", info.CIDR)
	}
	if info.SubnetMask != "255.255.255.192" {
		t.Errorf("SubnetMask = %v, want 255.255.255.192", info.SubnetMask)
	}
	if info.Hosts != "62" {
		t.Errorf("Hosts = %v, want 62", info.Hosts)
	}
}

func TestToSubnetInfoSlice(t *testing.T) {
	n, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	if err := n.Split(26); err != nil {
		t.Fatalf("Failed to split network: %v", err)
	}

	infos := ToSubnetInfoSlice(n.Subnets)

	if len(infos) != 4 {
		t.Errorf("Expected 4 subnet infos, got %d", len(infos))
	}

	if infos[0].CIDR != "192.168.1.0/26" {
		t.Errorf("First subnet CIDR = %v, want 192.168.1.0/26", infos[0].CIDR)
	}
	if infos[3].CIDR != "192.168.1.192/26" {
		t.Errorf("Last subnet CIDR = %v, want 192.168.1.192/26", infos[3].CIDR)
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"192.168.1.0/24", "24"},
		{"10.0.0.0/8", "8"},
		{"2001:db8::/64", "64"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		got := extractPrefix(tt.cidr)
		if got != tt.want {
			t.Errorf("extractPrefix(%s) = %v, want %v", tt.cidr, got, tt.want)
		}
	}
}

func TestFormatterInterface(t *testing.T) {
	var _ Formatter = (*JSONFormatter)(nil)
	var _ Formatter = (*TableFormatter)(nil)
	var _ Formatter = (*TextFormatter)(nil)
}

func TestJSONFormatter_LargeIPv6(t *testing.T) {
	n, err := subnet.NewNetwork("2001:db8::/32")
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	f := NewJSONFormatter(true)
	output, err := f.FormatNetwork(n)
	if err != nil {
		t.Fatalf("FormatNetwork() error = %v", err)
	}

	// JSON escapes > as \u003e
	if !strings.Contains(output, "2^") || !strings.Contains(output, "maxHosts") {
		t.Error("Expected IPv6 network to show capped host count")
	}

	// Parse JSON to verify actual value
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	maxHosts, ok := result["maxHosts"].(string)
	if !ok {
		t.Fatal("maxHosts should be a string")
	}
	if !strings.HasPrefix(maxHosts, ">2^") {
		t.Errorf("Expected maxHosts to start with '>2^', got %s", maxHosts)
	}
}

func TestFormatMaxHosts(t *testing.T) {
	tests := []struct {
		name     string
		maxHosts *big.Int
		want     string
	}{
		{
			name:     "Small number",
			maxHosts: big.NewInt(254),
			want:     "254",
		},
		{
			name:     "Thousands",
			maxHosts: big.NewInt(65534),
			want:     "65,534",
		},
		{
			name:     "Millions",
			maxHosts: big.NewInt(16777214),
			want:     "16,777,214",
		},
		{
			name:     "Zero",
			maxHosts: big.NewInt(0),
			want:     "0",
		},
		{
			name:     "Nil",
			maxHosts: nil,
			want:     "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMaxHosts(tt.maxHosts)
			if got != tt.want {
				t.Errorf("FormatMaxHosts() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test large IPv6 number
	t.Run("Large IPv6 number", func(t *testing.T) {
		// 2^64
		largeNum := new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)
		got := FormatMaxHosts(largeNum)
		if got != ">2^64" {
			t.Errorf("FormatMaxHosts(2^64) = %v, want >2^64", got)
		}
	})
}

func BenchmarkJSONFormatter(b *testing.B) {
	n, _ := subnet.NewNetwork("10.0.0.0/8")
	_ = n.Split(16)
	f := NewJSONFormatter(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = f.FormatSubnets(n)
	}
}

func BenchmarkTableFormatter(b *testing.B) {
	n, _ := subnet.NewNetwork("192.168.1.0/24")
	_ = n.Split(26)
	f := NewTableFormatter(120)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = f.FormatSubnets(n)
	}
}
