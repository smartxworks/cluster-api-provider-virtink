package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubridMachineTemplateSpec defines the desired state of KubridMachineTemplate
type KubridMachineTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template KubridMachineTemplateSpecTemplate `json:"template"`
}

type KubridMachineTemplateSpecTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KubridMachineSpec `json:"spec,omitempty"`
}

// KubridMachineTemplateStatus defines the observed state of KubridMachineTemplate
type KubridMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubridMachineTemplate is the Schema for the kubridmachinetemplates API
type KubridMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubridMachineTemplateSpec   `json:"spec,omitempty"`
	Status KubridMachineTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubridMachineTemplateList contains a list of KubridMachineTemplate
type KubridMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubridMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubridMachineTemplate{}, &KubridMachineTemplateList{})
}
