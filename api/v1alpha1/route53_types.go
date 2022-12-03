/*
Copyright 2022.

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

// Route53Spec defines the desired state of Route53
type Route53Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Domain you wanto to create (e.g. "mytest.jmichaud.net")
	// +required
	Domain string `json:"domain"`

	// Record type you want to create (e.g. "A|CNAME|TXT")
	// +required
	RecordType string `json:"recordtype"`

	// Value you want to create (e.g. "8.8.8.8")
	// +required
	Value string `json:"value"`

	// The amount of time, in seconds, that you want DNS recursive resolvers to cache information about this record (default 60)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=60
	// +kubebuilder:validation:Minimum=60
	TTL int64 `json:"ttl"`
}

// Route53Status defines the observed state of Route53
type Route53Status struct {
	// Represents the observations of a Route53's current state.
	// Route53.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// Route53.status.conditions.status are one of True, False, Unknown.
	// Route53.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// Route53.status.conditions.Message is a human readable message indicating details about the transition.
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

// Route53 is the Schema for the route53s API
// +kubebuilder:subresource:status
type Route53 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Route53Spec   `json:"spec,omitempty"`
	Status Route53Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Route53List contains a list of Route53
type Route53List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Route53 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Route53{}, &Route53List{})
}
