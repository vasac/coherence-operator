/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package reconciler_test

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
	"github.com/oracle/coherence-operator/controllers/reconciler"
	"github.com/oracle/coherence-operator/pkg/clients"
	"github.com/oracle/coherence-operator/pkg/fakes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestUpdateCoherenceJobStatusHashTargetsCoherenceJob(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.Background()
	key := types.NamespacedName{Namespace: "default", Name: "test-job"}

	job := &coh.CoherenceJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "coherence.oracle.com/v1",
			Kind:       "CoherenceJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
	}

	mgr, err := fakes.NewFakeManager(job)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	r := &reconciler.CommonReconciler{}
	r.SetCommonReconciler("test", mgr, clients.ClientSet{})

	g.Expect(r.UpdateCoherenceJobStatusHash(ctx, key, "hash-1")).To(gomega.Succeed())

	actual := &coh.CoherenceJob{}
	g.Expect(mgr.GetClient().Get(ctx, key, actual)).To(gomega.Succeed())
	g.Expect(actual.Status.Hash).To(gomega.Equal("hash-1"))
}
