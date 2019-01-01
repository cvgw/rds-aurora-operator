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

// SubnetGroupDeletionSpec defines the desired state of SubnetGroupDeletion
type SubnetGroupDeletionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SubnetGroupName string `json:"subnet_group_name,omitempty"`
}

// SubnetGroupDeletionStatus defines the observed state of SubnetGroupDeletion
type SubnetGroupDeletionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State      string `json:"state,omitempty"`
	ReadySince int64  `json:"ready_since,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubnetGroupDeletion is the Schema for the subnetgroupdeletions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SubnetGroupDeletion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetGroupDeletionSpec   `json:"spec,omitempty"`
	Status SubnetGroupDeletionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubnetGroupDeletionList contains a list of SubnetGroupDeletion
type SubnetGroupDeletionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubnetGroupDeletion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubnetGroupDeletion{}, &SubnetGroupDeletionList{})
}
