package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtinkMachineTemplateSpec defines the desired state of VirtinkMachineTemplate
type VirtinkMachineTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template VirtinkMachineTemplateSpecTemplate `json:"template"`
}

type VirtinkMachineTemplateSpecTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtinkMachineSpec `json:"spec,omitempty"`
}

// VirtinkMachineTemplateStatus defines the observed state of VirtinkMachineTemplate
type VirtinkMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VirtinkMachineTemplate is the Schema for the virtinkmachinetemplates API
type VirtinkMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtinkMachineTemplateSpec   `json:"spec,omitempty"`
	Status VirtinkMachineTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtinkMachineTemplateList contains a list of VirtinkMachineTemplate
type VirtinkMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtinkMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtinkMachineTemplate{}, &VirtinkMachineTemplateList{})
}
