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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/go-logr/logr"
)

// AvalanchegoReconciler reconciles a Avalanchego object
type AvalanchegoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=chain.avax.network,resources=avalanchegoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Avalanchego object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *AvalanchegoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Started")
	// your logic here
	instance := &chainv1alpha1.Avalanchego{}
	validatorsKeys := instance.Spec.NodeKeys[:min(instance.Spec.NodeCount, len(instance.Spec.NodeKeys))]

	for i, key := range validatorsKeys {
		l.Info("Key: ", key.Certificate)
		err := r.ensureSecret(req, instance, r.besuSecret(instance, "validator"+strconv.Itoa(i+1), key.Certificate, key.Key), l)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AvalanchegoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chainv1alpha1.Avalanchego{}).
		Complete(r)
}

func (r *AvalanchegoReconciler) besuSecret(instance *chainv1alpha1.Avalanchego,
	name string,
	certificate string,
	key string,
) *corev1.Secret {
	secr := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "avago-" + name + "-key",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "avago-" + name + "-key",
			},
		},
		Type: "Opaque",
		StringData: map[string]string{
			"staker.crt": certificate,
			"staker.key": key,
		},
	}
	controllerutil.SetControllerReference(instance, secr, r.Scheme)
	return secr
}

func (r *AvalanchegoReconciler) ensureSecret(req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.Secret,
	l logr.Logger,
) error {
	found := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		l.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Create(context.TODO(), s)
		if err != nil {
			// Creation failed
			l.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return err
		} else {
			// Creation was successful
			return nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		l.Error(err, "Failed to get Secret")
		return err
	}

	return nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
