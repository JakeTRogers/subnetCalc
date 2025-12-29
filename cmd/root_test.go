package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JakeTRogers/subnetCalc/formatter"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

func TestPrintNetworkOutput_formats(t *testing.T) {
	t.Parallel()
	// Create a test network
	network, err := subnet.NewNetwork("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	t.Run("JSON formatter outputs valid JSON", func(t *testing.T) {
		f := formatter.NewJSONFormatter(true)

		var buf bytes.Buffer
		err := printNetworkOutput(&buf, formatter.FormatJSON, f, network)
		if err != nil {
			t.Fatalf("printNetworkOutput() error = %v", err)
		}
		output := buf.String()

		// Verify it contains expected JSON fields
		if !strings.Contains(output, `"cidr"`) {
			t.Errorf("JSON output should contain 'cidr' field, got: %s", output)
		}
		if !strings.Contains(output, `"192.168.1.0"`) {
			t.Error("JSON output should contain network address")
		}
		if !strings.Contains(output, `"maskBits"`) {
			t.Errorf("JSON output should contain 'maskBits' field, got: %s", output)
		}
	})

	t.Run("Table formatter outputs formatted table", func(t *testing.T) {
		f := formatter.NewTableFormatter(120)

		var buf bytes.Buffer
		err := printNetworkOutput(&buf, formatter.FormatTable, f, network)
		if err != nil {
			t.Fatalf("printNetworkOutput() error = %v", err)
		}
		output := buf.String()

		// Verify it contains expected table content
		if !strings.Contains(output, "192.168.1.0") {
			t.Error("Table output should contain network address")
		}
	})

	t.Run("Network with subnets outputs subnet table", func(t *testing.T) {
		// Create network with subnets
		networkWithSubnets, err := subnet.NewNetwork("192.168.0.0/24")
		if err != nil {
			t.Fatalf("Failed to create test network: %v", err)
		}
		if err := networkWithSubnets.Split(26); err != nil {
			t.Fatalf("Failed to split network: %v", err)
		}

		f := formatter.NewTableFormatter(120)

		var buf bytes.Buffer
		err = printNetworkOutput(&buf, formatter.FormatTable, f, networkWithSubnets)
		if err != nil {
			t.Fatalf("printNetworkOutput() error = %v", err)
		}
		subnetsOutput := buf.String()

		// Should have 4 subnets for /24 split to /26
		if !strings.Contains(subnetsOutput, "192.168.0.0") {
			t.Error("Subnet output should contain first subnet")
		}
		if !strings.Contains(subnetsOutput, "192.168.0.64") {
			t.Error("Subnet output should contain second subnet")
		}
	})
}

func TestRootCmd_validation(t *testing.T) {
	t.Parallel()
	// Create a fresh command instance for isolated testing
	cmd := NewRootCommand()

	if cmd.Use != "subnetCalc <CIDR>" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "subnetCalc <CIDR>")
	}

	// Check that required flags exist
	flags := []string{"json", "interactive", "subnet-size", "verbose"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil && cmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q not found", flag)
		}
	}
}

func TestRootCmd_mutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	// Set both flags
	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"192.168.1.0/24", "-i", "-j"})

	// Execute should fail due to mutually exclusive flags
	err := testCmd.Execute()
	if err == nil {
		t.Error("Expected error when using mutually exclusive flags -i and -j together")
	}
}

func TestRootCmd_noArgs(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{})

	// Execute with no args should show help (not error)
	err := testCmd.Execute()
	if err != nil {
		t.Errorf("Expected no error when no args provided, got: %v", err)
	}

	// Verify help output contains usage information
	output := buf.String()
	if !strings.Contains(output, "subnetCalc <CIDR>") {
		t.Error("Expected help output to contain usage information")
	}
	if !strings.Contains(output, "Examples:") {
		t.Error("Expected help output to contain examples")
	}
}

func TestRootCmd_validCIDR(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"192.168.1.0/24"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for valid CIDR, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "192.168.1.0") {
		t.Error("Output should contain network address")
	}
	if !strings.Contains(output, "Network:") {
		t.Error("Output should contain Network label")
	}
}

func TestRootCmd_invalidCIDR(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"invalid-cidr"})

	err := testCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid CIDR")
	}
	if !strings.Contains(err.Error(), "parsing network") {
		t.Errorf("Expected 'parsing network' in error, got: %v", err)
	}
}

func TestRootCmd_subnetSplit(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"192.168.0.0/24", "-s", "26"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for subnet split, got: %v", err)
	}

	output := buf.String()
	// /24 split to /26 produces 4 subnets
	if !strings.Contains(output, "192.168.0.0") {
		t.Error("Output should contain first subnet")
	}
	if !strings.Contains(output, "192.168.0.64") {
		t.Error("Output should contain second subnet")
	}
	if !strings.Contains(output, "192.168.0.128") {
		t.Error("Output should contain third subnet")
	}
	if !strings.Contains(output, "192.168.0.192") {
		t.Error("Output should contain fourth subnet")
	}
}

func TestRootCmd_invalidSubnetSplit(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	// Subnet size smaller than network prefix is invalid
	testCmd.SetArgs([]string{"192.168.0.0/24", "-s", "16"})

	err := testCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid subnet split (size smaller than network)")
	}
	if !strings.Contains(err.Error(), "splitting network") {
		t.Errorf("Expected 'splitting network' in error, got: %v", err)
	}
}

func TestRootCmd_jsonOutput(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"10.0.0.0/8", "-j"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for JSON output, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"cidr"`) {
		t.Error("JSON output should contain 'cidr' field")
	}
	if !strings.Contains(output, `"networkAddr"`) {
		t.Error("JSON output should contain 'networkAddr' field")
	}
	if !strings.Contains(output, `"10.0.0.0"`) {
		t.Error("JSON output should contain network address")
	}
}

func TestRootCmd_jsonWithSubnets(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"192.168.0.0/30", "-s", "31", "-j"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for JSON output with subnets, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"subnets"`) {
		t.Error("JSON output should contain 'subnets' field")
	}
}

func TestRootCmd_IPv6(t *testing.T) {
	t.Parallel()
	testCmd := NewRootCommand()

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"2001:db8::/32"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for IPv6 CIDR, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "2001:db8::") {
		t.Error("Output should contain IPv6 network address")
	}
}

func TestTerminalWidth_nonFile(t *testing.T) {
	t.Parallel()
	// bytes.Buffer is not an *os.File, so should return fallback
	var buf bytes.Buffer
	width := terminalWidth(&buf, 120)
	if width != 120 {
		t.Errorf("Expected fallback width 120, got %d", width)
	}
}

func TestExecute_usesRootCmd(t *testing.T) {
	// This test verifies Execute() uses the package-level rootCmd
	// We can't easily test this in isolation without affecting global state,
	// but we verify it returns an error type (not calling os.Exit)

	// Save original args and restore after test
	originalArgs := rootCmd.Args
	defer func() { rootCmd.Args = originalArgs }()

	// Temporarily allow any args for this test
	rootCmd.SetArgs([]string{"192.168.1.0/24"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := Execute()
	if err != nil {
		t.Errorf("Execute() returned unexpected error: %v", err)
	}
}
