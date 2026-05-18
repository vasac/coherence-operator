/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package v1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
)

func TestFilterOperatorInternalAnnotationsReturnsCopyWithoutInternalKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	source := map[string]string{
		"user":                        "value",
		coh.AnnotationLastError:       "large error",
		coh.AnnotationDiagnosticInfo:  "diagnostics",
		"coherence.oracle.com/custom": "keep",
	}

	filtered := coh.FilterOperatorInternalAnnotations(source)

	g.Expect(filtered).To(Equal(map[string]string{
		"user":                        "value",
		"coherence.oracle.com/custom": "keep",
	}))
	g.Expect(source).To(HaveKey(coh.AnnotationLastError), "filtering must not mutate the caller's annotation map")
}

func TestCoherenceCreateAnnotationsFiltersInternalKeysFromFinalMap(t *testing.T) {
	g := NewGomegaWithT(t)

	deployment := createTestCoherenceDeployment(coh.CoherenceStatefulSetResourceSpec{
		CoherenceResourceSpec: coh.CoherenceResourceSpec{},
		Global: &coh.GlobalSpec{
			Annotations: map[string]string{
				"global":                 "value",
				coh.AnnotationErrorCount: "42",
				coh.AnnotationLastError:  "from-global",
			},
		},
		StatefulSetAnnotations: map[string]string{
			"statefulset":                     "value",
			coh.AnnotationLastUnhandledError:  "from-statefulset",
			coh.AnnotationSchedulingIssueTime: "from-statefulset",
		},
	})

	annotations := deployment.CreateAnnotations()

	g.Expect(annotations).To(HaveKeyWithValue("global", "value"))
	g.Expect(annotations).To(HaveKeyWithValue("statefulset", "value"))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationErrorCount))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationLastError))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationLastUnhandledError))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationSchedulingIssueTime))
}

func TestCoherenceCreateAnnotationsFiltersInternalKeysFromMetadataFallback(t *testing.T) {
	g := NewGomegaWithT(t)

	deployment := createTestCoherenceDeployment(coh.CoherenceStatefulSetResourceSpec{
		CoherenceResourceSpec: coh.CoherenceResourceSpec{},
	})
	deployment.SetAnnotations(map[string]string{
		"metadata":                    "value",
		"coherence.oracle.com/custom": "keep",
		coh.AnnotationLastError:       "do-not-copy",
	})

	// Bug39366679/PLAN.md: with statefulSetAnnotations unset, metadata fallback
	// must keep user annotations while preventing internal diagnostics from being
	// echoed onto child StatefulSets and feeding recursive patch errors.
	annotations := deployment.CreateAnnotations()

	g.Expect(annotations).To(HaveKeyWithValue("metadata", "value"))
	g.Expect(annotations).To(HaveKeyWithValue("coherence.oracle.com/custom", "keep"))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationLastError))
}

func TestCoherenceJobCreateAnnotationsPreservesGlobalsAndFiltersInternalKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	deployment := createTestCoherenceJobDeployment(coh.CoherenceJobResourceSpec{
		CoherenceResourceSpec: coh.CoherenceResourceSpec{},
		Global: &coh.GlobalSpec{
			Annotations: map[string]string{
				"global":                "value",
				coh.AnnotationLastError: "from-global",
			},
		},
		JobAnnotations: map[string]string{
			"job":                        "value",
			coh.AnnotationDiagnosticInfo: "from-job",
			coh.AnnotationQuotaIssueTime: "from-job",
		},
	})

	annotations := deployment.CreateAnnotations()

	g.Expect(annotations).To(HaveKeyWithValue("global", "value"))
	g.Expect(annotations).To(HaveKeyWithValue("job", "value"))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationLastError))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationDiagnosticInfo))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationQuotaIssueTime))
}

func TestCoherenceJobCreateAnnotationsFiltersInternalKeysFromMetadataFallback(t *testing.T) {
	g := NewGomegaWithT(t)

	deployment := createTestCoherenceJobDeployment(coh.CoherenceJobResourceSpec{
		CoherenceResourceSpec: coh.CoherenceResourceSpec{},
	})
	deployment.SetAnnotations(map[string]string{
		"metadata":                    "value",
		"coherence.oracle.com/custom": "keep",
		coh.AnnotationLastError:       "do-not-copy",
	})

	// Bug39366679/PLAN.md: with jobAnnotations unset, metadata fallback should
	// still reach the batch Job, but operator-owned diagnostic keys must stay on
	// the CoherenceJob CR to avoid recursive patch payload growth.
	annotations := deployment.CreateAnnotations()
	job := deployment.Spec.CreateJob(deployment)

	g.Expect(annotations).To(HaveKeyWithValue("metadata", "value"))
	g.Expect(annotations).To(HaveKeyWithValue("coherence.oracle.com/custom", "keep"))
	g.Expect(annotations).NotTo(HaveKey(coh.AnnotationLastError))
	g.Expect(job.Annotations).To(HaveKeyWithValue("metadata", "value"))
	g.Expect(job.Annotations).To(HaveKeyWithValue("coherence.oracle.com/custom", "keep"))
	g.Expect(job.Annotations).NotTo(HaveKey(coh.AnnotationLastError))
}

func TestCreateJobUsesCoherenceJobAnnotations(t *testing.T) {
	g := NewGomegaWithT(t)

	deployment := createTestCoherenceJobDeployment(coh.CoherenceJobResourceSpec{
		CoherenceResourceSpec: coh.CoherenceResourceSpec{},
		Global: &coh.GlobalSpec{
			Annotations: map[string]string{
				"global": "value",
			},
		},
		JobAnnotations: map[string]string{
			"job":                   "value",
			coh.AnnotationLastError: "do-not-copy",
		},
	})

	job := deployment.Spec.CreateJob(deployment)

	g.Expect(job.Annotations).To(HaveKeyWithValue("global", "value"))
	g.Expect(job.Annotations).To(HaveKeyWithValue("job", "value"))
	g.Expect(job.Annotations).NotTo(HaveKey(coh.AnnotationLastError))
}
