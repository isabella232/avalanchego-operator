package client

import (
	"testing"

	"gotest.tools/assert"
)

func TestConnection(t *testing.T) {

	conf := NetworkConfig{
		Namespace: "dev",
		Name:      "unit-test",
	}

	network := NewNetwork(&conf)
	isReady := <-network.Ready()

	assert.NilError(t, isReady.Err)
	assert.Equal(t, isReady.Ready, true)

	_, err := network.AddNode(NodeConfig{Name: "derpy"})
	assert.NilError(t, err)

	isReady = <-network.Ready()
	assert.NilError(t, isReady.Err)
	assert.Equal(t, isReady.Ready, true)

	_, err = network.AddNode(NodeConfig{Name: "derpy2"})
	assert.NilError(t, err)

	isReady = <-network.Ready()
	assert.NilError(t, isReady.Err)
	assert.Equal(t, isReady.Ready, true)

	//err = network.Stop()
	//assert.NilError(t, err)
	//node, err := network.AddNode(NodeConfig{})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//node, err = network.AddNode(NodeConfig{})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//logrus.Info(node)
}
