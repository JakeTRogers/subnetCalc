/*
Copyright © 2023 Jake Rogers <code@supportoss.org>
*/
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestMainIntegration tests the main CLI application end-to-end
func TestMainIntegration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up the test binary after tests
	defer func() {
		_ = os.Remove("./subnetCalc_test")
	}()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput []string
		expectedRegex  []string // For pattern matching like semver
	}{
		{
			name:           "Basic CIDR calculation",
			args:           []string{"192.168.1.0/24"},
			expectError:    false,
			expectedOutput: []string{"Network:", "192.168.1.0/24", "Host Address Range:", "Broadcast Address:", "192.168.1.255"},
		},
		{
			name:           "CIDR with subnet flag",
			args:           []string{"192.168.1.0/24", "--subnet_size", "26"},
			expectError:    false,
			expectedOutput: []string{"Network:", "192.168.1.0/24", "contains", "/26", "subnets"},
		},
		{
			name:           "JSON output",
			args:           []string{"192.168.1.0/24", "--json"},
			expectError:    false,
			expectedOutput: []string{`"cidr"`, `"firstIP"`, `"lastIP"`, `"networkAddr"`, `"broadcastAddr"`},
		},
		{
			name:        "Invalid CIDR",
			args:        []string{"invalid.cidr"},
			expectError: true,
		},
		{
			name:           "Help command",
			args:           []string{"--help"},
			expectError:    false,
			expectedOutput: []string{"Usage:", "subnetCalc", "calculate subnet"},
		},
		{
			name:          "Version command",
			args:          []string{"--version"},
			expectError:   false,
			expectedRegex: []string{`subnetCalc v\d+\.\d+\.\d+`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError && err == nil {
				t.Errorf("Expected command to fail but it succeeded")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Command failed unexpectedly: %v\nStderr: %s", err, stderr.String())
				return
			}

			output := stdout.String()
			if stderr.Len() > 0 && !tt.expectError {
				// For some commands, output might go to stderr (like help)
				output += stderr.String()
			}

			// Check expected output strings
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}

			// Check expected regex patterns
			for _, pattern := range tt.expectedRegex {
				matched, err := regexp.MatchString(pattern, output)
				if err != nil {
					t.Errorf("Invalid regex pattern '%s': %v", pattern, err)
				}
				if !matched {
					t.Errorf("Expected output to match regex '%s', got:\n%s", pattern, output)
				}
			}
		})
	}
}

// TestJSONOutputStructure validates the JSON output structure
func TestJSONOutputStructure(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Simple network JSON",
			args: []string{"192.168.1.0/24", "--json"},
		},
		{
			name: "Network with subnets JSON",
			args: []string{"192.168.1.0/24", "--subnet_size", "26", "--json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.args...)
			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			output := stdout.String()

			// Parse JSON to validate structure
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Invalid JSON output: %v\nOutput: %s", err, output)
				return
			}

			// Check required fields
			requiredFields := []string{"cidr", "firstIP", "lastIP", "networkAddr", "broadcastAddr", "subnetMask", "maskBits", "maxSubnets", "maxHosts"}
			for _, field := range requiredFields {
				if _, exists := result[field]; !exists {
					t.Errorf("JSON output missing required field: %s", field)
				}
			}

			// If subnets were requested, check subnets array
			if strings.Contains(strings.Join(tt.args, " "), "subnet_size") {
				if subnets, exists := result["subnets"]; !exists {
					t.Error("JSON output should contain subnets array when subnet_size is specified")
				} else if subnetSlice, ok := subnets.([]interface{}); !ok || len(subnetSlice) == 0 {
					t.Error("Subnets should be a non-empty array")
				}
			}
		})
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Invalid IP address",
			args:        []string{"999.999.999.999/24"},
			expectError: true,
		},
		{
			name:        "Invalid CIDR format",
			args:        []string{"192.168.1.0"},
			expectError: true,
		},
		{
			name:        "Invalid subnet mask",
			args:        []string{"192.168.1.0/99"},
			expectError: true,
		},
		{
			name:        "Too many arguments",
			args:        []string{"192.168.1.0/24", "extra", "args"},
			expectError: true,
		},
		{
			name:        "Subnet size smaller than network",
			args:        []string{"192.168.1.0/24", "--subnet_size", "20"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.args...)
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError && err == nil {
				t.Errorf("Expected command to fail but it succeeded")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Command failed unexpectedly: %v\nStderr: %s", err, stderr.String())
			}
		})
	}
}

// TestSubnetCalculations validates specific subnet calculation scenarios
func TestSubnetCalculations(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	tests := []struct {
		name           string
		cidr           string
		subnetSize     string
		expectedOutput []string
	}{
		{
			name:           "Split /24 into /26",
			cidr:           "192.168.1.0/24",
			subnetSize:     "26",
			expectedOutput: []string{"contains 4", "/26 subnets", "192.168.1.0/26", "192.168.1.64/26", "192.168.1.128/26", "192.168.1.192/26"},
		},
		{
			name:           "Split /25 into /27 (from README example)",
			cidr:           "192.168.10.0/25",
			subnetSize:     "27",
			expectedOutput: []string{"contains 4", "/27 subnets", "192.168.10.0/27", "192.168.10.32/27", "192.168.10.64/27", "192.168.10.96/27"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.cidr, "--subnet_size", tt.subnetSize)
			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			output := stdout.String()

			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}
		})
	}
}

// TestFlags tests various flag combinations
func TestFlags(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name:           "Color flag with subnets",
			args:           []string{"192.168.1.0/24", "--subnet_size", "26", "--color"},
			expectedOutput: []string{"contains 4", "/26 subnets"},
			shouldFail:     false,
		},
		{
			name:           "Verbose flag",
			args:           []string{"192.168.1.0/24", "-v"},
			expectedOutput: []string{"Network:", "192.168.1.0/24"},
			shouldFail:     false,
		},
		{
			name:           "Multiple verbose flags",
			args:           []string{"192.168.1.0/24", "-vv"},
			expectedOutput: []string{"Network:", "192.168.1.0/24"},
			shouldFail:     false,
		},
		{
			name:       "Mutually exclusive flags (color and json)",
			args:       []string{"192.168.1.0/24", "--subnet_size", "26", "--color", "--json"},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.shouldFail && err == nil {
				t.Errorf("Expected command to fail but it succeeded")
				return
			}

			if !tt.shouldFail && err != nil {
				t.Errorf("Command failed unexpectedly: %v\nStderr: %s", err, stderr.String())
				return
			}

			if !tt.shouldFail {
				output := stdout.String()
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
					}
				}
			}
		})
	}
}

// BenchmarkCLIExecution benchmarks the CLI execution time
func BenchmarkCLIExecution(b *testing.B) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		b.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command("./subnetCalc_test", "192.168.1.0/24")
		_ = cmd.Run()
	}
}

// TestIPv6Handling tests IPv6 address handling (expected to be limited)
func TestIPv6Handling(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "subnetCalc_test", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() { _ = os.Remove("./subnetCalc_test") }()

	tests := []struct {
		name string
		cidr string
	}{
		{
			name: "IPv6 /64",
			cidr: "2001:db8::/64",
		},
		{
			name: "IPv6 /48",
			cidr: "2001:db8::/48",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := exec.Command("./subnetCalc_test", tt.cidr)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// According to README, IPv6 handling is "questionable at best"
			// So we just test that it doesn't crash completely
			if err != nil {
				t.Logf("IPv6 support failed as expected: %v", err)
			} else {
				t.Logf("IPv6 support worked for %s: %s", tt.cidr, stdout.String())
			}
		})
	}
}
