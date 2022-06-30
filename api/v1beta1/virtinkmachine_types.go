package v1beta1

import (
	virtv1alpha1 "github.com/smartxworks/virtink/pkg/apis/virt/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtinkMachineSpec defines the desired state of VirtinkMachine
type VirtinkMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProviderID *string `json:"providerID,omitempty"`

	VMSpec virtv1alpha1.VirtualMachineSpec `json:"vmSpec"`
}

// VirtinkMachineStatus defines the observed state of VirtinkMachine
type VirtinkMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderID",type=string,JSONPath=`.spec.providerID`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`

// VirtinkMachine is the Schema for the virtinkmachines API
type VirtinkMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtinkMachineSpec   `json:"spec,omitempty"`
	Status VirtinkMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtinkMachineList contains a list of VirtinkMachine
type VirtinkMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtinkMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtinkMachine{}, &VirtinkMachineList{})
}
