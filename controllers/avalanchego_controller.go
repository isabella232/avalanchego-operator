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
	"encoding/base64"
	"fmt"
	"reflect"
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

const (
	createStsAsync asyncCreateStatefulSet = true
	createStsSync  asyncCreateStatefulSet = false
)

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

	//Pre flight checks
	//Number of certs should match nodeCount
	if len(instance.Spec.Certificates) > 0 && len(instance.Spec.Certificates) != instance.Spec.NodeCount {
		err = errors.NewBadRequest("Number of provided certificate does not match nodeCount")
		instance.Status.Error = err.Error()
		if err := r.Status().Update(ctx, instance); err != nil {
			l.Error(err, "error calling Update")
		}
		return ctrl.Result{}, err
	}

	// Ignore these environment variables if given
	for i, v := range instance.Spec.Env {
		switch v.Name {
		case "AVAGO_PUBLIC_IP", "AVAGO_HTTP_HOST", "AVAGO_STAKING_TLS_CERT_FILE", "AVAGO_STAKING_TLS_KEY_FILE", "AVAGO_DB_DIR", "AVAGO_HTTP_PORT", "AVAGO_STAKING_PORT", "AVAGO_GENESIS":
			instance.Spec.Env[i] = instance.Spec.Env[len(instance.Spec.Env)-1]
			instance.Spec.Env = instance.Spec.Env[:len(instance.Spec.Env)-1]
		}
	}

	var network common.Network
	if (instance.Status.BootstrapperURL == "") && (instance.Spec.BootstrapperURL == "") && (instance.Spec.Genesis == "") {
		l.Info("Making new network")
		var err error
		network, err = common.NewNetwork(instance.Spec.NodeCount)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("couldn't make new network")
		}
	}

	if instance.Spec.BootstrapperURL == "" {
		instance.Status.BootstrapperURL = avaGoPrefix + instance.Spec.DeploymentName + "-0-service"
		if network.Genesis != "" {
			instance.Status.Genesis = network.Genesis
		}
	} else {
		instance.Status.BootstrapperURL = instance.Spec.BootstrapperURL
		instance.Status.Genesis = instance.Spec.Genesis
	}

	// NetworkMembersURI is mandatory field, if NetworkMembersURI has not been previously set up, then making it as empty struct
	if reflect.ValueOf(instance.Status.NetworkMembersURI).IsZero() {
		instance.Status.NetworkMembersURI = make([]string, 0)
	}

	if err := r.Status().Update(ctx, instance); err != nil {
		l.Error(err, "Failed to update instance status")
	}

	if err := r.ensureConfigMap(
		ctx,
		req,
		instance,
		r.avagoConfigMap(instance, avaGoPrefix+instance.Spec.DeploymentName+"init-script", common.AvagoBootstraperFinderScript),
		l,
	); err != nil {
		return ctrl.Result{}, err
	}

	for i := 0; i < instance.Spec.NodeCount; i++ {
		switch {
		case (instance.Spec.BootstrapperURL == "") && (network.Genesis != ""):
			if err := r.ensureSecret(
				ctx,
				req,
				instance,
				r.avagoSecret(
					instance,
					instance.Spec.DeploymentName+"-"+strconv.Itoa(i),
					network.KeyPairs[i].Cert,
					network.KeyPairs[i].Key,
					network.Genesis,
				),
				l,
			); err != nil {
				return ctrl.Result{}, err
			}
		case (instance.Spec.Genesis != "") && (len(instance.Spec.Certificates) > 0):
			bytes, err := base64.StdEncoding.DecodeString(instance.Spec.Certificates[i].Cert)
			if err != nil {
				instance.Status.Error = err.Error()
				if err := r.Status().Update(ctx, instance); err != nil {
					l.Error(err, "error calling Update")
				}
				return ctrl.Result{}, err
			}
			tempCert := string(bytes)
			bytes, err = base64.StdEncoding.DecodeString(instance.Spec.Certificates[i].Key)
			if err != nil {
				instance.Status.Error = err.Error()
				if err := r.Status().Update(ctx, instance); err != nil {
					l.Error(err, "error calling Update")
				}
				return ctrl.Result{}, err
			}
			tempKey := string(bytes)
			if err := r.ensureSecret(
				ctx,
				req,
				instance,
				r.avagoSecret(
					instance,
					instance.Spec.DeploymentName+"-"+strconv.Itoa(i),
					tempCert,
					tempKey,
					instance.Spec.Genesis,
				),
				l,
			); err != nil {
				return ctrl.Result{}, err
			}
		default:
			if err := r.ensureSecret(
				ctx,
				req,
				instance,
				r.avagoSecret(
					instance,
					instance.Spec.DeploymentName+"-"+strconv.Itoa(i),
					"",
					"",
					instance.Spec.Genesis,
				), l,
			); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Running ensureStatefulSet in a separate loop
	// Otherwise ensureSecret will create secret with an empty certificate
	for i := 0; i < instance.Spec.NodeCount; i++ {
		serviceName := instance.Spec.DeploymentName + "-" + strconv.Itoa(i)
		networkMemberUriName := avaGoPrefix + serviceName + "-service"

		if err := r.ensureService(
			ctx,
			req,
			r.avagoService(instance, serviceName),
			l,
		); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.ensurePVC(
			ctx,
			req,
			r.avagoPVC(instance, instance.Spec.DeploymentName+"-"+strconv.Itoa(i)),
			l,
		); err != nil {
			return ctrl.Result{}, err
		}

		var async asyncCreateStatefulSet

		// In case of brand new network - create all statefulSets asynchronously
		if !reflect.ValueOf(network).IsZero() {
			async = createStsAsync
		} else {
			async = createStsSync
		}

		if err := r.ensureStatefulSet(
			ctx,
			req,
			instance,
			r.avagoStatefulSet(instance, instance.Spec.DeploymentName+"-"+strconv.Itoa(i)),
			l,
			async,
		); err != nil {
			instance.Status.Error = err.Error()
			if err := r.Status().Update(ctx, instance); err != nil {
				l.Error(err, "error calling ensureStatefulSet error status update")
			}
			return ctrl.Result{}, err
		} else if notContainsS(instance.Status.NetworkMembersURI, networkMemberUriName) {
			instance.Status.NetworkMembersURI = append(instance.Status.NetworkMembersURI, networkMemberUriName)
			if err := r.Status().Update(ctx, instance); err != nil {
				l.Error(err, "error calling NetworkMembersURI status update")
			}
		}
	}
	// Assuming that all the above operations are now finished successfully, clearing the error status
	instance.Status.Error = ""
	if err := r.Status().Update(ctx, instance); err != nil {
		l.Error(err, "error cleating error status update")
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
