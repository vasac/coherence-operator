/*
 * Copyright (c) 2020, 2025, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package certification

import (
	goctx "context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	v1 "github.com/oracle/coherence-operator/api/v1"
	"github.com/oracle/coherence-operator/test/e2e/helper"
	"golang.org/x/net/context"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

type snapshotActionType int

const (
	canaryServiceName = "CanaryService"

	createSnapshot snapshotActionType = iota
	recoverSnapshot
	deleteSnapshot
)

// Deploy a Coherence resource with persistence enabled (this should enable active persistence).
// A PVC should be created for the StatefulSet. Create data in some caches, delete the deployment,
// re-deploy the deployment and assert that the data is recovered.
func TestActivePersistenceScaleUpAndDown(t *testing.T) {
	var yamlFile = "persistence-active-1.yaml"
	var pVolName = "persistence-volume"

	// Ensure that everything is cleaned up after the test!
	testContext.CleanupAfterTest(t)
	g := NewGomegaWithT(t)

	ns := helper.GetTestClusterNamespace()
	deployment, pods := ensurePods(g, yamlFile, ns)

	// check the pvc is created for the given volume
	pvcName := ""
	for _, vol := range pods[0].Spec.Volumes {
		if vol.Name == pVolName {
			if vol.PersistentVolumeClaim != nil {
				pvcName = vol.PersistentVolumeClaim.ClaimName
			}
			break
		}
	}

	// check the pvc is created
	g.Expect(pvcName).NotTo(Equal(""))
	pvc := testContext.KubeClient.CoreV1().PersistentVolumeClaims(pvcName)
	g.Expect(pvc).ShouldNot(BeNil())

	// create data in some caches
	err := helper.StartCanary(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// Start with one replica
	// g.Expect(err).NotTo(HaveOccurred())

	// Scale Up to three
	err = scale(t, ns, deployment.Name, 3)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = helper.WaitForStatefulSet(testContext, ns, deployment.Name, 3, time.Second*10, time.Minute*5)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale down to one
	err = scale(t, ns, deployment.Name, 1)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = helper.WaitForStatefulSet(testContext, ns, deployment.Name, 1, time.Second*10, time.Minute*5)
	g.Expect(err).NotTo(HaveOccurred())

	// assert that the data is recovered
	err = helper.CheckCanaryEventuallyGood(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// cleanup the data
	_ = helper.ClearCanary(testContext, ns, deployment.GetName())
}

// Deploy a Coherence resource with persistence enabled (this should enable active persistence).
// A PVC should be created for the StatefulSet. Create data in some caches, scale down to 0 the deployment,
// re-deploy the deployment and assert that the data is recovered.
func TestActivePersistenceScaleDownAndUp(t *testing.T) {
	var yamlFile = "persistence-active-3.yaml"
	var pVolName = "persistence-volume"

	// Ensure that everything is cleaned up after the test!
	testContext.CleanupAfterTest(t)
	g := NewGomegaWithT(t)

	ns := helper.GetTestClusterNamespace()

	deployment, pods := ensurePods(g, yamlFile, ns)

	// check the pvc is created for the given volume
	pvcName := ""
	for _, vol := range pods[0].Spec.Volumes {
		if vol.Name == pVolName {
			if vol.PersistentVolumeClaim != nil {
				pvcName = vol.PersistentVolumeClaim.ClaimName
			}
			break
		}
	}

	// check the pvc is created
	g.Expect(pvcName).NotTo(Equal(""))
	pvc := testContext.KubeClient.CoreV1().PersistentVolumeClaims(pvcName)
	g.Expect(pvc).ShouldNot(BeNil())

	// create data in some caches
	err := helper.StartCanary(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// Start with three replicas
	// _, err = helper.WaitForStatefulSetForDeployment(f.KubeClient, ns, &deployment, time.Second*10, time.Minute*5, t)
	// g.Expect(err).NotTo(HaveOccurred())

	// Scale Down to One
	err = scale(t, ns, deployment.Name, 1)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = helper.WaitForStatefulSet(testContext, ns, deployment.Name, 1, time.Second*10, time.Minute*5)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale back up to Three
	err = scale(t, ns, deployment.Name, 3)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = helper.WaitForStatefulSet(testContext, ns, deployment.Name, 3, time.Second*10, time.Minute*5)
	g.Expect(err).NotTo(HaveOccurred())

	// assert that the data is recovered
	err = helper.CheckCanaryEventuallyGood(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// cleanup the data
	_ = helper.ClearCanary(testContext, ns, deployment.GetName())
}

// Deploy a Coherence resource with persistence enabled (this should enable active persistence).
// A PVC should be created for the StatefulSet. Create data in some caches, scale down to 0 the deployment,
// re-deploy the deployment and assert that the data is recovered.
func TestActivePersistenceTakeAndRestoreSnapshot(t *testing.T) {
	var yamlFile = "persistence-snapshot.yaml"
	var pVolName = "snapshot-volume"

	// Ensure that everything is cleaned up after the test!
	testContext.CleanupAfterTest(t)
	g := NewGomegaWithT(t)

	ns := helper.GetTestClusterNamespace()

	deployment, pods := ensurePods(g, yamlFile, ns)

	// check the pvc is created for the given volume
	pvcName := ""
	for _, vol := range pods[0].Spec.Volumes {
		if vol.Name == pVolName {
			if vol.PersistentVolumeClaim != nil {
				pvcName = vol.PersistentVolumeClaim.ClaimName
			}
			break
		}
	}

	// check the pvc is created
	g.Expect(pvcName).NotTo(Equal(""))
	pvc := testContext.KubeClient.CoreV1().PersistentVolumeClaims(pvcName)
	g.Expect(pvc).ShouldNot(BeNil())

	// create data in some caches
	err := helper.StartCanary(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// take a snapshot
	snapshotName := "snapshotOne"
	_ = processSnapshotRequest(pods[0], deleteSnapshot, snapshotName)
	err = processSnapshotRequest(pods[0], createSnapshot, snapshotName)
	g.Expect(err).NotTo(HaveOccurred())

	localStorageRestartEnv := os.Getenv("LOCAL_STORAGE_RESTART")
	localStorageRestart, err := strconv.ParseBool(localStorageRestartEnv)
	if err != nil {
		localStorageRestart = false
	}
	// restart Coherence may be on a different instance, local storage will not work
	if !localStorageRestart {
		fmt.Println("Restarting...")
		// dump the pod logs before deleting
		helper.DumpPodsForTest(testContext, t)
		// delete the deployment
		err = helper.WaitForCoherenceCleanup(testContext, ns)
		g.Expect(err).NotTo(HaveOccurred())

		// re-deploy the deployment
		deployment, pods = ensurePods(g, yamlFile, ns)
	}

	// recover the snapshot
	err = processSnapshotRequest(pods[0], recoverSnapshot, snapshotName)
	g.Expect(err).NotTo(HaveOccurred())

	// assert that the data is recovered
	err = helper.CheckCanaryEventuallyGood(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())

	// cleanup the data
	_ = helper.ClearCanary(testContext, ns, deployment.GetName())
}

func ensurePods(g *GomegaWithT, yamlFile, ns string) (v1.Coherence, []corev1.Pod) {
	deployment, err := helper.NewSingleCoherenceFromYaml(ns, yamlFile)
	g.Expect(err).NotTo(HaveOccurred())

	d, _ := json.Marshal(deployment)
	fmt.Printf("Persistence Test installing deployment:\n%s\n", string(d))

	err = testContext.Client.Create(goctx.TODO(), &deployment)
	g.Expect(err).NotTo(HaveOccurred())

	_, err = helper.WaitForStatefulSetForDeployment(testContext, ns, &deployment, helper.RetryInterval, helper.Timeout)
	g.Expect(err).NotTo(HaveOccurred())

	// Get the list of Pods
	pods, err := helper.ListCoherencePodsForDeployment(testContext, ns, deployment.GetName())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(pods)).Should(BeNumerically(">", 0))

	return deployment, pods
}

func processSnapshotRequest(pod corev1.Pod, actionType snapshotActionType, snapshotName string) error {
	pf, ports, err := helper.StartPortForwarderForPodWithBackoff(&pod)
	if err != nil {
		return err
	}

	defer pf.Close()

	url := fmt.Sprintf("http://%s:%d/management/coherence/cluster/services/%s/persistence/snapshots/%s",
		pf.Hostname, ports[v1.PortNameManagement], canaryServiceName, snapshotName)
	httpMethod := "POST"
	if actionType == deleteSnapshot {
		httpMethod = "DELETE"
	}
	if actionType == recoverSnapshot {
		url += "/recover"
	}

	client := &http.Client{}
	var resp *http.Response
	var req *http.Request
	// try a max of 5 times
	for i := 0; i < 5; i++ {
		req, err = http.NewRequest(httpMethod, url, nil)
		if err == nil {
			resp, err = client.Do(req)
			if err == nil {
				break
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
		}
		time.Sleep(5 * time.Second)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("snapshot request returned non-200 status %d", resp.StatusCode)
	}

	// wait for idle
	err = wait.PollUntilContextTimeout(context.Background(), helper.RetryInterval, helper.Timeout, true, func(context.Context) (done bool, err error) {
		url = fmt.Sprintf("http://127.0.0.1:%d/management/coherence/cluster/services/%s/persistence?fields=operationStatus",
			ports[v1.PortNameManagement], canaryServiceName)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Cannot create idle check request: %v\n", url)
			return false, err
		}
		resp, err = client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			fmt.Printf("Error in send idle check request: %v\n", url)
			return false, err
		}
		defer closeBody(resp)

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Idle check request with incorrect status code: %v\n", resp.StatusCode)
			return false, err
		}

		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Print("Error in reading idle check response")
			return false, err
		}

		var data map[string]interface{}
		if err = json.Unmarshal(bs, &data); err != nil {
			fmt.Print("Error in unmarshalling idle check response")
			return false, nil
		}
		opStatus := data["operationStatus"]
		fmt.Printf("Persistence opStatus: %v\n", opStatus)
		if opStatus == "Idle" {
			return true, nil
		}
		return false, nil
	})

	return err
}

func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		testContext.Logger.Error(err, "error closing http response body")
	}
}
