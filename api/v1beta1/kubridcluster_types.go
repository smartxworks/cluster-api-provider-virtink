/*
 * Copyright (C) 2022 SmartX, Inc. <info@smartx.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubridClusterSpec defines the desired state of KubridCluster
type KubridClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ControlPlaneEndpoint capiv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`
}

// KubridClusterStatus defines the observed state of KubridCluster
type KubridClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.controlPlaneEndpoint.host`
//+kubebuilder:printcolumn:name="Port",type=integer,JSONPath=`.spec.controlPlaneEndpoint.port`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`

// KubridCluster is the Schema for the kubridclusters API
type KubridCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubridClusterSpec   `json:"spec,omitempty"`
	Status KubridClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubridClusterList contains a list of KubridCluster
type KubridClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubridCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubridCluster{}, &KubridClusterList{})
}
