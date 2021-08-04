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

const (
	ProbeAssistantSpecChangePodPolicyWaitUpdate string = "waitUpdate"
	ProbeAssistantSpecChangePodPolicyKeepSave   string = "keepSave"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProbeAssistantSpec defines the desired state of ProbeAssistant
type ProbeAssistantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Update Liveness/Readiness and metadata policy for Pods bound to ProbeAssistant
	// when ProbeAssistant is updated or removed.
	// "waitUpdate"(default): The next Pod update.
	// "keepSave" warning it's best not to use.
	ChangePodPolicy string `json:"changePodPolicy,omitempty"`

	// a configmap inject to pod.
	// readiness Probe will exec <configMap-mount-point>/readiness/pre.sh and after.sh.
	DefaultReadinessTmpl string `json:"defaultReadinessTmpl,omitempty"`

	// a configmap inject to pod.
	// readiness Probe will exec <configMap-mount-point>/liveness/pre.sh and after.sh.
	DefaultLivenessTmpl string `json:"defaultLivenessTmpl,omitempty"`

	// Selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector"`

	// Foo is an example field of ProbeAssistant. Edit probeassistant_types.go to remove/update
	// MaxSavePods indicates the maximum number of problem containers to be retained.
	MaxSavePods *int32 `json:"maxSavePods"`

	// Namespace ProbeAssistant will only match the pods in the namespace
	// otherwise, match pods in all namespaces(in cluster)
	Namespace string `json:"namespace,omitempty"`
}

// ProbeAssistantStatus defines the observed state of ProbeAssistant
type ProbeAssistantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// probe exec script got errors pods.
	ProblemPods int32 `json:"problemPods"`

	// matchedPods is the number of Pods whose labels are matched with this SidecarSet's selector and are created after sidecarset creates
	MatchedPods int32 `json:"matchedPods"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProbeAssistant is the Schema for the probeassistants API
type ProbeAssistant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProbeAssistantSpec   `json:"spec,omitempty"`
	Status ProbeAssistantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProbeAssistantList contains a list of ProbeAssistant
type ProbeAssistantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProbeAssistant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProbeAssistant{}, &ProbeAssistantList{})
}
