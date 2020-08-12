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

package zendesksource

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/reconciler"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	reconcilerv1alpha1 "github.com/triggermesh/knative-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
)

// Reconciler implements controller.Reconciler for the event source type.
type Reconciler struct {
	base       common.GenericServiceReconciler
	kubeClient kubernetes.Interface
	adapterCfg *adapterConfig
}

// Check that our Reconciler implements Interface
var _ reconcilerv1alpha1.Interface = (*Reconciler)(nil)

// Check that our Reconciler implements Finalizer
var _ reconcilerv1alpha1.Finalizer = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.ZendeskSource) reconciler.Event {
	// inject source into context for usage in reconciliation logic
	ctx = v1alpha1.WithSource(ctx, src)

	if err := r.base.ReconcileSource(ctx, adapterServiceBuilder(src, r.adapterCfg)); err != nil {
		return fmt.Errorf("failed to reconcile source: %w", err)
	}

	return r.ensureZendeskTargetAndTrigger(ctx)
}

// FinalizeKind is called when the resource is deleted.
func (r *Reconciler) FinalizeKind(ctx context.Context, src *v1alpha1.ZendeskSource) reconciler.Event {
	// inject source into context for usage in finalization logic
	ctx = v1alpha1.WithSource(ctx, src)

	// The finalizer blocks the deletion of the source object until
	// ensureNoZendeskTargetAndTrigger succeeds to ensure that we don't
	// leave any dangling Zendesk Target/Trigger behind us.
	return r.ensureNoZendeskTargetAndTrigger(ctx)
}
