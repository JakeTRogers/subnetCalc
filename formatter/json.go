package formatter

import (
	"encoding/json"

	"github.com/JakeTRogers/subnetCalc/logger"
	"github.com/JakeTRogers/subnetCalc/subnet"
)

// JSONFormatter formats network information as JSON.
type JSONFormatter struct {
	Indent bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{Indent: indent}
}

// jsonNetwork is the JSON representation of a network.
type jsonNetwork struct {
	CIDR          string        `json:"cidr"`
	FirstIP       string        `json:"firstIP"`
	LastIP        string        `json:"lastIP"`
	NetworkAddr   string        `json:"networkAddr"`
	BroadcastAddr string        `json:"broadcastAddr"`
	SubnetMask    string        `json:"subnetMask"`
	MaskBits      int           `json:"maskBits"`
	MaxHosts      string        `json:"maxHosts"`
	Subnets       []jsonNetwork `json:"subnets,omitempty"`
}

// toJSONNetwork converts a subnet.Network to jsonNetwork.
func toJSONNetwork(n subnet.Network) jsonNetwork {
	jn := jsonNetwork{
		CIDR:          n.CIDR.String(),
		FirstIP:       n.FirstHostIP.String(),
		LastIP:        n.LastHostIP.String(),
		NetworkAddr:   n.NetworkAddr.String(),
		BroadcastAddr: n.BroadcastAddr.String(),
		SubnetMask:    n.SubnetMask.String(),
		MaskBits:      n.MaskBits,
		MaxHosts:      FormatMaxHosts(n.MaxHosts),
	}

	if len(n.Subnets) > 0 {
		jn.Subnets = make([]jsonNetwork, len(n.Subnets))
		for i, s := range n.Subnets {
			jn.Subnets[i] = toJSONNetwork(s)
		}
	}

	return jn
}

// FormatNetwork formats a single network's information as JSON.
func (f *JSONFormatter) FormatNetwork(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Msg("formatting network as JSON")
	jn := toJSONNetwork(n)
	return f.marshal(jn)
}

// FormatSubnets formats a network with its subnets as JSON.
func (f *JSONFormatter) FormatSubnets(n subnet.Network) (string, error) {
	log := logger.GetLogger()
	log.Trace().Str("cidr", n.CIDR.String()).Int("subnet_count", len(n.Subnets)).Msg("formatting subnets as JSON")
	jn := toJSONNetwork(n)
	return f.marshal(jn)
}

func (f *JSONFormatter) marshal(v any) (string, error) {
	var data []byte
	var err error

	if f.Indent {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Ensure JSONFormatter implements Formatter.
var _ Formatter = (*JSONFormatter)(nil)
