/*
 * Copyright (c) 2020, 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package errorhandling

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	coh "github.com/oracle/coherence-operator/api/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestOperationError tests the OperationError type
func TestOperationError(t *testing.T) {
	// Create a simple error
	baseErr := errors.New("base error")

	// Create an OperationError
	opErr := NewOperationError("test_operation", baseErr)

	// Test basic properties
	assert.Equal(t, "test_operation", opErr.Operation)
	assert.Equal(t, baseErr, opErr.Err)

	// Test adding context
	_ = opErr.WithContext("key1", "value1").WithContext("key2", "value2")
	assert.Equal(t, "value1", opErr.Context["key1"])
	assert.Equal(t, "value2", opErr.Context["key2"])

	// Test error message formatting
	assert.Contains(t, opErr.Error(), "operation 'test_operation' failed")
	assert.Contains(t, opErr.Error(), "base error")

	// Test Unwrap
	assert.Equal(t, baseErr, opErr.Unwrap())

	// Test Cause (for compatibility with github.com/pkg/errors)
	assert.Equal(t, baseErr, opErr.Cause())
}

// TestResourceError tests the resource-specific error creation
func TestResourceError(t *testing.T) {
	// Create a simple error
	baseErr := errors.New("base error")

	// Create a ResourceError
	resErr := NewResourceError("test_operation", "test-resource", "test-namespace", baseErr)

	// Test basic properties
	assert.Equal(t, "test_operation", resErr.Operation)
	assert.Equal(t, "test-resource", resErr.Resource)
	assert.Equal(t, "test-namespace", resErr.Namespace)
	assert.Equal(t, baseErr, resErr.Err)

	// Test error message formatting
	assert.Contains(t, resErr.Error(), "operation 'test_operation' failed for resource 'test-resource' in namespace 'test-namespace'")
	assert.Contains(t, resErr.Error(), "base error")
}

// TestErrorWrapping tests the error wrapping functions
func TestErrorWrapping(t *testing.T) {
	// Create a simple error
	baseErr := errors.New("base error")

	// Test WrapError
	wrappedErr := WrapError(baseErr, "wrapped message")
	assert.Contains(t, wrappedErr.Error(), "wrapped message: base error")

	// Test WrapErrorf
	wrappedErrf := WrapErrorf(baseErr, "formatted %s", "message")
	assert.Contains(t, wrappedErrf.Error(), "formatted message: base error")

	// Test WithStack
	stackErr := WithStack(baseErr)
	assert.Contains(t, fmt.Sprintf("%+v", stackErr), "base error")
	assert.Contains(t, fmt.Sprintf("%+v", stackErr), "error_handler_test.go")
}

// TestHelperFunctions tests the error helper functions
func TestHelperFunctions(t *testing.T) {
	// Test NewCreateResourceError
	createErr := NewCreateResourceError("test-resource", "test-namespace", errors.New("create error"))
	opErr, ok := createErr.(*OperationError)
	assert.True(t, ok, "Expected *OperationError type")
	assert.Equal(t, "create", opErr.Operation)
	assert.Equal(t, "test-resource", opErr.Resource)
	assert.Equal(t, "test-namespace", opErr.Namespace)
	assert.Contains(t, createErr.Error(), "create error")

	// Test NewUpdateResourceError
	updateErr := NewUpdateResourceError("test-resource", "test-namespace", errors.New("update error"))
	opErr, ok = updateErr.(*OperationError)
	assert.True(t, ok, "Expected *OperationError type")
	assert.Equal(t, "update", opErr.Operation)
	assert.Contains(t, updateErr.Error(), "update error")

	// Test NewDeleteResourceError
	deleteErr := NewDeleteResourceError("test-resource", "test-namespace", errors.New("delete error"))
	opErr, ok = deleteErr.(*OperationError)
	assert.True(t, ok, "Expected *OperationError type")
	assert.Equal(t, "delete", opErr.Operation)
	assert.Contains(t, deleteErr.Error(), "delete error")

	// Test NewGetResourceError
	getErr := NewGetResourceError("test-resource", "test-namespace", errors.New("get error"))
	opErr, ok = getErr.(*OperationError)
	assert.True(t, ok, "Expected *OperationError type")
	assert.Equal(t, "get", opErr.Operation)
	assert.Contains(t, getErr.Error(), "get error")

	// Test NewReconcileError
	reconcileErr := NewReconcileError("test-resource", "test-namespace", errors.New("reconcile error"))
	opErr, ok = reconcileErr.(*OperationError)
	assert.True(t, ok, "Expected *OperationError type")
	assert.Equal(t, "reconcile", opErr.Operation)
	assert.Contains(t, reconcileErr.Error(), "reconcile error")
}

// TestGetCallerInfo tests the GetCallerInfo function
func TestGetCallerInfo(t *testing.T) {
	callerInfo := GetCallerInfo(0)
	assert.Contains(t, callerInfo, "error_handler_test.go")
	assert.Contains(t, callerInfo, ":")
}

func TestBoundPersistedAnnotationPreservesHeadAndTail(t *testing.T) {
	value := "head-" + strings.Repeat("x", persistedAnnotationMaxBytes) + "-tail"

	bounded := boundPersistedAnnotation(value)

	assert.Less(t, len(bounded), persistedAnnotationMaxBytes)
	assert.Contains(t, bounded, "head-")
	assert.Contains(t, bounded, "-tail")
	assert.Contains(t, bounded, "truncated")
}

func TestUpdateStatusRepairsBloatedConditionFixture(t *testing.T) {
	ctx := context.Background()
	key := types.NamespacedName{Namespace: "default", Name: "bloated"}

	deployment := &coh.Coherence{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "coherence.oracle.com/v1",
			Kind:       "Coherence",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
	}
	deployment.Status.Conditions = make(coh.Conditions, 0, 50001)
	for i := 0; i < 50000; i++ {
		deployment.Status.Conditions = append(deployment.Status.Conditions, coh.Condition{})
	}
	deployment.Status.Conditions = append(deployment.Status.Conditions, coh.Condition{
		Type:   coh.ConditionTypeReady,
		Status: corev1.ConditionTrue,
	})

	s := runtime.NewScheme()
	assert.NoError(t, clientgoscheme.AddToScheme(s))
	gv := schema.GroupVersion{Group: "coherence.oracle.com", Version: "v1"}
	s.AddKnownTypes(gv, &coh.Coherence{}, &coh.CoherenceList{})
	metav1.AddToGroupVersion(s, gv)

	client := fake.NewClientBuilder().
		WithScheme(s).
		WithRuntimeObjects(deployment).
		WithStatusSubresource(deployment).
		Build()

	handler := &ErrorHandler{
		Client: client,
		Log:    logr.Discard(),
	}

	// Bug39366679/PLAN.md: this uses a controller-level bloated fixture so the
	// failure-path status write proves it repairs old empty conditions while adding
	// the Failed condition expected for transient reconcile errors.
	assert.NoError(t, handler.updateStatus(ctx, deployment, ErrorCategoryTransient))

	actual := &coh.Coherence{}
	assert.NoError(t, client.Get(ctx, key, actual))
	assert.Less(t, len(actual.Status.Conditions), 10)
	for _, condition := range actual.Status.Conditions {
		assert.NotEmpty(t, condition.Type)
		assert.NotEmpty(t, condition.Status)
	}

	failed := actual.Status.Conditions.GetCondition(coh.ConditionTypeFailed)
	assert.NotNil(t, failed)
	assert.Equal(t, coh.ConditionTypeFailed, actual.Status.Phase)
	assert.Equal(t, corev1.ConditionTrue, failed.Status)
}
