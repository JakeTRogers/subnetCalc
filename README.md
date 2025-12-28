# subnetCalc

## Description

This is a Go CLI utility (built with [Cobra](https://github.com/spf13/cobra)) to calculate network information from a CIDR.

It can also split a “supernet” into smaller subnets and render the result either as a styled table or JSON. There is also an optional interactive TUI for splitting/joining subnets.

IPv4 is the primary target. IPv6 works for basic calculations/output, but deeper functionality (especially around very large splits) is intentionally limited.

## Usage

`subnetCalc <ip address>/<subnet mask>`

### Flags

- `--subnet-size`, `-s` — split the input network into subnets of this prefix length
- `--json`, `-j` — output JSON
- `--interactive`, `-i` — launch the interactive TUI (mutually exclusive with `--json`)
- `--verbose`, `-v` — increase verbosity (repeat for more)

## Examples

### List /12 Subnets Contained in a /8 Network Using the Interactive TUI

![svg](/assets/demo_interactive-mode.svg)

### Get Network Information for a /19 Network

`subnetCalc 10.12.34.56/19`

```text
               Network: 10.12.32.0/19
    Host Address Range: 10.12.32.1 - 10.12.63.254
     Broadcast Address: 10.12.63.255
           Subnet Mask: 255.255.224.0
         Maximum Hosts: 8,190
```

### List /27 Subnets Contained in a /25 Network

`subnetCalc 192.168.10.0/25 --subnet-size 27`

```text
               Network: 192.168.10.0/25
    Host Address Range: 192.168.10.1 - 192.168.10.126
     Broadcast Address: 192.168.10.127
           Subnet Mask: 255.255.255.128
         Maximum Hosts: 126

  192.168.10.0/25 contains 4 /27 subnets:
╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
│ #   Subnet              Subnet Mask     Assignable Range              Broadcast       Hosts      │
│1   192.168.10.0/27     255.255.255.224 192.168.10.1 - 192.168.10.30  192.168.10.31   30          │
│2   192.168.10.32/27    255.255.255.224 192.168.10.33 - 192.168.10.62 192.168.10.63   30          │
│3   192.168.10.64/27    255.255.255.224 192.168.10.65 - 192.168.10.94 192.168.10.95   30          │
│4   192.168.10.96/27    255.255.255.224 192.168.10.97 - 192.168.10.126192.168.10.127  30          │
╰──────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### List /20 Subnets Contained in a /19 Network in JSON Format

`subnetCalc 10.12.34.56/19 --subnet-size 20 --json`

```json
{
  "cidr": "10.12.32.0/19",
  "firstIP": "10.12.32.1",
  "lastIP": "10.12.63.254",
  "networkAddr": "10.12.32.0",
  "broadcastAddr": "10.12.63.255",
  "subnetMask": "255.255.224.0",
  "maskBits": 19,
  "maxHosts": "8,190",
  "subnets": [
    {
      "cidr": "10.12.32.0/20",
      "firstIP": "10.12.32.1",
      "lastIP": "10.12.47.254",
      "networkAddr": "10.12.32.0",
      "broadcastAddr": "10.12.47.255",
      "subnetMask": "255.255.240.0",
      "maskBits": 20,
      "maxHosts": "4,094"
    },
    {
      "cidr": "10.12.48.0/20",
      "firstIP": "10.12.48.1",
      "lastIP": "10.12.63.254",
      "networkAddr": "10.12.48.0",
      "broadcastAddr": "10.12.63.255",
      "subnetMask": "255.255.240.0",
      "maskBits": 20,
      "maxHosts": "4,094"
    }
  ]
}
```

## Notes / Limitations

- `--subnet-size` will refuse to generate an extremely large number of subnets (currently capped at 1,000,000) to avoid accidental OOM/hangs.
- IPv6 output is supported, but some concepts (like “broadcast”) are displayed as the last address in the range.

## Feedback

Bug reports, feature requests, and pull requests are welcome but may not be responded to in an even remotely timely manner.
