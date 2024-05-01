/*
Copyright © 2023 Jake Rogers <code@supportoss.org>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"net/netip"
	"os"

	"github.com/JakeTRogers/subnetCalc/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// flipBytes performs a bitwise XOR on each byte in the slice.
// returns a slice of bytes with the bits flipped.
func flipBytes(b []byte) []byte {
	for i := 0; i < len(b); i++ {
		b[i] ^= 0xFF
	}
	return b
}

type network struct {
	CIDR          netip.Prefix `json:"cidr"`
	FirstHostIP   netip.Addr   `json:"firstIP"`
	LastHostIP    netip.Addr   `json:"lastIP"`
	NetworkAddr   netip.Addr   `json:"networkAddr"`
	BroadcastAddr netip.Addr   `json:"broadcastAddr"`
	SubnetMask    netip.Addr   `json:"subnetMask"`
	MaskBits      int          `json:"maskBits"`
	SubnetBits    int          `json:"subnetBits"`
	MaxSubnets    uint         `json:"maxSubnets"`
	MaxHosts      uint         `json:"maxHosts"`
	MaskSize      int          `json:"-"`
	Subnets       []network    `json:"subnets,omitempty"`
}

// getBroadcastAddr calculates the broadcast address for a subnet by ORing the network address and the inverted subnet mask.
// returns the broadcast address as a netip.Addr.
func (n network) getBroadcastAddr() netip.Addr {
	invertedMask := flipBytes(n.SubnetMask.AsSlice())
	var lastIPBytes = make([]byte, len(n.NetworkAddr.AsSlice()))

	for i := 0; i < len(n.NetworkAddr.AsSlice()); i++ {
		lastIPBytes[i] = n.NetworkAddr.AsSlice()[i] | invertedMask[i]
	}
	b, _ := netip.AddrFromSlice(lastIPBytes)
	return b
}

// getSubnetBits calculates the available subnet bits for a given network address and mask bits based on the network class.
// returns an integer representing the number of subnet bits.
func (n network) getSubnetBits() int {
	firstOctet := n.NetworkAddr.AsSlice()[0]
	switch {
	case firstOctet < 128:
		return n.MaskBits - 8
	case firstOctet < 192:
		return n.MaskBits - 16
	case firstOctet < 224:
		return n.MaskBits - 24
	case firstOctet < 240:
		return n.MaskBits - 32
	default:
		return n.MaskBits - 40
	}
}

// getSubnetMask calculates the subnet mask given the number of mask bits and the mask size.
// returns the subnet mask as a netip.Addr.
func (n network) getSubnetMask() netip.Addr {
	var maskBytes = make([]byte, n.MaskSize/8)
	for i := 0; i < len(maskBytes); i++ {
		for j := 0; j < 8; j++ {
			if n.MaskBits > 0 {
				maskBytes[i] |= 1 << uint(7-j)
				n.MaskBits--
			}
		}
	}
	subnetMask, _ := netip.AddrFromSlice(maskBytes)
	return subnetMask
}

// getSubnets calculates the number of subnets that will fit in a supernet using the provided subnet mask bits.
// returns a slice of network structs contained in a supernet.
func (n *network) getSubnets(subnetMaskBits int) {
	// get the number of subnets of size 'subnetMaskBits' that will fit in the supernet
	numSubnets := int(math.Pow(2, float64(subnetMaskBits-n.MaskBits)))

	for i := 0; i < numSubnets; i++ {
		if i == 0 {
			n.Subnets = append(n.Subnets, getNetworkDetails(fmt.Sprintf("%s/%d", n.NetworkAddr, subnetMaskBits)))
		} else {
			n.Subnets = append(n.Subnets, getNetworkDetails(fmt.Sprintf("%s/%d", n.Subnets[i-1].BroadcastAddr.Next(), subnetMaskBits)))
		}
	}
}

// printNetwork prints information about an IP network to stdout.
func (n network) printNetwork() {
	// Use the message package to format large numbers with commas
	p := message.NewPrinter(language.English)

	fmt.Println()
	fmt.Println("               Network:", n.CIDR)
	fmt.Println("    Host Address Range:", n.FirstHostIP, "-", n.LastHostIP)
	fmt.Println("     Broadcast Address:", n.BroadcastAddr)
	fmt.Println("           Subnet Mask:", n.SubnetMask)
	p.Println("       Maximum Subnets:", n.MaxSubnets)
	p.Println("         Maximum Hosts:", n.MaxHosts)
}

// printJSON will print a network struct in json format.
func (n network) printNetworkJSON() {
	netJSON, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		utils.Log.Fatal().Msg(err.Error())
	}
	fmt.Println(string(netJSON))
}

// printSubnets uses the table package to print subnet information in a table.
func (n network) printSubnets(color bool) {
	p := message.NewPrinter(language.English)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if color {
		t.SetStyle(table.StyleColoredBlackOnBlueWhite)
	} else {
		t.SetStyle(table.StyleRounded)
	}
	t.AppendHeader(table.Row{"#", "SUBNET", "FIRST IP", "LAST IP", "BROADCAST", "HOSTS"})

	for i, s := range n.Subnets {
		t.AppendRow([]interface{}{i + 1, s.CIDR, s.FirstHostIP, s.LastHostIP, s.BroadcastAddr, p.Sprint(s.MaxHosts)})
	}

	fmt.Printf("\n  %v contains %d /%d subnets:\n", n.CIDR, len(n.Subnets), n.Subnets[0].MaskBits)
	t.Render()
}

// getNetworkDetails takes a CIDR and returns a network struct with details about the network
// returns a network struct containing network details.
func getNetworkDetails(cidr string) network {
	var n network
	var err error

	// use netip package to confirm the provided input is a valid ipv4 or ipv6 CIDR
	inputCIDR, err := netip.ParsePrefix(cidr)
	if err != nil {
		utils.Log.Fatal().Msg(err.Error())
	}

	n.CIDR = netip.MustParsePrefix(fmt.Sprintf("%s/%d", inputCIDR.Masked().Addr(), inputCIDR.Bits()))
	n.NetworkAddr = n.CIDR.Masked().Addr()
	n.MaskBits = n.CIDR.Bits()
	n.MaskSize = n.CIDR.Addr().BitLen()
	n.SubnetMask = n.getSubnetMask()
	n.BroadcastAddr = n.getBroadcastAddr()
	n.FirstHostIP = n.NetworkAddr.Next()
	n.LastHostIP = n.BroadcastAddr.Prev()
	n.SubnetBits = n.getSubnetBits()
	n.MaxSubnets = uint(math.Pow(2, float64(n.SubnetBits)))
	n.MaxHosts = 1<<(n.MaskSize-n.MaskBits) - 2
	return n
}

var color bool
var subnetMaskBits int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "subnetCalc <CIDR>",
	Version: "v0.1.5",
	Short:   "calculate subnet",
	Long: `subnetCalc is a CLI application to calculate subnets when given an IP address and a subnet mask in CIDR notation. It
will return the requested network, host address range, broadcast address, subnet mask, maximum number of subnets, and
the maximum number hosts.

subnetCalc can also be used to carve up a network into subnets by providing subnet mask size. It then lists them in a
either table or JSON format.

Examples:
  # Get network information for a CIDR:
  subnetCalc 10.12.34.56/19

  # Get network information for a CIDR and carve it up into subnets:
  subnetCalc 10.12.0.0/16 --subnet_size 18

  # Get network information for a CIDR, carve it up into subnets, and print the output in JSON format:
  subnetCalc 192.168.10.0/24 --subnet_size 26 --json
`,

	PersistentPreRun: utils.SetLogLevel,
	Run: func(cmd *cobra.Command, args []string) {
		// if no arguments are provided, print help
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				utils.Log.Fatal().Msg(err.Error())
			}
			os.Exit(0)
		} else if len(args) > 1 {
			utils.Log.Fatal().Msg("too many arguments, expected CIDR notation")
		}

		// populate network struct with details of the provided CIDR
		n := getNetworkDetails(args[0])

		// if subnet_size flag is set, carve up the supernet into subnets of the requested size
		if cmd.Flags().Changed("subnet_size") {
			// check if subnet mask bits are larger than the supernet's mask bits
			if subnetMaskBits <= n.MaskBits {
				utils.Log.Fatal().Msgf("subnet mask bits, %d, must be larger than the supernet's mask bits: %d", subnetMaskBits, n.MaskBits)
			}
			// populate n.subnets with a slice of network structs containing subnet details
			n.getSubnets(subnetMaskBits)
		}

		// print the network details in the requested format
		if cmd.Flags().Changed("json") {
			n.printNetworkJSON()
		} else {
			n.printNetwork()
			if n.Subnets != nil {
				n.printSubnets(color)
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		utils.Log.Fatal().Msg(err.Error())
	}
}

func init() {
	rootCmd.SetVersionTemplate("subnetCalc {{.Version}}\n")
	rootCmd.Flags().BoolVarP(&color, "color", "c", false, "output subnet table in color")
	rootCmd.Flags().BoolP("json", "j", false, "output information for the requested CIDR in json format")
	rootCmd.MarkFlagsMutuallyExclusive("color", "json")
	rootCmd.Flags().IntVarP(&subnetMaskBits, "subnet_size", "s", 0, "number of subnet mask bits to be used in carving up the supernet")
	rootCmd.PersistentFlags().CountP("verbose", "v", "increase verbosity")
}
