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

const (
	AvagoBootstraperFinderScript = `#!/bin/bash

echo "Config path: $CONFIG_PATH/conf.json"
echo "DNSs to resolve: $BOOTSTRAPPERS"

IFS=',' read -r -a bootstrappers_array <<< "$BOOTSTRAPPERS"

delim=""
joined_ip=""

for bootstrapper in "${bootstrappers_array[@]}"
do
		retry=3
		dig_out=''
		echo "--------------------------"

		while [ -z "$dig_out" ] && [ "$retry" -ne "0" ]
		do
			if [ "$retry" -ne "3" ]; then
				sleep 10
			fi
			echo "Resolving $bootstrapper"
			dig_out=$(dig +search +short "$bootstrapper")
			retry=$((retry-1))
		done

		echo "List of IPs to add:"
		IFS=$'\n' read -r -d '' -a ips <<< "$dig_out"
		for ip in "${ips[@]}"
		do
			echo "$ip"
			joined_ip="$joined_ip$delim$ip"
			delim=","
		done
		echo "--------------------------"

done

if [ -z "$joined_ip" ]; then
	echo "ERROR no DNS adresses have been resolved"
	exit 1
fi

echo "Final json: {\"bootstrap-ips\":\"${joined_ip}:9651\"}"
touch "$CONFIG_PATH/conf.json"
echo "{\"bootstrap-ips\":\"${joined_ip}:9651\"}" > "$CONFIG_PATH/conf.json"
ls $CONFIG_PATH
echo "Guts of $CONFIG_PATH/conf.json"
cat $CONFIG_PATH/conf.json
`
)

// PrivateKey-vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE => P-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh
// PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN => X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p
// 56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027 => 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC

var defaultGenesisConfigJSON = `{
	"networkID": 12346,
	"allocations": [
		{
			"ethAddr": "0xb3d82b1367d362de99ab59a658165aff520cbd4d",
			"avaxAddr": "X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh",
			"initialAmount": 0,
			"unlockSchedule": [
				{
					"amount": 10000000000000000,
					"locktime": 1633824000
				}
			]
		},
		{
			"ethAddr": "0xb3d82b1367d362de99ab59a658165aff520cbd4d",
			"avaxAddr": "X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p",
			"initialAmount": 300000000000000000,
			"unlockSchedule": [
				{
					"amount": 20000000000000000
				},
				{
					"amount": 10000000000000000,
					"locktime": 1633824000
				}
			]
		},
		{
			"ethAddr": "0xb3d82b1367d362de99ab59a658165aff520cbd4d",
			"avaxAddr": "X-custom1ur873jhz9qnaqv5qthk5sn3e8nj3e0kmzpjrhp",
			"initialAmount": 10000000000000000,
			"unlockSchedule": [
				{
					"amount": 10000000000000000,
					"locktime": 1633824000
				}
			]
		}
	],
	"startTime": 1630987200,
	"initialStakeDuration": 31536000,
	"initialStakeDurationOffset": 5400,
	"initialStakedFunds": [
		"X-custom1g65uqn6t77p656w64023nh8nd9updzmxwd59gh"
	],
	"initialStakers": [],
	"cChainGenesis": "{\"config\":{\"chainId\":43112,\"homesteadBlock\":0,\"daoForkBlock\":0,\"daoForkSupport\":true,\"eip150Block\":0,\"eip150Hash\":\"0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0\",\"eip155Block\":0,\"eip158Block\":0,\"byzantiumBlock\":0,\"constantinopleBlock\":0,\"petersburgBlock\":0,\"istanbulBlock\":0,\"muirGlacierBlock\":0,\"apricotPhase1BlockTimestamp\":0,\"apricotPhase2BlockTimestamp\":0},\"nonce\":\"0x0\",\"timestamp\":\"0x0\",\"extraData\":\"0x00\",\"gasLimit\":\"0x5f5e100\",\"difficulty\":\"0x0\",\"mixHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\",\"coinbase\":\"0x0000000000000000000000000000000000000000\",\"alloc\":{\"8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC\":{\"balance\":\"0x295BE96E64066972000000\"}},\"number\":\"0x0\",\"gasUsed\":\"0x0\",\"parentHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}",
	"message": "Make time for fun"
}`
