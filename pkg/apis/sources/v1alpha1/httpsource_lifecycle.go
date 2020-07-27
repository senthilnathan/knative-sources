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
	"k8s.io/apimachinery/pkg/runtime/schema"

	pkgapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (*HTTPSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("HttpSource")
}

// GetConditionSet implements duckv1.KRShaped.
func (s *HTTPSource) GetConditionSet() pkgapis.ConditionSet {
	return eventSourceConditionSet
}

// GetStatus implements duckv1.KRShaped.
func (s *HTTPSource) GetStatus() *duckv1.Status {
	return &s.Status.Status
}

// GetSink implements EventSource.
func (s *HTTPSource) GetSink() *duckv1.Destination {
	return &s.Spec.Sink
}

// GetSourceStatus implements EventSource.
func (s *HTTPSource) GetSourceStatus() *EventSourceStatus {
	return &s.Status
}

// AsEventSource implements EventSource.
func (s *HTTPSource) AsEventSource() string {
	if s.Spec.EventSource != nil {
		return *s.Spec.EventSource
	}

	sourceName := s.Name
	if s.Namespace != "" {
		sourceName = s.Namespace + "." + sourceName
	}

	return sourceName
}

// GetEventTypes implements EventSource.
func (s *HTTPSource) GetEventTypes() []string {
	return []string{
		s.Spec.EventType,
	}
}
