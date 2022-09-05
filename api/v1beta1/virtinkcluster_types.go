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

	// ControlPlaneServiceType can be used to modify type of service that fronts the control plane nodes to handle the
	// api-server traffic (port 6443). This field is optional, by default control plane nodes will use a service
	// of type ClusterIP, which will make workload cluster only accessible within the same cluster.
	ControlPlaneServiceType *corev1.ServiceType `json:"controlPlaneServiceType,omitempty"`

	// InfraClusterSecretRef is a reference to a secret with a kubeconfig for external cluster used for infra.
	InfraClusterSecretRef *corev1.ObjectReference `json:"infraClusterSecretRef,omitempty"`
	NodeAddressConfig     *NodeAddressConfig      `json:"nodeAddressConfig,omitempty"`
}

type NodeAddressConfig struct {
	// Addresses is list of IP addresses for allocating to nested cluster nodes.
	Addresses []string `json:"addresses,omitempty"`

	// Annotations are CNI required annotations to specify static IP and MAC address for pod.
	// can use $IP_ADDRESS as a placeholder for IP address, provider will replace it by allocated IP address.
	// can use $MAC_ADDRESS as a placeholder for MAC address, provider will replace it by a self generated MAC address.
	// eg: ["cni.projectcalico.org/ipAddrs=[\"$IP_ADDRESS\"]", "cni.projectcalico.org/hwAddr=$MAC_ADDRESS"]
	Annotations []string `json:"annotations,omitempty"`
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
