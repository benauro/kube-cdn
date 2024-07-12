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

// ContentDeliveryNetworkSpec defines the desired state of ContentDeliveryNetwork
type (
	ContentDeliveryNetworkSpec struct {
		// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
		// Important: Run "make" to regenerate code after modifying this file

		// Source of the original content
		Origin string `json:"origin"`
		// CDN node domain name
		DomainName string `json:"domainName"`
		// Caching policy
		CacheBehavior string `json:"cacheBehavior,omitempty"`
		// Cache Rules
		CacheRules []CacheRule `json:"cacheRules,omitempty"`
		// SSL/TLS configuration
		SSLConfig *SSLConfig `json:"sslConfig,omitempty"`
		// Replicas
		MinReplicas int `json:"minReplicas"`
		MaxReplicas int `json:"maxReplicas"`
	}

	// CacheRule defines a specific caching rule
	CacheRule struct {
		PathPattern string `json:"pathPattern"`
		TTL         int    `json:"ttl"` // in seconds
	}

	// SSLConfig defines the SSL/TLS configuration for the CDN
	SSLConfig struct {
		Enabled bool   `json:"enabled"`
		Cert    string `json:"cert,omitempty"`
		Key     string `json:"key,omitempty"`
	}
)

// ContentDeliveryNetworkStatus defines the observed state of ContentDeliveryNetwork
type (
	ContentDeliveryNetworkStatus struct {
		// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
		// Important: Run "make" to regenerate code after modifying this file

		// CDN distribution status
		State string `json:"state"`
		// List of IP addresses of the CDN nodes
		Nodes []string `json:"nodes,omitempty"`
		// Last updated time
		LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
		// Metrics for monitoring
		Metrics CDNMetrics `json:"metrics,omitempty"`
	}

	// CDNMetrics contains monitoring metrics for the CDN
	CDNMetrics struct {
		RequestsPerSecond string `json:"requestsPerSecond"`
		CacheHitRate      string `json:"cacheHitRate"`
		AverageLatency    string `json:"averageLatency"` // in milliseconds
	}
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ContentDeliveryNetwork is the Schema for the contentdeliverynetworks API
type ContentDeliveryNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContentDeliveryNetworkSpec   `json:"spec,omitempty"`
	Status ContentDeliveryNetworkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContentDeliveryNetworkList contains a list of ContentDeliveryNetwork
type ContentDeliveryNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContentDeliveryNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContentDeliveryNetwork{}, &ContentDeliveryNetworkList{})
}
