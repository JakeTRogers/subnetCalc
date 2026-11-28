---
status: accepted
---

# Host Addresses follow RFC 3021 and RFC 6164

subnetCalc excludes the Network Address and Broadcast Address from every Network. That gives backwards Host Ranges and a Host Count of 0 for IPv4 `/31` and `/32` and IPv6 `/127`, and reports a Broadcast Address for IPv6, which has none. We decided Host Addresses are RFC-aware instead. Two-address Networks (`/31`, `/127`) have both addresses as hosts (RFC 3021, RFC 6164), single-address Networks (`/32`, `/128`) have one, and IPv6 Networks have no Broadcast Address. This matches how these prefixes are deployed on point-to-point links and loopbacks. Implementation is tracked in #128.

## Considered Options

- **Strict classic**: always exclude the Network Address and Broadcast Address, and report no Host Range for `/31` and `/32`. Rejected because it calls the most common point-to-point prefix unusable.
- **Every address**: Host Count is the total number of addresses. Rejected because it overstates the number of assignable hosts on ordinary IPv4 Networks.

## Consequences

Once implemented, the JSON fields `firstIP`, `lastIP`, `maxHosts` and `broadcastAddr` change for these prefixes and for all IPv6 Networks. Consumers that hard-coded "total minus 2" will see different numbers.
