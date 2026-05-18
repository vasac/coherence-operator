/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package v1

var operatorInternalAnnotations = map[string]struct{}{
	AnnotationLastError:               {},
	AnnotationErrorCount:              {},
	AnnotationLastRecoveryAttempt:     {},
	AnnotationLastUnhandledError:      {},
	AnnotationDiagnosticInfo:          {},
	AnnotationFinalizerBypass:         {},
	AnnotationPDBIssueDetected:        {},
	AnnotationPDBIssueTime:            {},
	AnnotationForceReconcile:          {},
	AnnotationQuotaIssueDetected:      {},
	AnnotationQuotaIssueTime:          {},
	AnnotationSchedulingIssueDetected: {},
	AnnotationSchedulingIssueTime:     {},
}

// FilterOperatorInternalAnnotations returns a copy of annotations without keys
// that are private to the operator's error handling and recovery logic.
//
// The filter is intentionally exact rather than prefix-based: Bug39366679/PLAN.md
// requires preventing recursive last-error/diagnostic data from being copied to
// child resources while still preserving legitimate user annotations under the
// coherence.oracle.com domain.
func FilterOperatorInternalAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	filtered := make(map[string]string, len(annotations))
	for k, v := range annotations {
		if _, internal := operatorInternalAnnotations[k]; internal {
			// Child resources should not inherit operator-owned diagnostics; otherwise
			// a failed patch can echo those values back into the next desired resource.
			continue
		}
		filtered[k] = v
	}
	return filtered
}
