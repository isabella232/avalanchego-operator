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
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
)

const (
	GenerationLabelName = "operator-generation-id"
)

func (r *AvalanchegoReconciler) avagoConfigMap(
	instance *chainv1alpha1.Avalanchego,
	name string,
	script string,
) *corev1.ConfigMap {
	data := make(map[string]string)
	data["config.sh"] = string(script)
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": avaGoPrefix + name,
			},
		},
		Data: data,
	}
	_ = controllerutil.SetControllerReference(instance, cm, r.Scheme) // TODO should we return this error if non-nil?
	return cm
}

func (r *AvalanchegoReconciler) avagoSecret(
	instance *chainv1alpha1.Avalanchego,
	name string,
	certificate string,
	key string,
	genesis string,
) *corev1.Secret {
	secr := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      avaGoPrefix + name + "-key",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": avaGoPrefix + name,
			},
		},
		Type: "Opaque",
		StringData: map[string]string{
			"staker.crt":   certificate,
			"staker.key":   key,
			"genesis.json": genesis,
		},
	}
	_ = controllerutil.SetControllerReference(instance, secr, r.Scheme) // TODO should we return this error if non-nil?
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
			Name:      avaGoPrefix + name + "-service",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": avaGoPrefix + name,
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": avaGoPrefix + name,
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
	_ = controllerutil.SetControllerReference(instance, svc, r.Scheme) // TODO should we return this error if non-nil?
	return svc
}

func (r *AvalanchegoReconciler) avagoPVC(
	instance *chainv1alpha1.Avalanchego,
	name string,
) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      avaGoPrefix + name + "-pvc",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": avaGoPrefix + name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("50Gi"),
				},
			},
		},
	}
	_ = controllerutil.SetControllerReference(instance, pvc, r.Scheme) // TODO should we return this error if non-nil?
	return pvc
}

func (r *AvalanchegoReconciler) avagoStatefulSet(
	instance *chainv1alpha1.Avalanchego,
	name string,
) *appsv1.StatefulSet {
	var initContainers []corev1.Container
	envVars := r.getEnvVars(instance)
	volumeMounts := r.getVolumeMounts(instance, name)
	volumes := r.getVolumes(instance, name)
	// volumeClaim := r.getVolumeClaimTemplate(instance, name)

	index := name[len(name)-1:]
	if (index == "0") && (instance.Spec.BootstrapperURL == "") {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "AVAGO_BOOTSTRAP_IPS",
			Value: "",
		})
	}
	if (index != "0") || (instance.Spec.BootstrapperURL != "") {
		initContainers = r.getAvagoInitContainer(instance)
		envVars = append(envVars, corev1.EnvVar{
			Name:  "AVAGO_CONFIG_FILE",
			Value: "/etc/avalanchego/conf/conf.json",
		})
	}

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      avaGoPrefix + name,
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"app": avaGoPrefix + name,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			// A hack to create a literal *int32 vatiable, set to 1
			Replicas:            &[]int32{1}[0],
			PodManagementPolicy: "OrderedReady",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": avaGoPrefix + name,
				},
			},
			ServiceName: avaGoPrefix + name + "-service",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":               avaGoPrefix + name,
						GenerationLabelName: fmt.Sprint(instance.Generation),
					},
					//TODO Add checksum for cert/key
				},
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:            "avago",
							Image:           instance.Spec.Image + ":" + instance.Spec.Tag,
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
					ImagePullSecrets: instance.Spec.ImagePullSecrets,
					Volumes:          volumes,
				},
			},
			// VolumeClaimTemplates: volumeClaim,
		},
	}

	if !reflect.DeepEqual(instance.Spec.Resources, corev1.ResourceRequirements{}) {
		sts.Spec.Template.Spec.Containers[0].Resources = instance.Spec.Resources
	}

	_ = controllerutil.SetControllerReference(instance, sts, r.Scheme) // TODO should we return this error if non-nil?
	return sts
}

func (r *AvalanchegoReconciler) getAvagoInitContainer(instance *chainv1alpha1.Avalanchego) []corev1.Container {
	initContainers := []corev1.Container{
		{
			Name:  "init-bootnode-ip",
			Image: "avaplatform/dnsutils:1.0.0",
			Env: []corev1.EnvVar{
				{
					Name:  "CONFIG_PATH",
					Value: "/tmp/conf",
				},
				{
					Name:  "BOOTSTRAPPERS",
					Value: instance.Status.BootstrapperURL,
				},
			},
			Command: []string{
				"sh",
				"-c",
				"/tmp/script/config.sh",
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "avalanchego-init-script",
					MountPath: "/tmp/script",
					ReadOnly:  true,
				},
				{
					Name:      "init-volume",
					MountPath: "/tmp/conf",
					ReadOnly:  false,
				},
			},
		},
	}
	return initContainers
}

func (r *AvalanchegoReconciler) getEnvVars(instance *chainv1alpha1.Avalanchego) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "AVAGO_HTTP_HOST",
			Value: "0.0.0.0",
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
			Name:  "AVAGO_NETWORK_ID",
			Value: "12346",
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
			Name:  "AVAGO_DB_DIR",
			Value: "/root/.avalanchego",
		},
		{
			Name:  "AVAGO_GENESIS",
			Value: "/etc/avalanchego/st-certs/genesis.json",
		},
	}
	//Append genesis and certificates, if it is a new network or cert provided
	if (instance.Spec.BootstrapperURL == "") || (len(instance.Spec.Certificates) > 0) {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "AVAGO_STAKING_TLS_CERT_FILE",
			Value: "/etc/avalanchego/st-certs/staker.crt",
		}, corev1.EnvVar{
			Name:  "AVAGO_STAKING_TLS_KEY_FILE",
			Value: "/etc/avalanchego/st-certs/staker.key",
		})
	}

	for _, v := range instance.Spec.Env {
		envVarsI := indexOf(envVars, v.Name)
		if envVarsI == -1 {
			envVars = append(envVars, v)
		} else {
			envVars[envVarsI].Value = v.Value
		}
	}
	return envVars
}

// Returns -1 if not found, index otherwise
func indexOf(env []corev1.EnvVar, name string) int {
	for i, v := range env {
		if v.Name == name {
			return i
		}
	}
	return -1
}

func (r *AvalanchegoReconciler) getVolumeMounts(instance *chainv1alpha1.Avalanchego, name string) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      avaGoPrefix + "db-" + name,
			MountPath: "/root/.avalanchego",
			ReadOnly:  false,
		},
		{
			Name:      "init-volume",
			MountPath: "/etc/avalanchego/conf",
			ReadOnly:  true,
		},
		{
			Name:      avaGoPrefix + "cert-" + name,
			MountPath: "/etc/avalanchego/st-certs",
			ReadOnly:  true,
		},
	}
}

func (r *AvalanchegoReconciler) getVolumes(instance *chainv1alpha1.Avalanchego, name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: avaGoPrefix + "db-" + name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: avaGoPrefix + name + "-pvc",
				},
			},
		},
		{
			Name: "avalanchego-init-script",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: avaGoPrefix + instance.Spec.DeploymentName + "init-script",
					},
					// A hack to create a literal *int32 vatiable, set to 0777
					DefaultMode: &[]int32{0777}[0],
				},
			},
		},
		{
			Name: "init-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: avaGoPrefix + "cert-" + name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: avaGoPrefix + name + "-key",
				},
			},
		},
	}
}
