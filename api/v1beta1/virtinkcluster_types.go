package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtinkClusterSpec defines the desired state of VirtinkCluster
type VirtinkClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	ControlPlaneEndpoint capiv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// ControlPlaneServiceTemplate can be used to modify service that fronts the control plane nodes to handle the
	// api-server traffic (port 6443). This field is optional, by default control plane nodes will use a service
	// of type ClusterIP, which will make workload cluster only accessible within the same cluster. Note, this does
	// not aim to expose the entire service spec to users, but only provides capability to modify the service metadata
	// and the service type.
	ControlPlaneServiceTemplate ControlPlaneServiceTemplate `json:"controlPlaneServiceTemplate,omitempty"`

	// InfraClusterSecretRef is a reference to a secret with a kubeconfig for external cluster used for infra.
	InfraClusterSecretRef *corev1.ObjectReference `json:"infraClusterSecretRef,omitempty"`
}

// ControlPlaneServiceTemplate describes the template for the control plane service.
type ControlPlaneServiceTemplate struct {
	// Service metadata allows to set labels and annotations for the service.
	// This field is optional.
	// +kubebuilder:pruning:PreserveUnknownFields
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	// Type can be used to modify type of service that fronts the control plane nodes to handle the
	// api-server traffic (port 6443). This field is optional, by default control plane nodes will use a service
	// of type ClusterIP, which will make workload cluster only accessible within the same cluster.
	Type *corev1.ServiceType `json:"type,omitempty"`
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
