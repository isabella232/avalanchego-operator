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
	var network common.Network

	if instance.Status.BootstrapperURL == "" {
		network = *common.NewNetwork(instance.Spec.NodeCount)
	}

	l.Info("Instance Spec: ", "instance.Spec", instance.Spec, "nodeSpecs", instance.Spec.NodeSpecs)

	err = r.ensureConfigMap(req, instance, r.avagoConfigMap(instance, "avago-init-script", common.AvagoBootstraperFinderScript), l)
	if err != nil {
		return ctrl.Result{}, err
	}

	for i, key := range network.KeyPairs {
		err = r.ensureSecret(req, instance, r.avagoSecret(instance, "validator-"+strconv.Itoa(i), key.Cert, key.Key, network.Genesis), l)
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

// func min(a int, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
