/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type TestPodSpec struct {
	Image string `json:"image,omitempty"`
}

type TestCaseScope struct {
	Performance   bool `json:"performance,omitempty"`
	Realtime      bool `json:"realtime,omitempty"`
	Configuration bool `json:"configuration,omitempty"`
	Connectivity  bool `json:"connectivity,omitempty"`
	Robot         bool `json:"robot,omitempty"`
	Postman       bool `json:"postman,omitempty"`
	Cypress       bool `json:"cypress,omitempty"`
}

// TesterSpec defines the desired state of Tester
type TesterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Tester. Edit tester_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	TestPodSpec   TestPodSpec   `json:"testPodSpec,omitempty"`
	TestCaseScope TestCaseScope `json:"testCaseScope,omitempty"`
	TestDebugMode bool          `json:"debugMode,omitempty"`
}

// TesterStatus defines the observed state of Tester
type TesterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CaseStatus string             `json:"caseStatus,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.caseStatus`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Tester is the Schema for the testers API
type Tester struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TesterSpec   `json:"spec,omitempty"`
	Status TesterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TesterList contains a list of Tester
type TesterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tester `json:"items"`
}

const (
	DefaultTestreportImg string = "container-infra-local.hzisoj70.china.nsn-net.net/cnfmark/testreporter:v0.2"
	// DefaultKubeconfigPath is the default local path of kubeconfig file.
)

func init() {
	SchemeBuilder.Register(&Tester{}, &TesterList{})
}
