/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package v1_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNormalizeConditionsDropsEmptyAndDeduplicatesByNewestTransition(t *testing.T) {
	g := NewGomegaWithT(t)

	older := metav1.NewTime(time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC))
	newer := metav1.NewTime(time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC))

	conditions := coh.NormalizeConditions(coh.Conditions{
		{Type: "", Status: corev1.ConditionTrue},
		{Type: coh.ConditionTypeWaiting, Status: ""},
		{Type: coh.ConditionTypeReady, Status: corev1.ConditionFalse, LastTransitionTime: older},
		{Type: coh.ConditionTypeReady, Status: corev1.ConditionTrue, LastTransitionTime: newer, Message: "newest"},
		{Type: coh.ConditionTypeVersioned, Status: corev1.ConditionTrue, LastTransitionTime: newer},
		{Type: coh.ConditionTypeVersioned, Status: corev1.ConditionTrue, LastTransitionTime: newer, Reason: "OperatorVersion", Message: "3.5.11"},
	})

	g.Expect(conditions).To(HaveLen(2))
	g.Expect(conditions[0].Type).To(Equal(coh.ConditionTypeReady))
	g.Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
	g.Expect(conditions[0].Message).To(Equal("newest"))
	g.Expect(conditions[1].Type).To(Equal(coh.ConditionTypeVersioned))
	g.Expect(conditions[1].Reason).To(Equal(coh.ConditionReason("OperatorVersion")))
	g.Expect(conditions[1].Message).To(Equal("3.5.11"))
}

func TestNormalizeStatusReportsDirtyForEmptyConditionBloat(t *testing.T) {
	g := NewGomegaWithT(t)

	status := coh.CoherenceResourceStatus{
		Conditions: coh.Conditions{
			{},
			{Type: coh.ConditionTypeReady, Status: corev1.ConditionTrue},
			{Type: "", Status: ""},
		},
	}

	g.Expect(status.NormalizeStatus()).To(BeTrue())
	g.Expect(status.Conditions).To(Equal(coh.Conditions{
		{Type: coh.ConditionTypeReady, Status: corev1.ConditionTrue},
	}))
	g.Expect(status.NormalizeStatus()).To(BeFalse())
}
