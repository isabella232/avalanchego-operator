# Avalanchego Operator
Propose of this operator is to ease the automation of test processes. Operator provides an abstraction layer, that allows programmatically create private Avalanche chains at any time and scale.
## Creating Validators and custom nodes
Avalanchego Operator extends kubernetes API with new kind of resources, `Avalanchego`. To new create a network from scratch, apply this object.

```
apiVersion: chain.avax.network/v1alpha1
kind: Avalanchego
metadata:
  name: avalanchego-test-validator
spec:
  # Add fields here
  deploymentName: test-validator
  nodeCount: 5
  image: avaplatform/avalanchego
  tag: v1.6.0
  env:
  - name: AVAGO_LOG_LEVEL
    value: debug
  resources:
    limits:
      cpu: "1"
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 1Gi
```

`name` this name is a K8s's reference to the chain validators, use this name to check the status or delete the chain

`deploymentName` is a suffix for downstream k8s objects (pods, services, secrets, etc.)

`nodeCount` initial number of validators, these nodes will be added to genesis.json as initial stakers

`image` and `tag` docker image and tag

`env` common configuration for chain nodes, check the full list here: https://github.com/ava-labs/avalanchego/blob/master/config/keys.go

`resources` amount of CPU and RAM, an individual node would be able to use (https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

