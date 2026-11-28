# subnetCalc

Calculates the properties of IP address blocks and carves them into smaller blocks.

## Language

### Blocks

**Network**:
A contiguous IP address block identified by a network address and a Prefix Length (e.g. `10.12.32.0/19`).
_Avoid_: block, range, CIDR (as a noun for the block itself)

**Subnet**:
A Network in its role as a part of a larger Network.
_Avoid_: child network, sub-network

**Supernet**:
A Network in its role as the container of smaller Subnets.
_Avoid_: parent network, aggregate

**Prefix Length**:
The number of leading bits that identify a Network, written as `/N`.
_Avoid_: mask bits, subnet size, CIDR bits

### Addresses

**Network Address**:
The first address of a Network, which identifies the Network itself.
_Avoid_: base address, network ID

**Broadcast Address**:
The last address of an IPv4 Network, reserved for reaching every host on it. IPv6 Networks have none.
_Avoid_: broadcast (for IPv6), last address

**Host Address**:
An address in a Network that can be assigned to a host. Normally excludes the Network Address and Broadcast Address. In a two-address Network (IPv4 `/31`, IPv6 `/127`) both addresses are Host Addresses, and a single-address Network (`/32`, `/128`) is its one Host Address.
_Avoid_: assignable address, usable IP

**Host Range**:
The first through last Host Address of a Network.
_Avoid_: host address range, assignable range, usable range

**Host Count**:
The number of Host Addresses in a Network.
_Avoid_: maximum hosts, max hosts, hosts, usable hosts

### Operations

**Split**:
Halving a Network into two Subnets whose Prefix Length is one greater.
_Avoid_: using it for dividing to a target Prefix Length in one step

**Divide**:
Carving a Network into all Subnets of a target Prefix Length in one step.
_Avoid_: split (for this operation), subnetting

**Join**:
Merging two sibling Subnets back into their Supernet; the inverse of Split.
_Avoid_: merge, collapse, summarize
