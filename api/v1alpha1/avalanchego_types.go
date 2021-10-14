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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AvalanchegoSpec defines the desired state of Avalanchego
type AvalanchegoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Number of nodes to create. All the nodes will be created as validators
	// +optional
	// +kubebuilder:default:=5
	NodeCount int `json:"nodeCount,omitempty"`

	// Docker image name. Will be used in chain deployments
	// +optional
	// +kubebuilder:default:="avaplatform/avalanchego"
	Image string `json:"image,omitempty"`

	// Docker image tag. Will be used in chain deployments
	// +optional
	// +kubebuilder:default:="latest"
	Tag string `json:"tag,omitempty"`

	// Node specifications. If used will ignore NodeCount
	// +optional
	// +kubebuilder:default:="9661"
	NodeSpecs int `json:"nodeSpecs,omitempty"`
}

// AvalanchegoStatus defines the observed state of Avalanchego
type AvalanchegoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service URL of the Bootstrapper node
	BootstrapperURL   string   `json:"bootstrapperURL"`
	NetworkMembersURI []string `json:"networkMembersURI"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Avalanchego is the Schema for the avalanchegoes API
type Avalanchego struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AvalanchegoSpec   `json:"spec,omitempty"`
	Status AvalanchegoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AvalanchegoList contains a list of Avalanchego
type AvalanchegoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Avalanchego `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Avalanchego{}, &AvalanchegoList{})
}
