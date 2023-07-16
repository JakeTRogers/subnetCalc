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
	"github.com/spf13/cobra"
)

// flipBytes performs a bitwise XOR on each byte in the slice.
func flipBytes(b []byte) []byte {
	for i := 0; i < len(b); i++ {
		b[i] ^= 0xFF
	}
	return b
}

// getSubnetMask returns the subnet mask given the number of mask bits and the mask size.
func getSubnetMask(maskBits, maskSize int) netip.Addr {
	var maskBytes = make([]byte, maskSize/8)
	for i := 0; i < len(maskBytes); i++ {
		for j := 0; j < 8; j++ {
			if maskBits > 0 {
				maskBytes[i] |= 1 << uint(7-j)
				maskBits--
			}
		}
	}
	subnetMask, _ := netip.AddrFromSlice(maskBytes)
	return subnetMask
}

// getSubnetBits returns the available subnet bits for a given network address and mask bits.
func getSubnetBits(networkAddr []byte, maskBits int) int {
	firstOctet := networkAddr[0]
	if firstOctet < 128 {
		return maskBits - 8
	} else if firstOctet < 192 {
		return maskBits - 16
	} else if firstOctet < 224 {
		return maskBits - 24
	} else if firstOctet < 240 {
		return maskBits - 32
	} else {
		return maskBits - 40
	}
}

// lastIP returns the last IP address in a subnet given an IP address and a subnet mask.
func lastIP(ip []byte, mask []byte) netip.Addr {
	invertedMask := flipBytes(mask)
	var lastBytes = make([]byte, len(ip))
	for i := 0; i < len(ip); i++ {
		lastBytes[i] = ip[i] | invertedMask[i]
	}
	lastIP, _ := netip.AddrFromSlice(lastBytes)
	return lastIP
}

// printNetworkInformation will print a network struct in a human readable format.
func printNetworkInformation(n network) {
	fmt.Printf("\n            IP Address: %s\n", n.IpAddr)
	fmt.Printf("           Subnet Mask: %s\n\n", n.SubnetMask)
	fmt.Println("    Host Address Range:", n.FirstHostIP, "-", n.LastHostIP)
	fmt.Println("       Network Address:", n.NetworkAddr)
	fmt.Println("     Broadcast Address:", n.BroadcastAddr)
	fmt.Println("       Maximum Subnets:", n.MaxSubnets)
	fmt.Println("      Hosts Per Subnet:", n.HostsPerSubnet)
}

type network struct {
	BroadcastAddr  netip.Addr   `json:"broadcast_addr"`
	CIDR           netip.Prefix `json:"cidr"`
	FirstHostIP    netip.Addr   `json:"first_ip"`
	HostsPerSubnet int          `json:"hosts_per_subnet"`
	IpAddr         netip.Addr   `json:"ip_addr"`
	LastHostIP     netip.Addr   `json:"last_ip"`
	MaskBits       int          `json:"mask_bits"`
	MaskSize       int          `json:"mask_size"`
	MaxSubnets     int          `json:"max_subnets"`
	NetworkAddr    netip.Addr   `json:"network_addr"`
	SubnetBits     int          `json:"subnet_bits"`
	SubnetMask     netip.Addr   `json:"subnet_mask"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "subnetCalc",
	Version: "v0.0.1",
	Short:   "calculate subnet",
	Long: `subnetCalc is a CLI application to calculate subnets when given an IP address and a subnet mask in CIDR notation. It
will return the requested IP, subnet mask, host address range, network address, broadcast address, subnet bits, mask
bits, mask size, maximum number of subnets, max hosts per subnet.

	example:
	    > subnetCalc 10.12.34.56/19

                    IP Address: 10.12.34.56
                   Subnet Mask: 255.255.224.0

            Host Address Range: 10.12.32.1 - 10.12.63.254
               Network Address: 10.12.32.0
             Broadcast Address: 10.12.63.255
               Maximum Subnets: 2048
              Hosts Per Subnet: 8190`,

	PersistentPreRun: utils.SetLogLevel,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				utils.Log.Fatal().Msg(err.Error())
			}
			os.Exit(0)
		} else if len(args) > 1 {
			utils.Log.Fatal().Msg("too many arguments, expected CIDR notation")
		}

		var (
			n   network
			err error
		)

		// use netip package to confirm the provided input is a valid ipv4 or ipv6 CIDR
		inputCIDR, err := netip.ParsePrefix(args[0])
		if err != nil {
			utils.Log.Fatal().Msg(err.Error())
		}
		utils.Log.Debug().Msgf("input cidr: %s", inputCIDR)

		// use netip package to extract IP and mask information from user provided CIDR
		n.CIDR = netip.MustParsePrefix(fmt.Sprintf("%s/%d", inputCIDR.Masked().Addr(), inputCIDR.Bits()))
		n.IpAddr = inputCIDR.Addr()
		n.MaskBits = n.CIDR.Bits()
		n.MaskSize = n.IpAddr.BitLen()
		utils.Log.Debug().Msgf("ip: %s, mask bits: %d, mask size: %d", n.IpAddr, n.MaskBits, n.MaskSize)

		// calculate subnet mask from mask bits
		n.SubnetMask = getSubnetMask(n.MaskBits, n.MaskSize)
		utils.Log.Debug().Msgf("subnet mask: %s", n.SubnetMask)

		// get network address using IP and mask
		n.NetworkAddr = n.CIDR.Masked().Addr()

		// calculate the broadcast address using the network address and mask
		n.BroadcastAddr = lastIP(n.NetworkAddr.AsSlice(), n.SubnetMask.AsSlice())
		utils.Log.Debug().Msgf("Network Address: %s, Broadcast Address: %s", n.NetworkAddr, n.BroadcastAddr)

		// calculate the host address range using the network address and broadcast address
		n.FirstHostIP = n.NetworkAddr.Next()
		n.LastHostIP = n.BroadcastAddr.Prev()
		utils.Log.Debug().Msgf("ip range: %s - %s", n.FirstHostIP, n.LastHostIP)

		// calculate available subnet bits based on network address and mask bits
		n.SubnetBits = getSubnetBits(n.NetworkAddr.AsSlice(), n.MaskBits)
		utils.Log.Debug().Msgf("subnet bits %d", n.SubnetBits)

		// calculate maximum number of subnets
		n.MaxSubnets = int(math.Pow(2, float64(n.SubnetBits)))
		utils.Log.Debug().Msgf("max subnets: %d", n.MaxSubnets)

		// calculate hosts per subnet
		n.HostsPerSubnet = 1<<(n.MaskSize-n.MaskBits) - 2
		utils.Log.Debug().Msgf("hosts per subnet: %d", n.HostsPerSubnet)

		if cmd.Flags().Changed("json") {
			netJSON, err := json.MarshalIndent(n, "", "  ")
			if err != nil {
				utils.Log.Fatal().Msg(err.Error())
			}
			fmt.Print(string(netJSON))
		} else {
			printNetworkInformation(n)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		utils.Log.Fatal().Msg(err.Error())
	}
}

func init() {
	rootCmd.SetVersionTemplate("subnetCalc {{.Version}}\n")
	rootCmd.Flags().BoolP("json", "j", false, "output information for the requested CIDR in json format")
	rootCmd.PersistentFlags().CountP("verbose", "v", "increase verbosity")
}
