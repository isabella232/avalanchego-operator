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

// These structs are the analogs of those in https://github.com/ava-labs/avalanchego/blob/master/genesis/config.go
// Except these have string fields where the structs in the linked file have ids.ShortID
type Genesis struct {
	NetworkID                  int             `json:"networkID"`
	Allocations                []Allocation    `json:"allocations"`
	StartTime                  int             `json:"startTime"`
	InitialStakeDuration       int             `json:"initialStakeDuration"`
	InitialStakeDurationOffset int             `json:"initialStakeDurationOffset"`
	InitialStakedFunds         []string        `json:"initialStakedFunds"`
	InitialStakers             []InitialStaker `json:"initialStakers"`
	CChainGenesis              string          `json:"cChainGenesis"`
	Message                    string          `json:"message"`
}

type Allocation struct {
	EthAddr        string        `json:"ethAddr"`
	AvaxAddr       string        `json:"avaxAddr"`
	InitialAmount  int           `json:"initialAmount"`
	UnlockSchedule []UnlockSched `json:"unlockSchedule"`
}

type UnlockSched struct {
	Amount   int `json:"amount"`
	Locktime int `json:"locktime,omitempty"`
}

type InitialStaker struct {
	NodeID        string `json:"nodeID"`
	RewardAddress string `json:"rewardAddress"`
	DelegationFee int    `json:"delegationFee"`
}
