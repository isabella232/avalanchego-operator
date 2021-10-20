package client

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Network is an abstraction of an Avalanche network
type Network interface {
	// Stop all the nodes in the network
	Stop() error
	// Ready Returns a chan that is closed when the network is ready to be used for the first time
	Ready() chan IsReady
	// AddNode Start a new node with the config
	AddNode(NodeConfig) (Node, error)
	// RemoveNode stops/removes the node with this ID.
	RemoveNode(name string) error
	// GetNode returns the node with this ID.
	GetNode(ids.ID) (Node, error)
}

// NewNetwork returns a new network whose initial state is specified in the config
func NewNetwork(conf *NetworkConfig) Network {
	logrus.Infof("Creating a new k8s network: %s...", conf.Name)
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	if err != nil {
		logrus.Fatal(err)
	}

	validatorObj := &v1alpha1.Avalanchego{
		ObjectMeta: metav1.ObjectMeta{Namespace: conf.Namespace, Name: conf.Name},
		Spec: v1alpha1.AvalanchegoSpec{
			NodeCount: 3,
			Image:     "avaplatform/avalanchego",
			Tag:       "v1.6.2",
		},
	}

	err = kubeClient.Create(context.Background(), validatorObj)
	if err != nil {
		logrus.Fatal(err)
	}

	return &NetworkDefinition{
		validatorNetwork: validatorObj,
		kubeClient:       kubeClient,
		ctx:              context.Background(),
		nameSpace:        conf.Namespace,
		nodes:            map[string]*v1alpha1.Avalanchego{},
		clientset:        clientset,
		networkName:      conf.Name,
	}
}

type IsReady struct {
	Ready bool
	Err   error
}

type NetworkDefinition struct {
	ctx              context.Context
	validatorNetwork *v1alpha1.Avalanchego
	kubeClient       client.Client
	nameSpace        string
	nodes            map[string]*v1alpha1.Avalanchego
	clientset        *kubernetes.Clientset
	networkName      string
}

func (n *NetworkDefinition) Stop() error {
	for nodeName := range n.nodes {
		err := n.RemoveNode(nodeName)
		if err != nil {
			return err
		}
	}
	n.nodes = map[string]*v1alpha1.Avalanchego{}

	logrus.Info("Removing network...")
	err := n.kubeClient.Delete(n.ctx, n.validatorNetwork)
	if err != nil {
		return err
	}

	n.validatorNetwork = nil
	return nil
}

// Ready returns a IsReady channel and starts a function to check if the network is ready for use
func (n *NetworkDefinition) Ready() chan IsReady {
	readyChan := make(chan IsReady)
	go func() {

		deployedObjs := []*v1alpha1.Avalanchego{n.validatorNetwork}
		for _, nodeObj := range n.nodes {
			deployedObjs = append(deployedObjs, nodeObj)
		}

		var pollStartTime time.Time
		for pollStartTime = time.Now(); time.Since(pollStartTime) < 5*time.Minute; time.Sleep(5 * time.Second) {

			areObjectsDeployed := false
			for _, deployedObj := range deployedObjs {
				// check if the object has been created to kubernetes
				currentObj := &v1alpha1.Avalanchego{}
				err := n.kubeClient.Get(context.Background(), client.ObjectKey{Namespace: n.nameSpace, Name: deployedObj.Name}, currentObj)
				if err != nil {
					readyChan <- IsReady{Err: err}
					return
				}

				// check if the operator has deployed the object instances
				if len(currentObj.Status.NetworkMembersURI) != deployedObj.Spec.NodeCount {
					logrus.Infof("Operator has deployed %d/%d instances after %s... ",
						len(currentObj.Status.NetworkMembersURI),
						deployedObj.Spec.NodeCount,
						time.Since(pollStartTime),
					)
					areObjectsDeployed = false
					break
				}
				areObjectsDeployed = true
			}

			// if some object have not been deployed don't check the pods
			if !areObjectsDeployed {
				logrus.Infof("Network not ready after %s", time.Since(pollStartTime))
				continue
			}

			// check if the deployed pods are up and running
			pods, err := n.clientset.CoreV1().Pods(n.nameSpace).List(n.ctx, metav1.ListOptions{})
			if err != nil {
				readyChan <- IsReady{Err: err}
				return
			}
			logrus.Infof("There are %d pods in the cluster", len(pods.Items))

			ready := false
			for _, pod := range pods.Items {
				v1pod, err := n.clientset.CoreV1().Pods(pod.Namespace).Get(n.ctx, pod.Name, metav1.GetOptions{})
				if err != nil {
					if statusError, isStatus := err.(*errors.StatusError); isStatus {
						logrus.Infof("Error getting pod %v", statusError.ErrStatus.Message)
						ready = false
						break
					}
					readyChan <- IsReady{Err: err}
					return
				}
				logrus.Infof("Found %s with status %s", v1pod.Name, v1pod.Status.Phase)
				if v1pod.Status.Phase != v1.PodRunning {
					ready = false
					break
				}
				ready = true
				time.Sleep(time.Second)
			}

			if ready {
				logrus.Info("it's ready now")
				readyChan <- IsReady{Ready: true}
				return
			}

			logrus.Infof("Network not ready after %s", time.Since(pollStartTime))
		}
		readyChan <- IsReady{Err: fmt.Errorf("Network not ready after %s", time.Since(pollStartTime))}
	}()
	return readyChan
}

func (n *NetworkDefinition) AddNode(conf NodeConfig) (Node, error) {
	logrus.Infof("Adding node %s...", conf.Name)
	if _, ok := n.nodes[conf.Name]; ok {
		return nil, fmt.Errorf("node already added")
	}
	workerObj := &v1alpha1.Avalanchego{
		ObjectMeta: metav1.ObjectMeta{Namespace: n.nameSpace, Name: conf.Name},
		Spec: v1alpha1.AvalanchegoSpec{
			NodeCount:       1,
			Image:           "avaplatform/avalanchego",
			Tag:             "v1.6.0",
			BootstrapperURL: "avago-test-validator-0-service",
			DeploymentName:  "test-worker",
		},
	}

	err := n.kubeClient.Create(n.ctx, workerObj)
	if err != nil {
		logrus.Fatal(err)
	}

	n.nodes[conf.Name] = workerObj

	return NewNodeDefinition(), nil
}

func (n *NetworkDefinition) RemoveNode(name string) error {
	logrus.Infof("Removing node %s...", name)
	node, ok := n.nodes[name]
	if !ok {
		return fmt.Errorf("no node registered with %s", name)
	}

	err := n.kubeClient.Delete(n.ctx, node)
	if err != nil {
		return err
	}

	delete(n.nodes, name)
	return nil
}

func (n *NetworkDefinition) GetNode(ids.ID) (Node, error) {
	panic("implement me")
}

// NetworkConfig holds the configuration of the network
type NetworkConfig struct {
	Namespace string
	Name      string
}
