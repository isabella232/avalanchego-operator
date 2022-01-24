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
	"reflect"
	"time"

	//"time"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/go-logr/logr"
)

func (r *AvalanchegoReconciler) ensureConfigMap(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.ConfigMap,
	l logr.Logger,
) error {
	_, err := upsertObject(ctx, r, s, true, l)
	return err
}

func (r *AvalanchegoReconciler) ensureSecret(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.Secret,
	l logr.Logger,
) error {
	_, err := upsertObject(ctx, r, s, false, l)
	return err
}

func (r *AvalanchegoReconciler) ensureService(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.Service,
	l logr.Logger,
) error {
	_, err := upsertObject(ctx, r, s, true, l)
	return err
}

func (r *AvalanchegoReconciler) ensurePVC(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.PersistentVolumeClaim,
	l logr.Logger,
) error {
	_, err := upsertObject(ctx, r, s, false, l)
	return err
}

func (r *AvalanchegoReconciler) ensureStatefulSet(
	ctx context.Context,
	req ctrl.Request,
	s *appsv1.StatefulSet,
	l logr.Logger,
) error {
	existed, err := upsertObject(ctx, r, s, true, l)
	if err != nil {
		return err
	}
	if existed {
		sleepTimeSeconds := 3
		timeoutSeconds := 10
		for i := 0; i <= (timeoutSeconds / sleepTimeSeconds); i++ {
			time.Sleep(time.Second * time.Duration(sleepTimeSeconds))
			found := &appsv1.StatefulSet{}
			err = r.Get(ctx, types.NamespacedName{
				Name:      s.ObjectMeta.Name,
				Namespace: s.ObjectMeta.Namespace,
			}, found)
			if err != nil {
				l.Error(err, "Failed to get StatefulSet")
				return err
			}
			if found.Status.ReadyReplicas == *s.Spec.Replicas {
				l.Info("Successfully updated existing StatefulSet", "StatefulSet.Namespace", s.Namespace, "StatefulSet.Name", s.Name)
				return nil
			}
		}
		err = errors.NewTimeoutError("Timed out waiting for StatefulSet to become ready", timeoutSeconds)
		l.Error(err, "StatefulSet.Namespace"+s.Namespace+"StatefulSet.Name"+s.Name)
		return err
	}
	// Create or Update was successful
	return nil
}

// Creates or updates k8s object if it already exists.
// isUpdateable must be true to allow update (some objects like PVC are immutable)
func upsertObject(
	ctx context.Context,
	r *AvalanchegoReconciler,
	targetObj client.Object,
	isUpdateable bool,
	l logr.Logger) (existed bool, err error) {

	commonLogInfo := []interface{}{"Namespace:", targetObj.GetNamespace(), "Type:", targetObj.GetObjectKind().GroupVersionKind().String(), "Name:", targetObj.GetName()}

	targetObjType := reflect.TypeOf(targetObj).Elem()
	foundObj := reflect.New(targetObjType).Interface().(client.Object)

	err = r.Get(ctx, types.NamespacedName{
		Name:      targetObj.GetName(),
		Namespace: targetObj.GetNamespace(),
	}, foundObj)

	if err == nil && isUpdateable {
		l.Info("Updating existing object", commonLogInfo...)
		// Some of k8s services require ResourceVersion to be specified within update
		targetObj.SetResourceVersion(foundObj.GetResourceVersion())
		if err := r.Update(ctx, targetObj); err != nil {
			// Update failed
			l.Error(err, "Failed to update object", commonLogInfo...)
			return true, err
		} else {
			l.Info("Updated existing object", commonLogInfo...)
			return true, err
		}
	} else if err == nil && !isUpdateable {
		l.Info("Found existing object but it's not updatable", commonLogInfo...)
		return true, err
	} else if !errors.IsNotFound(err) {
		l.Error(err, "Failed to find existing object", commonLogInfo...)
		return true, err
	}

	// Create the Object
	l.Info("Creating a new object", commonLogInfo...)
	if err := r.Create(ctx, targetObj); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new object", commonLogInfo...)
		return existed, err
	}
	l.Info("Successfully created a new StatefulSet", commonLogInfo...)
	// Creation was successful
	return false, nil
}
