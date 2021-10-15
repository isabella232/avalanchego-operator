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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/ava-labs/avalanchego-operator/controllers/common"
)

// AvalanchegoReconciler reconciles a Avalanchego object
type AvalanchegoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Avalanchego object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *AvalanchegoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Started")
	// Fetch the Avalanchego instance
	instance := &chainv1alpha1.Avalanchego{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("Not found so maybe deleted")
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	err = r.ensureConfigMap(req, instance, r.avagoConfigMap(instance, "avago-init-script", common.AvagoBootstraperFinderScript), l)
	if err != nil {
		l.Info("err on ensureConfigMap", "err", err)
		return ctrl.Result{}, err
	}

	instance.Spec.NodeSpecs = generateNodeSpecs(l, instance.Spec.NodeCount)
	l.Info("Instance Spec: ", "nodeSpecs", instance.Spec)

	for i, node := range instance.Spec.NodeSpecs {
		err = r.ensureSecret(req, instance, r.avagoSecret(instance, node, i), l)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensureService(req, instance, r.avagoService(instance, "validator-"+strconv.Itoa(i)), l)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensurePVC(req, instance, r.avagoPVC(instance, "validator-"+strconv.Itoa(i)), l)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensureService(req, instance, r.avagoService(instance, "validator-"+strconv.Itoa(i)), l)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensureStatefulSet(req, instance, r.avagoStatefulSet(instance, "validator-"+strconv.Itoa(i)), l)
		if err != nil {
			return ctrl.Result{}, err
		}
		if i == 0 {
			instance.Status.BootstrapperURL = "avago-validator-0-service"
			r.Status().Update(ctx, instance)
		}
		instance.Status.NetworkMembersURI = append(instance.Status.NetworkMembersURI, "avago-validator-"+strconv.Itoa(i)+"-service")
		r.Status().Update(ctx, instance)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AvalanchegoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chainv1alpha1.Avalanchego{}).
		Complete(r)
}

func generateNodeSpecs(l logr.Logger, nodeCount int) []chainv1alpha1.NodeSpecs {
	// todo only generate the network once
	l.Info("creating a new network")
	network := *common.NewNetwork(5)

	nodeSpecs := make([]chainv1alpha1.NodeSpecs, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodeSpecs[i] = chainv1alpha1.NodeSpecs{
			HTTPPort: 9658,
			NodeName: fmt.Sprintf("avago-node-%d", i),
			Genesis: network.Genesis,
		}
	}

	// first five are validators - nodes cannot be removed
	for i := 0; i < nodeCount && i < 5; i++ {
		nodeSpecs[i].IsValidator = true
		nodeSpecs[i].NodeName = fmt.Sprintf("avago-validator-%d", i)
		nodeSpecs[i].Cert = network.KeyPairs[i].Cert
		nodeSpecs[i].CertKey = network.KeyPairs[i].Key
	}



	return nodeSpecs
}
