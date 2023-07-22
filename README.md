# subnetCalc

## Description

This is a golang-based [Cobra](https://github.com/spf13/cobra) CLI utility to calculate subnets when given an IP address and a subnet mask in CIDR notation. It will return the host address range, network address, broadcast address, subnet mask, maximum number of subnets, and maximum number of hosts. It works as expected with IPv4, however IPv6 handling is questionable at best.

## Usage

`subnetCalc <ip address>/<subnet mask>`

## Examples

### Get Network Information for a /19 Network

`subnetCalc 10.12.34.56/19`

```text
               Network: 10.12.32.0/19
    Host Address Range: 10.12.32.1 - 10.12.63.254
     Broadcast Address: 10.12.63.255
           Subnet Mask: 255.255.224.0
       Maximum Subnets: 2,048
         Maximum Hosts: 8,190
```

### List /27 Subnets Contained in a /25 Network

`subnetCalc 192.168.10.0/25 --subnet_size 27`

```text
               Network: 192.168.10.0/25
    Host Address Range: 192.168.10.1 - 192.168.10.126
     Broadcast Address: 192.168.10.127
           Subnet Mask: 255.255.255.128
       Maximum Subnets: 2
         Maximum Hosts: 126

  192.168.10.0/25 contains 4 /27 subnets:
╭───┬──────────────────┬───────────────┬────────────────┬────────────────┬───────╮
│ # │ SUBNET           │ FIRST IP      │ LAST IP        │ BROADCAST      │ HOSTS │
├───┼──────────────────┼───────────────┼────────────────┼────────────────┼───────┤
│ 1 │ 192.168.10.0/27  │ 192.168.10.1  │ 192.168.10.30  │ 192.168.10.31  │ 30    │
│ 2 │ 192.168.10.32/27 │ 192.168.10.33 │ 192.168.10.62  │ 192.168.10.63  │ 30    │
│ 3 │ 192.168.10.64/27 │ 192.168.10.65 │ 192.168.10.94  │ 192.168.10.95  │ 30    │
│ 4 │ 192.168.10.96/27 │ 192.168.10.97 │ 192.168.10.126 │ 192.168.10.127 │ 30    │
╰───┴──────────────────┴───────────────┴────────────────┴────────────────┴───────╯
```

### List /20 Subnets Contained in a /19 Network in JSON Format

`subnetCalc 10.12.34.56/19 --subnet_size 20 --json`

```json
{
  "cidr": "10.12.32.0/19",
  "firstIP": "10.12.32.1",
  "lastIP": "10.12.63.254",
  "networkAddr": "10.12.32.0",
  "broadcastAddr": "10.12.63.255",
  "subnetMask": "255.255.224.0",
  "maskBits": 19,
  "subnetBits": 11,
  "maxSubnets": 2048,
  "maxHosts": 8190,
  "subnets": [
    {
      "cidr": "10.12.32.0/20",
      "firstIP": "10.12.32.1",
      "lastIP": "10.12.47.254",
      "networkAddr": "10.12.32.0",
      "broadcastAddr": "10.12.47.255",
      "subnetMask": "255.255.240.0",
      "maskBits": 20,
      "subnetBits": 12,
      "maxSubnets": 4096,
      "maxHosts": 4094
    },
    {
      "cidr": "10.12.48.0/20",
      "firstIP": "10.12.48.1",
      "lastIP": "10.12.63.254",
      "networkAddr": "10.12.48.0",
      "broadcastAddr": "10.12.63.255",
      "subnetMask": "255.255.240.0",
      "maskBits": 20,
      "subnetBits": 12,
      "maxSubnets": 4096,
      "maxHosts": 4094
    }
  ]
}
```

## Getting Started

To get started using `subnetCalc`, put the binary into your preferred OS's `$PATH` and run it from the command line.

## Feedback

Bug reports, feature requests, and pull requests are welcome but may not be responded to in an even remotely timely manner.
