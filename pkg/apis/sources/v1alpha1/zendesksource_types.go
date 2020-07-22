/*
Copyright (c) 2020 TriggerMesh Inc.

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	pkgapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZendeskSource is the schema for the Zendesk source
type ZendeskSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ZendeskSourceSpec `json:"spec"`
	// +optional
	Status ZendeskSourceStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *ZendeskSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ZendeskSource")
}

var (
	_ runtime.Object      = (*ZendeskSource)(nil)
	_ kmeta.OwnerRefable  = (*ZendeskSource)(nil)
	_ pkgapis.Validatable = (*ZendeskSource)(nil)
	_ pkgapis.Defaultable = (*ZendeskSource)(nil)
	_ duckv1.KRShaped     = (*ZendeskSource)(nil)
)

// ZendeskSourceSpec holds the desired state of the ZendeskSource (from the client).
type ZendeskSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// Token identifies the API token used for creating the proper credentials to interface with Zendesk
	// allowing the source to auto-register the webhook to authenticate callbacks.
	Token SecretValueFromSource `json:"token,omitempty"`

	// Email identifies the email used for creating the proper credentials to interface with Zendesk
	// allowing the source to auto-register the webhook to authenticate callbacks.
	Email string `json:"email,omitempty"`

	// WebhookPassword used for basic authentication for events sent from Zendesk
	// to the adapter.
	WebhookPassword SecretValueFromSource `json:"webhookPassword,omitempty"`

	// WebhookUsername used for basic authentication for events sent from Zendesk
	// to the adapter.
	WebhookUsername string `json:"webhookUsername,omitempty"`

	// Subdomain identifies Zendesk subdomain
	Subdomain string `json:"subdomain,omitempty"`
}

// ZendeskSourceStatus communicates the observed state of the ZendeskSource (from the controller).
type ZendeskSourceStatus struct {
	duckv1.SourceStatus  `json:",inline"`
	duckv1.AddressStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZendeskSourceList is a list of ZendeskSource resources
type ZendeskSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZendeskSource `json:"items"`
}