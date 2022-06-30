package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtinkClusterSpec defines the desired state of VirtinkCluster
type VirtinkClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ControlPlaneEndpoint capiv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`
}

// VirtinkClusterStatus defines the observed state of VirtinkCluster
type VirtinkClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.controlPlaneEndpoint.host`
//+kubebuilder:printcolumn:name="Port",type=integer,JSONPath=`.spec.controlPlaneEndpoint.port`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`

// VirtinkCluster is the Schema for the virtinkclusters API
type VirtinkCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtinkClusterSpec   `json:"spec,omitempty"`
	Status VirtinkClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtinkClusterList contains a list of VirtinkCluster
type VirtinkClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtinkCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtinkCluster{}, &VirtinkClusterList{})
}
