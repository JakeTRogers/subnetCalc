package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/JakeTRogers/subnetCalc/formatter"
	"github.com/JakeTRogers/subnetCalc/logger"
	"github.com/JakeTRogers/subnetCalc/subnet"
	"github.com/JakeTRogers/subnetCalc/tui"
)

var version = "v1.1.0"

// rootCmd is the package-level command instance used by Execute().
// Tests should use NewRootCommand() for isolated instances.
var rootCmd *cobra.Command

// NewRootCommand creates and returns a new root command instance.
// This constructor enables test isolation by providing fresh command instances.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "subnetCalc <CIDR>",
		Version:      version,
		Short:        "calculate subnet",
		SilenceUsage: true,
		Long: `subnetCalc is a CLI application to calculate subnets when given an IP address and a subnet mask in CIDR notation. It
will return the requested network, host address range, broadcast address, subnet mask, and the maximum number of hosts.

subnetCalc can also be used to carve up a network into subnets by providing subnet mask size. It then lists them in a
either table or JSON format.

Examples:
  # Get network information for a CIDR:
  subnetCalc 10.12.34.56/19

  # Get network information for a CIDR and carve it up into subnets:
  subnetCalc 10.12.0.0/16 --subnet-size 18

  # Get network information for a CIDR, carve it up into subnets, and print the output in JSON format:
  subnetCalc 192.168.10.0/24 --subnet-size 26 --json

  # Launch interactive TUI for subnet splitting/joining:
  subnetCalc 10.0.0.0/8 --interactive

  # Launch interactive TUI with initial subnet split:
  subnetCalc 192.168.0.0/16 --subnet-size 24 --interactive
`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verboseCount, err := cmd.Flags().GetCount("verbose")
			if err != nil {
				log := logger.GetLogger()
				log.Error().Err(err).Msg("failed to read verbose flag")
				return
			}
			logger.SetLogLevel(verboseCount)
		},
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.GetLogger()

			// If no arguments are provided, print help and exit cleanly.
			if len(args) == 0 {
				return cmd.Help()
			}

			out := cmd.OutOrStdout()

			// Retrieve flag values from command instance (not package-level variables)
			interactive, _ := cmd.Flags().GetBool("interactive")
			subnetMaskBits, _ := cmd.Flags().GetInt("subnet-size")

			log.Debug().Str("cidr", args[0]).Bool("interactive", interactive).Int("subnet_size", subnetMaskBits).Msg("processing input")
			if interactive {
				// Pass subnet-size to TUI if specified, otherwise 0 for no initial split
				initialSplit := 0
				if cmd.Flags().Changed("subnet-size") {
					initialSplit = subnetMaskBits
				}
				log.Warn().Str("cidr", args[0]).Int("initial_split", initialSplit).Msg("entering interactive mode, logging disabled")
				logger.Disable()
				return tui.Run(args[0], initialSplit)
			}

			// Use the subnet package to calculate network details
			n, err := subnet.NewNetwork(args[0])
			if err != nil {
				return fmt.Errorf("parsing network: %w", err)
			}
			log.Info().Str("cidr", n.CIDR.String()).Int("mask_bits", n.MaskBits).Msg("network parsed successfully")

			// if subnet-size flag is set, carve up the supernet into subnets of the requested size
			if cmd.Flags().Changed("subnet-size") {
				if err := n.Split(subnetMaskBits); err != nil {
					return fmt.Errorf("splitting network: %w", err)
				}
				log.Info().Int("target_bits", subnetMaskBits).Int("subnet_count", len(n.Subnets)).Msg("network split completed")
			}

			// print the network details in the requested format
			cfg := formatter.DefaultConfig()
			if cmd.Flags().Changed("json") {
				cfg.Format = formatter.FormatJSON
			} else {
				cfg.Width = terminalWidth(out, cfg.Width)
			}

			f := formatter.New(cfg)
			return printNetworkOutput(out, cfg.Format, f, n)
		},
	}

	// Register flags on the command instance
	cmd.SetVersionTemplate("subnetCalc {{.Version}}\n")
	cmd.Flags().BoolP("json", "j", false, "output information for the requested CIDR in json format")
	cmd.Flags().BoolP("interactive", "i", false, "launch interactive TUI for subnet splitting/joining")
	cmd.Flags().IntP("subnet-size", "s", 0, "number of subnet mask bits to be used in carving up the supernet")
	cmd.PersistentFlags().CountP("verbose", "v", "increase verbosity")
	cmd.MarkFlagsMutuallyExclusive("interactive", "json")

	return cmd
}

// printNetworkOutput writes formatted network information to the provided writer.
// For JSON format, it outputs a single JSON object containing the network and any subnets.
// For other formats, it outputs the network summary followed by a subnet table if present.
func printNetworkOutput(w io.Writer, format formatter.OutputFormat, f formatter.Formatter, n subnet.Network) error {
	// JSON output is represented as a single object (with optional subnets included).
	if format == formatter.FormatJSON {
		output, err := f.FormatSubnets(n)
		if err != nil {
			return fmt.Errorf("formatting network as JSON: %w", err)
		}
		_, err = fmt.Fprintln(w, output)
		return err
	}

	networkInfo, err := f.FormatNetwork(n)
	if err != nil {
		return fmt.Errorf("formatting network info: %w", err)
	}
	if _, err := io.WriteString(w, networkInfo); err != nil {
		return err
	}

	if len(n.Subnets) == 0 {
		return nil
	}

	subnetsOutput, err := f.FormatSubnets(n)
	if err != nil {
		return fmt.Errorf("formatting subnets: %w", err)
	}
	_, err = fmt.Fprintln(w, subnetsOutput)
	return err
}

func terminalWidth(out io.Writer, fallback int) int {
	f, ok := out.(*os.File)
	if !ok {
		return fallback
	}
	fd := int(f.Fd())
	if !term.IsTerminal(fd) {
		return fallback
	}
	w, _, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		return fallback
	}
	return w
}

// Execute runs the root command and returns any error.
// The caller (main.go) is responsible for handling the error and exiting.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd = NewRootCommand()
}
