package v1beta1

import (
	kubridv1alpha1 "github.com/smartxworks/kubrid/pkg/apis/kubrid/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubridMachineSpec defines the desired state of KubridMachine
type KubridMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProviderID *string `json:"providerID,omitempty"`

	VMSpec kubridv1alpha1.VirtualMachineSpec `json:"vmSpec"`
}

// KubridMachineStatus defines the observed state of KubridMachine
type KubridMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderID",type=string,JSONPath=`.spec.providerID`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`

// KubridMachine is the Schema for the kubridmachines API
type KubridMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubridMachineSpec   `json:"spec,omitempty"`
	Status KubridMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubridMachineList contains a list of KubridMachine
type KubridMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubridMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubridMachine{}, &KubridMachineList{})
}
