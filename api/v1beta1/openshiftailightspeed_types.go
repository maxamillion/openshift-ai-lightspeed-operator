/*
Copyright 2025.

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

package v1beta1

import (
	"github.com/opendatahub-io/openshift-ai-lightspeed-operator/pkg/common/condition"
	"github.com/opendatahub-io/openshift-ai-lightspeed-operator/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Container image fall-back defaults

	// OpenShiftAILightspeedContainerImage is the fall-back container image for OpenShiftAILightspeed
	OpenShiftAILightspeedContainerImage = "quay.io/opendatahub-io/openshift-ai-lightspeed-rag-content:rhoai-docs-2025.1"
	MaxTokensForResponseDefault         = 2048
)

// OpenShiftAILightspeedSpec defines the desired state of OpenShiftAILightspeed
type OpenShiftAILightspeedSpec struct {
	OpenShiftAILightspeedCore `json:",inline"`

	// +kubebuilder:validation:Optional
	// ContainerImage for the OpenShift AI Lightspeed RAG container (will be set to environmental default if empty)
	RAGImage string `json:"ragImage"`
}

// OpenShiftAILightspeedCore defines the desired state of OpenShiftAILightspeed
type OpenShiftAILightspeedCore struct {
	// +kubebuilder:validation:Required
	// URL pointing to the LLM
	LLMEndpoint string `json:"llmEndpoint"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=azure_openai;bam;openai;watsonx;rhoai_vllm;rhelai_vllm;fake_provider
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Provider Type"
	// Type of the provider serving the LLM
	LLMEndpointType string `json:"llmEndpointType"`

	// +kubebuilder:validation:Required
	// Name of the model to use at the API endpoint provided in LLMEndpoint
	ModelName string `json:"modelName"`

	// +kubebuilder:validation:Required
	// Secret name containing API token for the LLMEndpoint. The key for the field
	// in the secret that holds the token should be "apitoken".
	LLMCredentials string `json:"llmCredentials"`

	// +kubebuilder:validation:Optional
	// Configmap name containing a CA Certificates bundle
	TLSCACertBundle string `json:"tlsCACertBundle"`

	// +kubebuilder:validation:Optional
	// MaxTokensForResponse defines the maximum number of tokens to be used for the response generation
	MaxTokensForResponse int `json:"maxTokensForResponse,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="openshift-marketplace"
	// Namespace where the CatalogSource containing the OLS operator is located
	CatalogSourceNamespace string `json:"catalogSourceNamespace"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="redhat-operators"
	// Name of the CatalogSource that contains the OLS Operator
	CatalogSourceName string `json:"catalogSourceName"`

	// +kubebuilder:validation:Optional
	// Project ID for LLM providers that require it (e.g., WatsonX)
	LLMProjectID string `json:"llmProjectID,omitempty"`

	// +kubebuilder:validation:Optional
	// Deployment name for LLM providers that require it (e.g., Microsoft Azure OpenAI)
	LLMDeploymentName string `json:"llmDeploymentName,omitempty"`

	// +kubebuilder:validation:Optional
	// LLM API Version for LLM providers that require it (e.g., Microsoft Azure OpenAI)
	LLMAPIVersion string `json:"llmAPIVersion,omitempty"`

	// +kubebuilder:validation:Optional
	// Disable feedback collection
	FeedbackDisabled bool `json:"feedbackDisabled,omitempty"`

	// +kubebuilder:validation:Optional
	// Disable conversation transcripts collection
	TranscriptsDisabled bool `json:"transcriptsDisabled,omitempty"`
}

// OpenShiftAILightspeedStatus defines the observed state of OpenShiftAILightspeed
type OpenShiftAILightspeedStatus struct {
	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ObservedGeneration - the most recent generation observed for this object.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

// OpenShiftAILightspeed is the Schema for the openshiftailightspeeds API
type OpenShiftAILightspeed struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenShiftAILightspeedSpec   `json:"spec,omitempty"`
	Status OpenShiftAILightspeedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpenShiftAILightspeedList contains a list of OpenShiftAILightspeed
type OpenShiftAILightspeedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenShiftAILightspeed `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenShiftAILightspeed{}, &OpenShiftAILightspeedList{})
}

// IsReady - returns true if OpenShiftAILightspeed is reconciled successfully
func (instance OpenShiftAILightspeed) IsReady() bool {
	return instance.Status.Conditions.IsTrue(OpenShiftAILightspeedReadyCondition)
}

type OpenShiftAILightspeedDefaults struct {
	RAGImageURL          string
	MaxTokensForResponse int
}

var OpenShiftAILightspeedDefaultValues OpenShiftAILightspeedDefaults

// SetupDefaults - initializes OpenShiftAILightspeedDefaultValues with default values from env vars
func SetupDefaults() {
	// Acquire environmental defaults and initialize OpenShiftAILightspeed defaults with them
	openShiftAILightspeedDefaults := OpenShiftAILightspeedDefaults{
		RAGImageURL: util.GetEnvVar(
			"RELATED_IMAGE_OPENSHIFT_AI_LIGHTSPEED_IMAGE_URL_DEFAULT", OpenShiftAILightspeedContainerImage),
		MaxTokensForResponse: MaxTokensForResponseDefault,
	}

	OpenShiftAILightspeedDefaultValues = openShiftAILightspeedDefaults
}
