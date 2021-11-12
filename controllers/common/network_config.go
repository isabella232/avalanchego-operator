/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/staking"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/hashing"
)

type Network struct {
	Genesis  string
	KeyPairs []KeyPair
}

type KeyPair struct {
	Cert string
	Key  string
	Id   string
}

func NewNetwork(networkSize int) (Network, error) {
	var (
		g Genesis
		n Network
	)
	if err := json.Unmarshal([]byte(defaultGenesisConfigJSON), &g); err != nil {
		return Network{}, fmt.Errorf("couldn't unmarshal local genesis: %w", err)
	}
	for i := 0; i < networkSize; i++ {
		// TODO handle the below error
		stakingKeyCertPair, _ := newStakingKeyCertPair()
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Cert)
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Key)
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Id)
		n.KeyPairs = append(n.KeyPairs, stakingKeyCertPair)
		g.InitialStakers = append(g.InitialStakers, InitialStaker{NodeID: stakingKeyCertPair.Id, RewardAddress: g.Allocations[1].AvaxAddr, DelegationFee: 5000})
	}
	genesisBytes, err := json.Marshal(g)
	if err != nil {
		panic("Error: cannot marshal genesis.json, common package is invalid")
	}
	n.Genesis = string(genesisBytes)

	fmt.Print("------------------------------------------")
	fmt.Print(n.Genesis)

	return n, nil
}

func newStakingKeyCertPair() (KeyPair, error) {
	certBytes, keyBytes, err := staking.NewCertAndKeyBytes()
	if err != nil {
		return KeyPair{}, err
	}
	nodeID, err := ids.ToShortID(hashing.PubkeyBytesToAddress(certBytes))
	if err != nil {
		return KeyPair{}, fmt.Errorf("problem deriving node ID from certificate: %w", err)
	}
	nodeIDStr := nodeID.PrefixedString(constants.NodeIDPrefix)
	return KeyPair{
		Cert: string(certBytes),
		Key:  string(keyBytes),
		Id:   nodeIDStr,
	}, nil
}
