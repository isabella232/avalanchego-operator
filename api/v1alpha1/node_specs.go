package v1alpha1

type NodeSpecs struct {
	// Http port
	// +optional
	// +kubebuilder:default:=9651
	HTTPPort int `json:"httpPort,omitempty"`

	NodeName            string `json:"nodeName,omitempty"`
	IsValidator         bool   `json:"isValidator,omitempty"`
	Genesis             string `json:"genesis,omitempty"`
	Cert                string `json:"cert,omitempty"`
	CertKey             string `json:"certKey,omitempty"`
	BootStrapperURL     string `json:"bootStrapperURL,omitempty"`
	IsStartingValidator bool   `json:"isStartingValidator,omitempty"`
}
