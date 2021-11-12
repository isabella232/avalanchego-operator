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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

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
	found := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		// Error that isn't due to the ConfigMap not existing
		l.Error(err, "Failed to get ConfigMap")
		return err
	}
	// Create the ConfigMap
	l.Info("Creating a new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
	if err := r.Create(context.TODO(), s); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
		return err
	}
	// Creation was successful
	return nil
}

func (r *AvalanchegoReconciler) ensureSecret(
	ctx context.Context,
	req ctrl.Request,
	instance *chainv1alpha1.Avalanchego,
	s *corev1.Secret,
	l logr.Logger,
) error {
	found := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		// Error that isn't due to the secret not existing
		l.Error(err, "Failed to get Secret")
		return err
	}
	// Create the secret
	l.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
	if err := r.Create(context.TODO(), s); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		return err
	}
	// Creation was successful
	return nil
}

func (r *AvalanchegoReconciler) ensureService(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.Service,
	l logr.Logger,
) error {
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		// Error that isn't due to the secret not existing
		l.Error(err, "Failed to get Service")
		return err
	}
	// Create the service
	l.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
	if err := r.Create(context.TODO(), s); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		return err
	}
	// Creation was successful
	return nil
}

func (r *AvalanchegoReconciler) ensurePVC(
	ctx context.Context,
	req ctrl.Request,
	s *corev1.PersistentVolumeClaim,
	l logr.Logger,
) error {
	found := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		// Error that isn't due to the secret not existing
		l.Error(err, "Failed to get PVC")
		return err
	}
	// Create the service
	l.Info("Creating a new PVC", "PersistentVolumeClaim.Namespace", s.Namespace, "PersistentVolumeClaim.Name", s.Name)
	if err := r.Create(context.TODO(), s); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new PVC", "PersistentVolumeClaim.Namespace", s.Namespace, "PersistentVolumeClaim.Name", s.Name)
		return err
	}
	// Creation was successful
	return nil
}

func (r *AvalanchegoReconciler) ensureStatefulSet(
	ctx context.Context,
	req ctrl.Request,
	s *appsv1.StatefulSet,
	l logr.Logger,
) error {
	found := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      s.ObjectMeta.Name,
		Namespace: s.ObjectMeta.Namespace,
	}, found)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		// Error that isn't due to the secret not existing
		l.Error(err, "Failed to get StatefulSet")
		return err
	}
	// Create the StatefulSet
	l.Info("Creating a new StatefulSet", "StatefulSet.Namespace", s.Namespace, "StatefulSet.Name", s.Name)
	if err := r.Create(context.TODO(), s); err != nil {
		// Creation failed
		l.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", s.Namespace, "StatefulSet.Name", s.Name)
		return err
	}
	// Creation was successful
	return nil
}
