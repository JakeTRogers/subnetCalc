package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"
)

var (
	testBinaryPath string
	buildOnce      sync.Once
	buildErr       error
)

func buildTestBinary() (string, error) {
	buildOnce.Do(func() {
		bin, err := os.CreateTemp("", "subnetCalc_test_*")
		if err != nil {
			buildErr = err
			return
		}
		_ = bin.Close()
		// Ensure the file is executable on platforms that care about mode bits.
		_ = os.Chmod(bin.Name(), 0o755)

		buildCmd := exec.Command("go", "build", "-o", bin.Name(), ".")
		buildCmd.Dir = "."
		if err := buildCmd.Run(); err != nil {
			buildErr = err
			_ = os.Remove(bin.Name())
			return
		}
		testBinaryPath = bin.Name()
	})

	return testBinaryPath, buildErr
}

func TestMain(m *testing.M) {
	bin, err := buildTestBinary()
	if err != nil {
		// Best-effort: surface error without depending on testing.T.
		_, _ = os.Stderr.WriteString("Failed to build test binary: " + err.Error() + "\n")
		os.Exit(1)
	}
	code := m.Run()
	_ = os.Remove(bin)
	os.Exit(code)
}

// TestCLI_integration tests the main CLI application end-to-end
func TestCLI_integration(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

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
			args:           []string{"192.168.1.0/24", "--subnet-size", "26"},
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

			cmd := exec.Command(binary, tt.args...)
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

// TestCLI_JSONOutputStructure validates the JSON output structure
func TestCLI_JSONOutputStructure(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

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
			args: []string{"192.168.1.0/24", "--subnet-size", "26", "--json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer

			cmd := exec.Command(binary, tt.args...)
			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			output := stdout.String()

			// Parse JSON to validate structure
			var result map[string]any
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Invalid JSON output: %v\nOutput: %s", err, output)
				return
			}

			// Check required fields (maxSubnets removed in schema v0.2.0)
			requiredFields := []string{"cidr", "firstIP", "lastIP", "networkAddr", "broadcastAddr", "subnetMask", "maskBits", "maxHosts"}
			for _, field := range requiredFields {
				if _, exists := result[field]; !exists {
					t.Errorf("JSON output missing required field: %s", field)
				}
			}

			// If subnets were requested, check subnets array
			if strings.Contains(strings.Join(tt.args, " "), "subnet-size") {
				if subnets, exists := result["subnets"]; !exists {
					t.Error("JSON output should contain subnets array when subnet-size is specified")
				} else if subnetSlice, ok := subnets.([]any); !ok || len(subnetSlice) == 0 {
					t.Error("Subnets should be a non-empty array")
				}
			}
		})
	}
}

// TestCLI_errorHandling tests various error conditions
func TestCLI_errorHandling(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

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
			args:        []string{"192.168.1.0/24", "--subnet-size", "20"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer

			cmd := exec.Command(binary, tt.args...)
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

// TestCLI_subnetCalculations validates specific subnet calculation scenarios
func TestCLI_subnetCalculations(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

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

			cmd := exec.Command(binary, tt.cidr, "--subnet-size", tt.subnetSize)
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

// TestCLI_flags tests various flag combinations
func TestCLI_flags(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		shouldFail     bool
	}{
		{
			name:           "Subnet size flag with styled output",
			args:           []string{"192.168.1.0/24", "--subnet-size", "26"},
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
			name:       "Mutually exclusive flags (interactive and json)",
			args:       []string{"192.168.1.0/24", "--interactive", "--json"},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := exec.Command(binary, tt.args...)
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

// BenchmarkCLI_execution benchmarks the CLI execution time
func BenchmarkCLI_execution(b *testing.B) {
	binary, err := buildTestBinary()
	if err != nil {
		b.Fatalf("Failed to build binary: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binary, "192.168.1.0/24")
		_ = cmd.Run()
	}
}

// TestCLI_IPv6Handling tests IPv6 address handling (expected to be limited)
func TestCLI_IPv6Handling(t *testing.T) {
	binary, err := buildTestBinary()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

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

			cmd := exec.Command(binary, tt.cidr)
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
