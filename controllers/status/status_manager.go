/*
 * Copyright (c) 2020, 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package status

import (
	"context"

	"github.com/go-logr/logr"
	coh "github.com/oracle/coherence-operator/api/v1"
	"github.com/oracle/coherence-operator/pkg/operator"
	"github.com/oracle/coherence-operator/pkg/patching"
	"github.com/oracle/coherence-operator/pkg/statuspatch"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatusManager manages the status of Coherence resources
type StatusManager struct {
	Client  client.Client
	Log     logr.Logger
	Patcher patching.ResourcePatcher
}

// UpdateCoherenceStatusPhase updates the phase of a Coherence resource
func (sm *StatusManager) UpdateCoherenceStatusPhase(ctx context.Context, namespacedName types.NamespacedName, phase coh.ConditionType) error {
	// Get the latest version of the Coherence resource
	deployment := &coh.Coherence{}
	err := sm.Client.Get(ctx, namespacedName, deployment)
	if err != nil {
		return errors.Wrapf(err, "getting Coherence resource %s/%s", namespacedName.Namespace, namespacedName.Name)
	}

	updated := deployment.DeepCopy()
	// Bug39366679/PLAN.md: use SetCondition rather than writing Phase directly so
	// scalar phase and status.conditions stay consistent on this status path.
	if !updated.Status.SetCondition(deployment, coh.Condition{Type: phase, Status: corev1.ConditionTrue}) {
		return nil
	}

	// Update the resource
	return sm.patchStatus(ctx, deployment, updated)
}

// UpdateDeploymentStatusHash updates the hash in the status of a Coherence resource
func (sm *StatusManager) UpdateDeploymentStatusHash(ctx context.Context, namespacedName types.NamespacedName, hash string) error {
	// Get the latest version of the Coherence resource
	deployment := &coh.Coherence{}
	err := sm.Client.Get(ctx, namespacedName, deployment)
	if err != nil {
		return errors.Wrapf(err, "getting Coherence resource %s/%s", namespacedName.Namespace, namespacedName.Name)
	}

	// Update the status hash
	updated := deployment.DeepCopy()
	updated.Status.Hash = hash
	updated.Status.SetVersion(operator.GetVersion())

	// Update the resource
	return sm.patchStatus(ctx, deployment, updated)
}

func (sm *StatusManager) patchStatus(ctx context.Context, original, updated *coh.Coherence) error {
	// Bug39366679/PLAN.md: this path previously created strategic-merge bytes
	// and applied them as a JSON merge patch, which made empty conditions append.
	patched, data, err := statuspatch.PatchStatus(ctx, sm.Client, original, updated, true)
	if err != nil {
		return errors.Wrapf(err, "updating status for Coherence resource %s/%s", original.Namespace, original.Name)
	}
	if patched {
		sm.Log.Info("Patched status", "Namespace", original.Namespace, "Name", original.Name, "PatchSize", len(data))
	}
	return nil
}
