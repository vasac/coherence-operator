/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package metadatapatch

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateAnnotationsMergePatch creates a metadata-only JSON merge patch for
// annotation updates.
//
// Bug39366679/PLAN.md requires error tracking to avoid full-object updates: a
// metadata-only patch updates diagnostics without re-sending a bloated status.
func CreateAnnotationsMergePatch(annotations map[string]string) ([]byte, bool, error) {
	if len(annotations) == 0 {
		return nil, false, nil
	}
	data, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// PatchAnnotations applies a metadata-only annotation merge patch.
func PatchAnnotations(ctx context.Context, c client.Client, obj client.Object, annotations map[string]string) (bool, []byte, error) {
	data, changed, err := CreateAnnotationsMergePatch(annotations)
	if err != nil || !changed {
		return changed, data, err
	}
	err = c.Patch(ctx, obj, client.RawPatch(types.MergePatchType, data))
	return changed, data, err
}
