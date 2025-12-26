// Package subnet provides interfaces for subnet calculation operations.
package subnet

// Calculator defines the interface for subnet calculation operations.
type Calculator interface {
	// Calculate creates a Network from a CIDR string.
	Calculate(cidr string) (Network, error)

	// Split divides a network into subnets of the specified prefix length.
	Split(network *Network, targetBits int) error
}

// DefaultCalculator is the standard implementation of Calculator.
type DefaultCalculator struct{}

// NewCalculator creates a new DefaultCalculator.
func NewCalculator() *DefaultCalculator {
	return &DefaultCalculator{}
}

// Calculate creates a Network from a CIDR string.
func (c *DefaultCalculator) Calculate(cidr string) (Network, error) {
	return NewNetwork(cidr)
}

// Split divides a network into subnets of the specified prefix length.
func (c *DefaultCalculator) Split(network *Network, targetBits int) error {
	return network.Split(targetBits)
}

// Ensure DefaultCalculator implements Calculator.
var _ Calculator = (*DefaultCalculator)(nil)
