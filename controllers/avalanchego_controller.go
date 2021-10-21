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
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/api/health"
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
	err := r.Get(ctx, req.NamespacedName, instance)
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

	if (instance.Status.BootstrapperURL == "") && (instance.Spec.BootstrapperURL == "") {
		network = *common.NewNetwork(l, instance.Spec.NodeCount)
	}

	if instance.Spec.BootstrapperURL == "" {
		instance.Status.BootstrapperURL = "avago-" + instance.Spec.DeploymentName + "-0-service"
	} else {
		instance.Status.BootstrapperURL = instance.Spec.BootstrapperURL
	}
	err = r.Status().Update(ctx, instance)
	if err != nil {
		l.Error(err, "unable to update instance BootstrapperURL")
	}

	err = r.ensureConfigMap(r.avagoConfigMap(l, instance, "avago-init-script", common.AvagoBootstraperFinderScript), l)
	if err != nil {
		return ctrl.Result{}, err
	}

	for i := 0; i < instance.Spec.NodeCount; i++ {
		nodeName := instance.Spec.DeploymentName + "-" + strconv.Itoa(i)

		if (instance.Spec.BootstrapperURL == "") && (network.Genesis != "") {
			err = r.ensureSecret(l, r.avagoSecret(l, instance, nodeName, network.KeyPairs[i].Cert, network.KeyPairs[i].Key, network.Genesis))
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		err = r.ensureService(l, r.avagoService(l, instance, nodeName))
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensurePVC(l, r.avagoPVC(l, instance, nodeName))
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensureService(l, r.avagoService(l, instance, nodeName))
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.ensureStatefulSet(l, r.avagoStatefulSet(l, instance, nodeName))
		if err != nil {
			return ctrl.Result{}, err
		}

		// add a new service to the NetworkMembersURI if it doesn't exist already
		if _, ok := instance.Status.NetworkMembersURI[nodeName+"-service"]; !ok {
			instance.Status.NetworkMembersURI[nodeName+"-service"] = false
			err = r.Status().Update(ctx, instance)
			if err != nil {
				l.Error(err, "unable to update the status of instance when creating a new NetworkMembersURI")
			}
		}
	}

	// check if a started node has booted successfully
	// once it has booted correctly don't check it again
	for nodeService, booted := range instance.Status.NetworkMembersURI {
		if booted {
			continue
		}

		healthClient := health.NewClient(fmt.Sprintf("http://avago-%s.dev.svc.cluster.local:%d", nodeService, 9650), 20*time.Second)
		healthResp, err := healthClient.Health()
		if err != nil {
			l.Info("error calling health on", "nodeService", nodeService, "err", err)
			// don't update it's status
			continue
		}

		if healthResp.Healthy {
			instance.Status.NetworkMembersURI[nodeService] = true
			err = r.Status().Update(ctx, instance)
			if err != nil {
				l.Error(err, "unable to update the status of instance when updating the state of NetworkMembersURI")
			}
		}
		l.Info("called health - ", "health", healthResp)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AvalanchegoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chainv1alpha1.Avalanchego{}).
		Complete(r)
}

func notContainsS(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return false
		}
	}

	return true
}
