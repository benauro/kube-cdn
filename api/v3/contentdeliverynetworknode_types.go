/*
Copyright 2024.

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

package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContentDeliveryNetworkNodeSpec defines the desired state of ContentDeliveryNetworkNode
type ContentDeliveryNetworkNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Size for caching
	CacheSize int `json:"cacheSize"`
}

// ContentDeliveryNetworkNodeStatus defines the observed state of ContentDeliveryNetworkNode
type ContentDeliveryNetworkNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Available bool `json:"available"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ContentDeliveryNetworkNode is the Schema for the contentdeliverynetworknodes API
type ContentDeliveryNetworkNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContentDeliveryNetworkNodeSpec   `json:"spec,omitempty"`
	Status ContentDeliveryNetworkNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContentDeliveryNetworkNodeList contains a list of ContentDeliveryNetworkNode
type ContentDeliveryNetworkNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContentDeliveryNetworkNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContentDeliveryNetworkNode{}, &ContentDeliveryNetworkNodeList{})
}
