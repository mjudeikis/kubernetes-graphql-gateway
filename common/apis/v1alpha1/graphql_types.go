package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GraphQL object represents a request to build a GraphQL endpoint
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GraphQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GraphQLSpec         `json:"spec,omitempty"`
	Status ClusterAccessStatus `json:"status,omitempty"`
}

type GraphQLSpec struct {
	// URL is the GraphQL endpoint URL
	URL string `json:"url"`
}

type GraphQLStatus struct {
	// Conditions represent the latest available observations of the GraphQL state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true

// GraphQLList contains a list of GraphQL
type GraphQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GraphQL `json:"items"`
}

// GetConditions returns the conditions from the GraphQL status
// This method implements the RuntimeObjectConditions interface
func (g *GraphQL) GetConditions() []metav1.Condition {
	return g.Status.Conditions
}

// SetConditions sets the conditions in the GraphQL status
// This method implements the RuntimeObjectConditions interface
func (g *GraphQL) SetConditions(conditions []metav1.Condition) {
	g.Status.Conditions = conditions
}
