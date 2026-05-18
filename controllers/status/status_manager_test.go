/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package status_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
	"github.com/oracle/coherence-operator/controllers/status"
	"github.com/oracle/coherence-operator/pkg/fakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestUpdateCoherenceStatusPhaseRepairsMissingPhaseCondition(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.Background()
	key := types.NamespacedName{Namespace: "default", Name: "test"}

	deployment := &coh.Coherence{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "coherence.oracle.com/v1",
			Kind:       "Coherence",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
		Status: coh.CoherenceResourceStatus{
			Phase: coh.ConditionTypeStopped,
			Conditions: coh.Conditions{
				{Type: coh.ConditionTypeInitialized, Status: corev1.ConditionTrue},
			},
		},
	}

	mgr, err := fakes.NewFakeManager(deployment)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	sm := &status.StatusManager{
		Client: mgr.GetClient(),
		Log:    logr.Discard(),
	}

	g.Expect(sm.UpdateCoherenceStatusPhase(ctx, key, coh.ConditionTypeStopped)).To(gomega.Succeed())

	actual := &coh.Coherence{}
	g.Expect(mgr.GetClient().Get(ctx, key, actual)).To(gomega.Succeed())
	g.Expect(actual.Status.Phase).To(gomega.Equal(coh.ConditionTypeStopped))

	condition := actual.Status.Conditions.GetCondition(coh.ConditionTypeStopped)
	g.Expect(condition).NotTo(gomega.BeNil())
	g.Expect(condition.Status).To(gomega.Equal(corev1.ConditionTrue))
}
