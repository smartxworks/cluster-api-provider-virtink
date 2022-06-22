package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirTinkMachineTemplateSpec defines the desired state of VirTinkMachineTemplate
type VirTinkMachineTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template VirTinkMachineTemplateSpecTemplate `json:"template"`
}

type VirTinkMachineTemplateSpecTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirTinkMachineSpec `json:"spec,omitempty"`
}

// VirTinkMachineTemplateStatus defines the observed state of VirTinkMachineTemplate
type VirTinkMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VirTinkMachineTemplate is the Schema for the virtinkmachinetemplates API
type VirTinkMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirTinkMachineTemplateSpec   `json:"spec,omitempty"`
	Status VirTinkMachineTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirTinkMachineTemplateList contains a list of VirTinkMachineTemplate
type VirTinkMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirTinkMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirTinkMachineTemplate{}, &VirTinkMachineTemplateList{})
}
