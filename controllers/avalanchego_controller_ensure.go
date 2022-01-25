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
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	"github.com/go-logr/logr"
)

const (
	isUpdateable              = true
	isNotUpdateable           = false
	stsUpdateSleepTimeSeconds = 3
	stsUpdateTimeoutSeconds   = 15
)

func (r *AvalanchegoReconciler) ensureConfigMap(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.ConfigMap,
	l logr.Logger,
) error {
	_, err := upsertObject(ctx, r, s, isUpdateable, l)
	return err
}

func (r *AvalanchegoReconciler) ensureSecret(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.Secret,
	l logr.Logger,
) error {
	isSecretUpdateable := isNotUpdateable

	if instance.Spec.Genesis != "" && len(instance.Spec.Certificates) != 0 {
		isSecretUpdateable = isUpdateable
	}
	_, err := upsertObject(ctx, r, s, isSecretUpdateable, l)
	return err
}

func (r *AvalanchegoReconciler) ensureService(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.Service,
	l logr.Logger,
) error {

	_, err := upsertObject(ctx, r, s, isUpdateable, l)
	return err
}

func (r *AvalanchegoReconciler) ensurePVC(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.PersistentVolumeClaim,
	l logr.Logger,
) error {
	// PVC is special in terms of update operation
	// In order to update it, we need to set Spec.VolumeName and Spec.StorageClassName from existing state

	// Searching for existent PVC first
	found := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.GetName(),
		Namespace: s.GetNamespace(),
	}, found)

	if err == nil {
		// Setting up Spec.VolumeName and Spec.StorageClassName values from existent PVC
		s.Spec.VolumeName = found.Spec.VolumeName
		s.Spec.StorageClassName = found.Spec.StorageClassName
	} else if !errors.IsNotFound(err) {
		l.Error(err, "Failed to get existing PVC", s.GetNamespace(), "Type:", s.GetObjectKind().GroupVersionKind().String(), "Name:", s.GetName())
		return err
	}
	_, err = upsertObject(ctx, r, s, isUpdateable, l)
	return err
}

func (r *AvalanchegoReconciler) ensureStatefulSet(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *appsv1.StatefulSet,
	l logr.Logger,
) error {
	existed, err := upsertObject(ctx, r, s, isUpdateable, l)
	if err != nil {
		return err
	}
	if existed {
		for i := 0; i <= (stsUpdateTimeoutSeconds / stsUpdateSleepTimeSeconds); i++ {

			foundPod := &corev1.Pod{}
			err = r.Get(ctx, types.NamespacedName{
				Name:      s.Name + "-0", //Assuming we'll always have 1 pod per StatefulSet
				Namespace: s.ObjectMeta.Namespace,
			}, foundPod)
			if err != nil {
				l.Error(err, "Failed to get Pod")
				return err
			}

			// GenerationLabelName contains a sequence number representing a specific generation of the desired state.
			// We need to ensure that we've got available pods exactly with this generation ID
			podGenerationMatch := false
			if generation, ok := foundPod.Labels[GenerationLabelName]; ok {
				podGenerationMatch = (generation == fmt.Sprint(instance.Generation))
			}

			// Watching if pod is in ready state
			podReady := false
			if foundPod.Status.Phase == corev1.PodRunning {
				podReady = true
			}

			if podGenerationMatch && podReady {
				l.Info("Successfully awaited for StatefulSet", "StatefulSet.Namespace", s.Namespace, "StatefulSet.Name", s.Name)
				return nil
			} else {
				time.Sleep(time.Second * time.Duration(stsUpdateSleepTimeSeconds))
			}
		}
		err = errors.NewTimeoutError("Timed out waiting for StatefulSet to become ready", stsUpdateTimeoutSeconds)
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

	commonLogLabels := []interface{}{"Namespace:", targetObj.GetNamespace(), "Type:", targetObj.GetObjectKind().GroupVersionKind().String(), "Name:", targetObj.GetName()}

	targetObjType := reflect.TypeOf(targetObj).Elem()
	foundObj := reflect.New(targetObjType).Interface().(client.Object)

	err = r.Get(ctx, types.NamespacedName{
		Name:      targetObj.GetName(),
		Namespace: targetObj.GetNamespace(),
	}, foundObj)

	if err == nil && isUpdateable {
		l.Info("Updating existing object", commonLogLabels...)
		// Some of k8s services require ResourceVersion to be specified within update
		targetObj.SetResourceVersion(foundObj.GetResourceVersion())
		if err := r.Update(ctx, targetObj); err != nil {
			// Update failed
			l.Error(err, "Failed to update object", commonLogLabels...)
			return true, err
		} else {
			l.Info("Updated existing object", commonLogLabels...)
			return true, err
		}
	} else if err == nil && !isUpdateable {
		l.Info("Found existing object but it's not updatable", commonLogLabels...)
		return true, err
	} else if !errors.IsNotFound(err) {
		l.Error(err, "Failed to find existing object", commonLogLabels...)
		return true, err
	}

	// Create the Object
	l.Info("Creating a new object", commonLogLabels...)
	if err := r.Create(ctx, targetObj); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new object", commonLogLabels...)
		return false, err
	}
	l.Info("Successfully created a new object", commonLogLabels...)
	// Creation was successful
	return false, nil
}
