/*
Copyright 2021.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HDFSSpec defines the desired state of HDFS
type HDFSSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Namenode      NamenodeSet   `json:"namenode"`

	Journalnode   Journalnode    `json:"journalnode"`

	Datanode      datanode      `json:"datanode"`
}

type NamenodeSet struct {

	Name string `json:"name"`

	Replicas int32 `json:"replicas"`

	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`

	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

}

type Journalnode struct {

	Name string `json:"name"`

	Replicas int32 `json:"replicas"`

	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`

	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

}

type datanode struct {

	Name string `json:"name"`

	PodTemplate corev1.PodTemplateSpec `json:"podTemplate,omitempty"`

	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

}


// HDFSStatus defines the observed state of HDFS
type HDFSStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HDFS is the Schema for the hdfs API
type HDFS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HDFSSpec   `json:"spec,omitempty"`
	Status HDFSStatus `json:"status,omitempty"`
}

// IsMarkedForDeletion returns true if the hdfs is going to be deleted
func (in HDFS) IsMarkedForDeletion() bool {
	return !in.DeletionTimestamp.IsZero()
}
//+kubebuilder:object:root=true

// HDFSList contains a list of HDFS
type HDFSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HDFS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HDFS{}, &HDFSList{})
}
