/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package statuspatch

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestCreateStatusMergePatchReplacesBloatedConditionsCompactly(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	original := &coh.Coherence{}
	original.Status.Conditions = make(coh.Conditions, 0, 50001)
	for i := 0; i < 50000; i++ {
		original.Status.Conditions = append(original.Status.Conditions, coh.Condition{})
	}
	original.Status.Conditions = append(original.Status.Conditions, coh.Condition{
		Type:   coh.ConditionTypeReady,
		Status: corev1.ConditionTrue,
	})

	updated := original.DeepCopy()
	normalizationDirty := updated.Status.NormalizeStatus()

	data, changed, err := CreateStatusMergePatch(original, updated, normalizationDirty)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(changed).To(gomega.BeTrue())
	g.Expect(len(data)).To(gomega.BeNumerically("<", 1024*1024))

	var patch map[string]map[string]json.RawMessage
	g.Expect(json.Unmarshal(data, &patch)).To(gomega.Succeed())
	var conditions []coh.Condition
	g.Expect(json.Unmarshal(patch["status"]["conditions"], &conditions)).To(gomega.Succeed())
	g.Expect(conditions).To(gomega.Equal([]coh.Condition{
		{Type: coh.ConditionTypeReady, Status: corev1.ConditionTrue},
	}))
}

func TestCreateStatusMergePatchIncludesZeroValueTransitions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	original := &coh.Coherence{}
	original.Status.Replicas = 3
	original.Status.CurrentReplicas = 2
	original.Status.ReadyReplicas = 1
	original.Status.ActionsExecuted = true

	updated := original.DeepCopy()
	updated.Status.Replicas = 0
	updated.Status.CurrentReplicas = 0
	updated.Status.ReadyReplicas = 0
	updated.Status.ActionsExecuted = false

	data, changed, err := CreateStatusMergePatch(original, updated, false)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(changed).To(gomega.BeTrue())

	var patch map[string]map[string]interface{}
	g.Expect(json.Unmarshal(data, &patch)).To(gomega.Succeed())
	g.Expect(patch["status"]).To(gomega.HaveKeyWithValue("replicas", float64(0)))
	g.Expect(patch["status"]).To(gomega.HaveKeyWithValue("currentReplicas", float64(0)))
	g.Expect(patch["status"]).To(gomega.HaveKeyWithValue("readyReplicas", float64(0)))
	g.Expect(patch["status"]).To(gomega.HaveKeyWithValue("actionsExecuted", false))
}
