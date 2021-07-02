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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InterlokSpec defines the desired state of Interlok
type InterlokSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The number of required instances.
	Instances int32 `json:"instances,omitempty"`
	// Your Interlok image you want to be managed.
	Image string `json:"image,omitempty"`
	// Your Interlok webserver port you want to open.
	JettyPort int32 `json:"jetty-port,omitempty"`
	// Set to true if you want to run Interlok in profiler mode (assumes your image has the required dependencies in place).
	WithProfiler bool `json:"profiler,omitempty"`
}

// InterlokStatus defines the observed state of Interlok
type InterlokStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// String status of a single instance.
	Status string `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="instances",type=string,JSONPath=`.spec.instances`
//+kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`
//+kubebuilder:subresource:status

// Interlok is the Schema for the interloks API
type Interlok struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InterlokSpec   `json:"spec,omitempty"`
	Status InterlokStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InterlokList contains a list of Interlok
type InterlokList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Interlok `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Interlok{}, &InterlokList{})
}
