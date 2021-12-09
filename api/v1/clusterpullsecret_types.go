/*


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

// ClusterPullSecretStatus defines the observed state of ClusterPullSecret
type ClusterPullSecretStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="SecretName",type=string,JSONPath=`.spec.secretRef.name`
//+kubebuilder:printcolumn:name="SecretNamespace",type=string,JSONPath=`.spec.secretRef.namespace`

// ClusterPullSecret is the Schema for the clusterpullsecrets API
type ClusterPullSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPullSecretSpec   `json:"spec,omitempty"`
	Status ClusterPullSecretStatus `json:"status,omitempty"`
}

// ObjectMeta contains enough information to locate the referenced Kubernetes resource object in any
// namespace.
type ObjectMeta struct {
	// Name of the referent.
	// +required
	Name string `json:"name"`

	// Namespace of the referent, when not specified it acts as LocalObjectReference.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// ClusterPullSecretSpec defines the desired state of ClusterPullSecret
type ClusterPullSecretSpec struct {
	SecretRef *ObjectMeta `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterPullSecretList contains a list of ClusterPullSecret
type ClusterPullSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPullSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPullSecret{}, &ClusterPullSecretList{})
}
