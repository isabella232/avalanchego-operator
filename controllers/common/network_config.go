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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
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

func NewNetwork(networkSize int) *Network {
	var n Network
	g := Genesis{}
	json.Unmarshal([]byte(localGenesisConfigJSON), &g)
	for i := 0; i < networkSize; i++ {
		cert, key, id, _ := newCertKeyIdString()
		fmt.Print("------------------------------------------")
		fmt.Print(cert)
		fmt.Print("------------------------------------------")
		fmt.Print(key)
		fmt.Print("------------------------------------------")
		fmt.Print(id)
		n.KeyPairs = append(n.KeyPairs, KeyPair{Cert: cert, Key: key, Id: id})
		g.InitialStakers = append(g.InitialStakers, InitialStaker{NodeID: id, RewardAddress: g.Allocations[1].AvaxAddr, DelegationFee: 5000})
	}
	data, _ := json.Marshal(g)
	n.Genesis = string(data)

	return &n
}

func newCertKeyIdString() (string, string, string, error) {
	// Create key to sign cert with
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", "", fmt.Errorf("couldn't generate rsa key: %w", err)
	}

	// Create self-signed staking cert
	certTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(0),
		NotBefore:             time.Date(2020, time.January, 0, 0, 0, 0, 0, time.UTC),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	if err != nil {
		return "", "", "", fmt.Errorf("couldn't create certificate: %w", err)
	}

	var certBuff bytes.Buffer
	if err := pem.Encode(&certBuff, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return "", "", "", fmt.Errorf("couldn't write cert file: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", "", "", fmt.Errorf("couldn't marshal private key: %w", err)
	}

	var keyBuff bytes.Buffer
	if err := pem.Encode(&keyBuff, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return "", "", "", fmt.Errorf("couldn't write private key: %w", err)
	}

	id, err := ids.ToShortID(hashing.PubkeyBytesToAddress(certBytes))
	if err != nil {
		return "", "", "", fmt.Errorf("problem deriving node ID from certificate: %w", err)
	}
	fullId := id.PrefixedString(constants.NodeIDPrefix)

	return certBuff.String(), keyBuff.String(), fullId, nil
}
