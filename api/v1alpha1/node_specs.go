package v1alpha1

type NodeSpecs struct {
	// Http port
	// +optional
	// +kubebuilder:default:=9651
	HTTPPort int `json:"httpPort,omitempty"`
}
