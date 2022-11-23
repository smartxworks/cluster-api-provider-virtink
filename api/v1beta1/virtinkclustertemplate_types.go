package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtinkClusterTemplateSpec defines the desired state of VirtinkClusterTemplate
type VirtinkClusterTemplateSpec struct {
	Template VirtinkClusterTemplateResource `json:"template"`
}

type VirtinkClusterTemplateResource struct {
	Spec VirtinkClusterSpec `json:"spec"`
}

//+kubebuilder:object:root=true

// VirtinkClusterTemplate is the Schema for the virtinkclustertemplates API
type VirtinkClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtinkClusterTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VirtinkClusterTemplateList contains a list of VirtinkClusterTemplate
type VirtinkClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtinkClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtinkClusterTemplate{}, &VirtinkClusterTemplateList{})
}
