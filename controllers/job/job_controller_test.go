/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package job

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
	coh "github.com/oracle/coherence-operator/api/v1"
	"github.com/oracle/coherence-operator/controllers/reconciler"
	"github.com/oracle/coherence-operator/pkg/clients"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

func TestUpdateDeploymentStatusPatchesJobStatusAndProbeResults(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctx := context.Background()
	key := types.NamespacedName{Namespace: "default", Name: "test-job"}
	ready := int32(1)
	now := metav1.NewTime(time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	previous := metav1.NewTime(time.Date(2026, 5, 18, 11, 0, 0, 0, time.UTC))

	deployment := &coh.CoherenceJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "coherence.oracle.com/v1",
			Kind:       "CoherenceJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
		Spec: coh.CoherenceJobResourceSpec{
			CoherenceResourceSpec: coh.CoherenceResourceSpec{
				Replicas: ptr.To[int32](2),
			},
		},
	}
	deployment.Status.Conditions = make(coh.Conditions, 0, 2001)
	for i := 0; i < 2000; i++ {
		deployment.Status.Conditions = append(deployment.Status.Conditions, coh.Condition{})
	}
	deployment.Status.Conditions = append(deployment.Status.Conditions, coh.Condition{
		Type:   coh.ConditionTypeReady,
		Status: corev1.ConditionTrue,
	})

	jobResource := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
		Status: batchv1.JobStatus{
			Active:    1,
			Succeeded: 1,
			Ready:     &ready,
		},
	}

	s := runtime.NewScheme()
	g.Expect(clientgoscheme.AddToScheme(s)).To(gomega.Succeed())
	gv := schema.GroupVersion{Group: "coherence.oracle.com", Version: "v1"}
	s.AddKnownTypes(gv, &coh.CoherenceJob{}, &coh.CoherenceJobList{})
	metav1.AddToGroupVersion(s, gv)
	baseClient := fake.NewClientBuilder().
		WithScheme(s).
		WithRuntimeObjects(deployment, jobResource).
		WithStatusSubresource(deployment, jobResource).
		Build()
	statusRecorder := &recordingStatusWriter{SubResourceWriter: baseClient.Status()}
	client := &recordingClient{Client: baseClient, status: statusRecorder}
	mgr := &testManager{client: client, scheme: s}

	controller := &ReconcileJob{
		ReconcileSecondaryResource: reconciler.ReconcileSecondaryResource{
			Kind:     coh.ResourceTypeJob,
			Template: &batchv1.Job{},
		},
	}
	controller.SetCommonReconciler("test-job-controller", mgr, clients.ClientSet{})

	probeStatuses := []coh.CoherenceJobProbeStatus{
		{
			Pod:           "test-job-pod-0",
			LastReadyTime: &previous,
			LastProbeTime: &now,
			Success:       ptr.To(true),
		},
	}

	// Bug39366679/PLAN.md: exercising updateDeploymentStatus directly proves the
	// Job controller persists jobProbes and repairs bloated conditions through the
	// compact status patch path, which is the expected controller-level behavior.
	err := controller.updateDeploymentStatus(ctx, ctrlreconcile.Request{NamespacedName: key}, probeStatuses)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(statusRecorder.patchCount).To(gomega.Equal(1))
	g.Expect(statusRecorder.updateCount).To(gomega.Equal(0))
	g.Expect(statusRecorder.patchType).To(gomega.Equal(types.MergePatchType))
	g.Expect(len(statusRecorder.patchData)).To(gomega.BeNumerically("<", 1024*1024))

	var statusPatch map[string]map[string]json.RawMessage
	g.Expect(json.Unmarshal(statusRecorder.patchData, &statusPatch)).To(gomega.Succeed())
	var patchedProbes []coh.CoherenceJobProbeStatus
	g.Expect(json.Unmarshal(statusPatch["status"]["jobProbes"], &patchedProbes)).To(gomega.Succeed())
	g.Expect(patchedProbes).To(gomega.HaveLen(1))
	g.Expect(patchedProbes[0].Pod).To(gomega.Equal("test-job-pod-0"))
	var patchedConditions coh.Conditions
	g.Expect(json.Unmarshal(statusPatch["status"]["conditions"], &patchedConditions)).To(gomega.Succeed())
	for _, condition := range patchedConditions {
		g.Expect(condition.Type).NotTo(gomega.BeEmpty())
		g.Expect(condition.Status).NotTo(gomega.BeEmpty())
	}

	actual := &coh.CoherenceJob{}
	g.Expect(client.Get(ctx, key, actual)).To(gomega.Succeed())
	g.Expect(actual.Status.Active).To(gomega.Equal(int32(1)))
	g.Expect(actual.Status.Succeeded).To(gomega.Equal(int32(1)))
	g.Expect(actual.Status.CurrentReplicas).To(gomega.Equal(int32(2)))
	g.Expect(actual.Status.ReadyReplicas).To(gomega.Equal(int32(1)))
	g.Expect(actual.Status.JobProbes).To(gomega.HaveLen(1))
	g.Expect(actual.Status.JobProbes[0].Pod).To(gomega.Equal("test-job-pod-0"))
	g.Expect(actual.Status.JobProbes[0].Success).NotTo(gomega.BeNil())
	g.Expect(actual.Status.JobProbes[0].LastReadyTime).NotTo(gomega.BeNil())
	g.Expect(actual.Status.JobProbes[0].LastProbeTime).NotTo(gomega.BeNil())
	g.Expect(*actual.Status.JobProbes[0].Success).To(gomega.BeTrue())
	g.Expect(actual.Status.JobProbes[0].LastReadyTime.Time.Equal(previous.Time)).To(gomega.BeTrue())
	g.Expect(actual.Status.JobProbes[0].LastProbeTime.Time.Equal(now.Time)).To(gomega.BeTrue())
	g.Expect(actual.Status.Conditions).NotTo(gomega.BeEmpty())
	for _, condition := range actual.Status.Conditions {
		g.Expect(condition.Type).NotTo(gomega.BeEmpty())
		g.Expect(condition.Status).NotTo(gomega.BeEmpty())
	}
}

// recordingClient wraps the fake client so this test can prove the controller
// uses the compact status patch path from Bug39366679/PLAN.md and not a full
// status update while still applying the patch to the fake API object.
type recordingClient struct {
	client.Client
	status *recordingStatusWriter
}

func (c *recordingClient) Status() client.StatusWriter {
	return c.status
}

type recordingStatusWriter struct {
	client.SubResourceWriter
	patchCount  int
	updateCount int
	patchType   types.PatchType
	patchData   []byte
}

func (w *recordingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	w.updateCount++
	return w.SubResourceWriter.Update(ctx, obj, opts...)
}

func (w *recordingStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	w.patchCount++
	w.patchType = patch.Type()
	data, err := patch.Data(obj)
	if err != nil {
		return err
	}
	w.patchData = append([]byte(nil), data...)
	return w.SubResourceWriter.Patch(ctx, obj, patch, opts...)
}

// testManager gives SetCommonReconciler a minimal manager backed by the local
// fake client; this keeps the test focused on the Job controller behavior from
// Bug39366679/PLAN.md without importing the repo e2e fake manager path.
type testManager struct {
	client client.Client
	scheme *runtime.Scheme
}

var _ manager.Manager = &testManager{}

func (m *testManager) GetHTTPClient() *http.Client {
	return &http.Client{}
}

func (m *testManager) GetConfig() *rest.Config {
	return &rest.Config{}
}

func (m *testManager) GetCache() cache.Cache {
	panic("not implemented")
}

func (m *testManager) GetScheme() *runtime.Scheme {
	return m.scheme
}

func (m *testManager) GetClient() client.Client {
	return m.client
}

func (m *testManager) GetFieldIndexer() client.FieldIndexer {
	panic("not implemented")
}

func (m *testManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (m *testManager) GetAPIReader() client.Reader {
	return m.client
}

func (m *testManager) Start(context.Context) error {
	return nil
}

func (m *testManager) GetEventRecorderFor(string) record.EventRecorder {
	return record.NewFakeRecorder(1)
}

func (m *testManager) GetEventRecorder(string) events.EventRecorder {
	return nil
}

func (m *testManager) Add(manager.Runnable) error {
	return nil
}

func (m *testManager) Elected() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (m *testManager) AddMetricsServerExtraHandler(string, http.Handler) error {
	return nil
}

func (m *testManager) AddHealthzCheck(string, healthz.Checker) error {
	return nil
}

func (m *testManager) AddReadyzCheck(string, healthz.Checker) error {
	return nil
}

func (m *testManager) GetWebhookServer() webhook.Server {
	return nil
}

func (m *testManager) GetLogger() logr.Logger {
	return logr.Discard()
}

func (m *testManager) GetControllerOptions() config.Controller {
	return config.Controller{}
}

func (m *testManager) GetConverterRegistry() conversion.Registry {
	return nil
}
