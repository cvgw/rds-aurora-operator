/*
Copyright 2018 Cole Wippern.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ParameterGroupDeletionSpec defines the desired state of ParameterGroupDeletion
type ParameterGroupDeletionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ParameterGroupName string `json:"parameter_group_name,omitempty"`
}

// ParameterGroupDeletionStatus defines the observed state of ParameterGroupDeletion
type ParameterGroupDeletionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State      string `json:"state,omitempty"`
	ReadySince int64  `json:"ready_since,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ParameterGroupDeletion is the Schema for the parametergroupdeletions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ParameterGroupDeletion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ParameterGroupDeletionSpec   `json:"spec,omitempty"`
	Status ParameterGroupDeletionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ParameterGroupDeletionList contains a list of ParameterGroupDeletion
type ParameterGroupDeletionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ParameterGroupDeletion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ParameterGroupDeletion{}, &ParameterGroupDeletionList{})
}
