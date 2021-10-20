package client

import "github.com/ava-labs/avalanchego/ids"

// Node defines the AvalancheGo node interface
type Node interface {
	// GetID returns the Node identifier
	// Each node has a unique ID that distinguishes it from
	// other nodes in this network.
	// This is distinct from the Avalanche notion of a node ID.
	// This ID is assigned by the Network; it is not the hash
	// of a staking certificate.
	// We don't use the Avalanche node ID to reference nodes
	// because we may want to start a network where multiple nodes
	// have the same Avalanche node ID.
	GetID() ids.ID
	GetIPAddress() string
}

type NodeConfig struct {
	// The contents of this node's genesis file
	genesis []byte
	// The contents of this node's config file
	configFile []byte

	Name string
}

type NodeDefinition struct {
}

func (n *NodeDefinition) GetIPAddress() string {
	panic("implement me")
}

func (n *NodeDefinition) GetID() ids.ID {
	panic("implement me")
}

func NewNodeDefinition() Node {
	return &NodeDefinition{}
}
