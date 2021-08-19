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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
)

func (r *AvalanchegoReconciler) avagoSecret(
	instance *chainv1alpha1.Avalanchego,
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
				"app": "avago-" + name,
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

func (r *AvalanchegoReconciler) avagoService(
	instance *chainv1alpha1.Avalanchego,
	name string,
) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "avago-" + name + "-service",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "avago-" + name,
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": "avago-" + name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     9650,
				},
				{
					Name:     "staking",
					Protocol: "TCP",
					Port:     9651,
				},
			},
		},
	}
	controllerutil.SetControllerReference(instance, svc, r.Scheme)
	return svc
}

func (r *AvalanchegoReconciler) avagoStatefulSet(
	instance *chainv1alpha1.Avalanchego,
	name string,
) *appsv1.StatefulSet {

	envVars := r.getEnvVars(instance)
	volumeMounts := r.getVolumeMounts(instance, name)
	volumes := r.getVolumes(instance, name)
	volumeClaim := r.getVolumeClaimTemplate(instance, name)

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "avago-" + name,
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": "avago-" + name,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			// A hack to create a literal *int32 vatiable, set to 1
			Replicas:            &[]int32{1}[0],
			PodManagementPolicy: "OrderedReady",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "avago-" + name,
				},
			},
			ServiceName: "avago-" + name + "-service",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "avago-" + name,
					},
					//TODO Add checksum for cert/key
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "avago",
							Image:           "avaplatform/avalanchego:latest",
							ImagePullPolicy: "IfNotPresent",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
							},
							Env:          envVars,
							VolumeMounts: volumeMounts,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      "TCP",
									ContainerPort: 9650,
								},
								{
									Name:          "staking",
									Protocol:      "TCP",
									ContainerPort: 9651,
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
			VolumeClaimTemplates: volumeClaim,
		},
	}
	controllerutil.SetControllerReference(instance, sts, r.Scheme)
	return sts
}

// func (r *AvalanchegoReconciler) getAvagoInitContainer(instance *chainv1alpha1.Avalanchego) []corev1.Container {
// 	initContainers := []corev1.Container{
// 		corev1.Container{
// 			Name:  "init-bootnode-ip",
// 			Image: "pegasyseng/k8s-helper:v1.18.4",
// 			Command: []string{
// 				"sh",
// 				"-c",
// 				r.getCurlCommand(instance),
// 			},
// 		},
// 	}
// 	return initContainers
// }

func (r *AvalanchegoReconciler) getEnvVars(instance *chainv1alpha1.Avalanchego) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name: "AVAGO_HTTP_HOST",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "AVAGO_PUBLIC_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "AVAGO_NETWORK",
			Value: "local",
		},
		{
			Name:  "AVAGO_STAKING_ENABLED",
			Value: "true",
		},
		{
			Name:  "AVAGO_HTTP_PORT",
			Value: "9650",
		},
		{
			Name:  "AVAGO_STAKING_PORT",
			Value: "9651",
		},
		{
			Name:  "AVAGO_LOG_LEVEL",
			Value: "debug",
		},
		{
			Name:  "AVAGO_STAKING_TLS_CERT_FILE",
			Value: "/etc/avalanchego/st-certs/staker.crt",
		},
		{
			Name:  "AVAGO_STAKING_TLS_KEY_FILE",
			Value: "/etc/avalanchego/st-certs/staker.key",
		},
		{
			Name:  "AVAGO_DB_DIR",
			Value: "/root/.avalanchego",
		},
	}

	return envVars
}

func (r *AvalanchegoReconciler) getVolumeMounts(instance *chainv1alpha1.Avalanchego, name string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "avago-db-" + name,
			MountPath: "/root/.avalanchego",
			ReadOnly:  false,
		},
		{
			Name:      "avago-cert-" + name,
			MountPath: "/etc/avalanchego/st-certs",
			ReadOnly:  true,
		},
	}
	return volumeMounts
}

func (r *AvalanchegoReconciler) getVolumes(instance *chainv1alpha1.Avalanchego, name string) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "avago-db-" + name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "avago-db-" + name,
				},
			},
		},
		{
			Name: "avago-cert-" + name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "avago-" + name + "-key",
				},
			},
		},
	}
	return volumes
}

func (r *AvalanchegoReconciler) getVolumeClaimTemplate(instance *chainv1alpha1.Avalanchego, name string) []corev1.PersistentVolumeClaim {
	pvcs := []corev1.PersistentVolumeClaim{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "avago-db-" + name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
			},
		},
	}
	return pvcs
}
