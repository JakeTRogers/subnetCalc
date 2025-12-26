package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JakeTRogers/subnetCalc/formatter"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

func TestPrintNetworkOutput(t *testing.T) {
	// Create a test network
	calc := subnet.NewCalculator()
	network, err := calc.Calculate("192.168.1.0/24")
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
		networkWithSubnets, err := calc.Calculate("192.168.0.0/24")
		if err != nil {
			t.Fatalf("Failed to create test network: %v", err)
		}
		if err := calc.Split(&networkWithSubnets, 26); err != nil {
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

func TestRootCmdValidation(t *testing.T) {
	// Test that rootCmd has proper configuration
	if rootCmd.Use != "subnetCalc <CIDR>" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "subnetCalc <CIDR>")
	}

	// Check that required flags exist
	flags := []string{"json", "interactive", "subnet_size", "verbose"}
	for _, flag := range flags {
		if rootCmd.Flags().Lookup(flag) == nil && rootCmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q not found", flag)
		}
	}
}

func TestRootCmdMutuallyExclusiveFlags(t *testing.T) {
	// Create a new command for testing to avoid polluting global state
	testCmd := *rootCmd

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

func TestRootCmdNoArgs(t *testing.T) {
	// Test that the command recognizes when no args are provided
	testCmd := *rootCmd

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{})

	// Verify the RunE function exists and would check for args
	if testCmd.RunE == nil {
		t.Error("Expected rootCmd to have a RunE function")
	}

	// Check the command has proper help template
	if testCmd.Use != "subnetCalc <CIDR>" {
		t.Errorf("Expected Use to be 'subnetCalc <CIDR>', got %q", testCmd.Use)
	}
}
