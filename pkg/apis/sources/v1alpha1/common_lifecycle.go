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
	"context"

	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/status"
)

// eventSourceConditionSet is a common set of conditions for event sources
// objects.
var eventSourceConditionSet = apis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *EventSourceStatus) InitializeConditions() {
	eventSourceConditionSet.Manage(s).InitializeConditions()
}

// MarkSink sets the SinkProvided condition to True using the given URI.
func (s *EventSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if uri == nil {
		eventSourceConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
			ReasonSinkEmpty, "The sink has no URI")
		return
	}
	eventSourceConditionSet.Manage(s).MarkTrue(ConditionSinkProvided)
}

// MarkNoSink sets the SinkProvided condition to False.
func (s *EventSourceStatus) MarkNoSink() {
	s.SinkURI = nil
	eventSourceConditionSet.Manage(s).MarkFalse(ConditionSinkProvided,
		ReasonSinkNotFound, "The sink does not exist or its URI is not set")
}

// PropagateAvailability uses the readiness of the provided Deployment or
// Service to determine whether the Deployed condition should be marked as True
// or False.
func (s *EventSourceStatus) PropagateAvailability(obj interface{}) {
	switch o := obj.(type) {
	case *servingv1.Service:
		s.propagateServiceAvailability(o)
	}
}

// PropagateDeploymentAvailability uses the readiness of the provided
// Deployment to determine whether the Deployed condition should be marked as
// True or False.
// Given an optional PodInterface, the status of dependant Pods is inspected to
// generate a more meaningful failure reason in case of non-ready status of the
// Deployment.
func (s *EventSourceStatus) PropagateDeploymentAvailability(ctx context.Context,
	d *appsv1.Deployment, pi coreclientv1.PodInterface) {

	// Deployments are not addressable
	s.Address = nil

	if d == nil {
		eventSourceConditionSet.Manage(s).MarkUnknown(ConditionDeployed, ReasonUnavailable,
			"The status of the adapter Deployment can not be determined")
		return
	}

	if duck.DeploymentIsAvailable(&d.Status, false) {
		eventSourceConditionSet.Manage(s).MarkTrue(ConditionDeployed)
		return
	}

	reason := ReasonUnavailable
	msg := "The adapter Deployment is unavailable"

	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Message != "" {
			msg += ": " + cond.Message
		}
	}

	if pi != nil {
		ws, err := status.DeploymentPodsWaitingState(d, pi)
		if err != nil {
			logging.FromContext(ctx).Warn("Unable to look up statuses of dependant Pods", zap.Error(err))
		} else if ws != nil {
			reason = ws.Reason
			msg += ": " + ws.Message
		}
	}

	eventSourceConditionSet.Manage(s).MarkFalse(ConditionDeployed, reason, msg)
}

func (s *EventSourceStatus) propagateServiceAvailability(ksvc *servingv1.Service) {
	if ksvc == nil {
		eventSourceConditionSet.Manage(s).MarkUnknown(ConditionDeployed, ReasonUnavailable,
			"The status of the adapter Service can not be determined")
		return
	}

	if s.Address == nil {
		s.Address = &duckv1.Addressable{}
	}
	s.Address.URL = ksvc.Status.URL

	if ksvc.IsReady() {
		eventSourceConditionSet.Manage(s).MarkTrue(ConditionDeployed)
		return
	}

	msg := "The adapter Service is unavailable"
	readyCond := ksvc.Status.GetCondition(servingv1.ServiceConditionReady)
	if readyCond != nil && readyCond.Message != "" {
		msg += ": " + readyCond.Message
	}

	eventSourceConditionSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, msg)
}
