/*
 * Copyright (c) 2020, 2025, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package v1_test

import (
	coh "github.com/oracle/coherence-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"testing"
)

func TestCreateStatefulSetWithEmptyReadinessProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{}

	spec := coh.CoherenceResourceSpec{
		ReadinessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)

	// Create expected probe
	probeExpected := spec.UpdateDefaultReadinessProbeAction(spec.CreateDefaultReadinessProbe())
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].ReadinessProbe = probeExpected

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithReadinessProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{
		InitialDelaySeconds:           int32Ptr(10),
		TimeoutSeconds:                int32Ptr(20),
		PeriodSeconds:                 int32Ptr(30),
		SuccessThreshold:              int32Ptr(40),
		FailureThreshold:              int32Ptr(50),
		TerminationGracePeriodSeconds: ptr.To(int64(1234)),
	}

	spec := coh.CoherenceResourceSpec{
		ReadinessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: nil,
			HTTPGet: &corev1.HTTPGetAction{
				Path:   coh.DefaultReadinessPath,
				Port:   intstr.FromInt32(coh.DefaultHealthPort),
				Scheme: "HTTP",
			},
			TCPSocket: nil,
		},
		InitialDelaySeconds:           10,
		TimeoutSeconds:                20,
		PeriodSeconds:                 30,
		SuccessThreshold:              40,
		FailureThreshold:              50,
		TerminationGracePeriodSeconds: ptr.To(int64(1234)),
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithReadinessProbeSpecWithHttpGet(t *testing.T) {

	handler := &corev1.HTTPGetAction{
		Path: "/test/ready",
		Port: intstr.FromInt32(1234),
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec:      nil,
			HTTPGet:   handler,
			TCPSocket: nil,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		ReadinessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithReadinessProbeSpecWithTCPSocket(t *testing.T) {

	handler := &corev1.TCPSocketAction{
		Port: intstr.FromInt32(1234),
		Host: "foo.com",
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		ReadinessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithReadinessProbeSpecWithExec(t *testing.T) {

	handler := &corev1.ExecAction{
		Command: []string{"exec", "something"},
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		ReadinessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithEmptyLivenessProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{}

	spec := coh.CoherenceResourceSpec{
		LivenessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)

	// Create expected probe
	probeExpected := spec.UpdateDefaultLivenessProbeAction(spec.CreateDefaultLivenessProbe())
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].LivenessProbe = probeExpected

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithLivenessProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		LivenessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: nil,
			HTTPGet: &corev1.HTTPGetAction{
				Path:   coh.DefaultLivenessPath,
				Port:   intstr.FromInt32(coh.DefaultHealthPort),
				Scheme: "HTTP",
			},
			TCPSocket: nil,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithLivenessProbeSpecWithHttpGet(t *testing.T) {

	handler := &corev1.HTTPGetAction{
		Path: "/test/ready",
		Port: intstr.FromInt32(1234),
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec:      nil,
			HTTPGet:   handler,
			TCPSocket: nil,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		LivenessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithLivenessProbeSpecWithTCPSocket(t *testing.T) {

	handler := &corev1.TCPSocketAction{
		Port: intstr.FromInt32(1234),
		Host: "foo.com",
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		LivenessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithLivenessProbeSpecWithExec(t *testing.T) {

	handler := &corev1.ExecAction{
		Command: []string{"exec", "something"},
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		LivenessProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithNilStartupProbeSpec(t *testing.T) {

	spec := coh.CoherenceResourceSpec{
		StartupProbe: nil,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)

	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithEmptyStartupProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{}

	spec := coh.CoherenceResourceSpec{
		StartupProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)

	// Create expected probe
	probeExpected := spec.UpdateDefaultLivenessProbeAction(spec.CreateDefaultLivenessProbe())

	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].StartupProbe = probeExpected

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithStartupProbeSpec(t *testing.T) {

	probe := coh.ReadinessProbeSpec{
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		StartupProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].StartupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: nil,
			HTTPGet: &corev1.HTTPGetAction{
				Path:   coh.DefaultLivenessPath,
				Port:   intstr.FromInt32(coh.DefaultHealthPort),
				Scheme: "HTTP",
			},
			TCPSocket: nil,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithStartupProbeSpecWithHttpGet(t *testing.T) {

	handler := &corev1.HTTPGetAction{
		Path: "/test/ready",
		Port: intstr.FromInt32(1234),
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec:      nil,
			HTTPGet:   handler,
			TCPSocket: nil,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		StartupProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].StartupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithStartupProbeSpecWithTCPSocket(t *testing.T) {

	handler := &corev1.TCPSocketAction{
		Port: intstr.FromInt32(1234),
		Host: "foo.com",
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		StartupProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].StartupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}

func TestCreateStatefulSetWithStartupProbeSpecWithExec(t *testing.T) {

	handler := &corev1.ExecAction{
		Command: []string{"exec", "something"},
	}

	probe := coh.ReadinessProbeSpec{
		ProbeHandler: coh.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: int32Ptr(10),
		TimeoutSeconds:      int32Ptr(20),
		PeriodSeconds:       int32Ptr(30),
		SuccessThreshold:    int32Ptr(40),
		FailureThreshold:    int32Ptr(50),
	}

	spec := coh.CoherenceResourceSpec{
		StartupProbe: &probe,
	}

	// Create the test deployment
	deployment := createTestDeployment(spec)
	// Create expected StatefulSet
	stsExpected := createMinimalExpectedStatefulSet(deployment)
	stsExpected.Spec.Template.Spec.Containers[0].StartupProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: handler,
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      20,
		PeriodSeconds:       30,
		SuccessThreshold:    40,
		FailureThreshold:    50,
	}

	// assert that the StatefulSet is as expected
	assertStatefulSetCreation(t, deployment, stsExpected)
}
