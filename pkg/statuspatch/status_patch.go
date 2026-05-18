/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package statuspatch

import (
	"context"
	"encoding/json"
	"reflect"

	coh "github.com/oracle/coherence-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateStatusMergePatch creates a compact JSON merge patch for a Coherence or
// CoherenceJob status update.
//
// The builder intentionally avoids strategic-merge status patches because
// Bug39366679/PLAN.md identified that strategic merge bytes applied as
// types.MergePatchType caused empty conditions to be appended instead of merged.
// The expected effect is a small patch with only changed scalar status fields and
// a single normalized replacement conditions array when conditions changed.
func CreateStatusMergePatch(original, updated coh.CoherenceResource, forceConditions bool) ([]byte, bool, error) {
	if original == nil || updated == nil {
		return nil, false, nil
	}

	current := original.GetStatus()
	desired := updated.GetStatus()
	if desired.NormalizeStatus() {
		forceConditions = true
	}

	status := make(map[string]interface{})
	addIfChanged(status, "phase", current.Phase, desired.Phase)
	addIfChanged(status, "coherenceCluster", current.CoherenceCluster, desired.CoherenceCluster)
	addIfChanged(status, "type", current.Type, desired.Type)
	addIfChanged(status, "replicas", current.Replicas, desired.Replicas)
	addIfChanged(status, "currentReplicas", current.CurrentReplicas, desired.CurrentReplicas)
	addIfChanged(status, "readyReplicas", current.ReadyReplicas, desired.ReadyReplicas)
	addIfChanged(status, "active", current.Active, desired.Active)
	addIfChanged(status, "succeeded", current.Succeeded, desired.Succeeded)
	addIfChanged(status, "failed", current.Failed, desired.Failed)
	addIfChanged(status, "role", current.Role, desired.Role)
	addIfChanged(status, "selector", current.Selector, desired.Selector)
	addIfChanged(status, "hash", current.Hash, desired.Hash)
	addIfChanged(status, "actionsExecuted", current.ActionsExecuted, desired.ActionsExecuted)

	conditionsChanged := !reflect.DeepEqual(current.Conditions, desired.Conditions)
	if conditionsChanged || (forceConditions && desired.Conditions != nil) {
		// Replace the whole normalized array. This keeps a Bug39366679 repair patch
		// proportional to the valid conditions, not to the old bloated list.
		status["conditions"] = desired.Conditions
	}
	if !reflect.DeepEqual(current.JobProbes, desired.JobProbes) {
		status["jobProbes"] = desired.JobProbes
	}

	if len(status) == 0 {
		return nil, false, nil
	}

	data, err := json.Marshal(map[string]interface{}{
		"status": status,
	})
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// PatchStatus applies a compact JSON merge patch to the status subresource.
func PatchStatus(ctx context.Context, c client.Client, original, updated coh.CoherenceResource, forceConditions bool) (bool, []byte, error) {
	data, changed, err := CreateStatusMergePatch(original, updated, forceConditions)
	if err != nil || !changed {
		return changed, data, err
	}
	err = c.Status().Patch(ctx, original, client.RawPatch(types.MergePatchType, data))
	return changed, data, err
}

func addIfChanged[T comparable](status map[string]interface{}, name string, current, desired T) {
	if current != desired {
		// Explicit comparison preserves meaningful zero-value transitions such as
		// readyReplicas:0 and actionsExecuted:false in the merge patch.
		status[name] = desired
	}
}
