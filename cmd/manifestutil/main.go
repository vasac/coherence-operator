/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package main

import (
	"fmt"
	"github.com/oracle/coherence-operator/pkg/fakes"
	"github.com/oracle/coherence-operator/test/e2e/helper"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
)

// This method is used by the Operator build to generate a yaml manifest that
// is used by the Operator SDK test framework to deploy an Operator. The manifest
// is generated by using the Helm API to run a Helm install of the Operator Helm
// chart with dry-run and debug enabled then capturing the yaml that the install
// would have produced.
func main() {
	var err error

	namespace := helper.GetTestNamespace()
	values := helper.OperatorValues{}
	mgr, err := fakes.NewFakeManager()
	panicIfErr(err)

	chartDir, err := helper.FindOperatorHelmChartDir()
	panicIfErr(err)

	err = values.LoadFromYaml(chartDir + string(os.PathSeparator) + "values.yaml")
	panicIfErr(err)

	helm, err := fakes.NewFakeHelm(mgr, nil, nil, namespace)
	panicIfErr(err)
	result, err := helm.FakeOperatorHelmInstall(mgr, namespace, values)
	panicIfErr(err)

	filterGlobal := func(o runtime.Object) bool {
		kind := o.GetObjectKind().GroupVersionKind().Kind
		return kind == "ClusterRole" || kind == "ClusterRoleBinding"
	}

	filterNamespaced := func(o runtime.Object) bool {
		return !filterGlobal(o)
	}

	filterLocal := func(o runtime.Object) bool {
		kind := o.GetObjectKind().GroupVersionKind().Kind
		return kind != "Deployment" && filterNamespaced(o)
	}

	namespacedName, err := helper.GetTestManifestFileName()
	panicIfErr(err)

	localName, err := helper.GetTestLocalManifestFileName()
	panicIfErr(err)

	globalName, err := helper.GetTestGlobalManifestFileName()
	panicIfErr(err)

	namespacedFile, err := os.Create(namespacedName)
	panicIfErr(err)
	defer closeFile(namespacedFile)

	localFile, err := os.Create(localName)
	panicIfErr(err)
	defer closeFile(localFile)

	globalFile, err := os.Create(globalName)
	panicIfErr(err)
	defer closeFile(globalFile)

	err = result.ToString(filterLocal, localFile)
	panicIfErr(err)

	err = result.ToString(filterNamespaced, namespacedFile)
	panicIfErr(err)

	err = result.ToString(filterGlobal, globalFile)
	panicIfErr(err)
}

func panicIfErr(err error) {
	if err != nil {
		fmt.Println("****** Error:")
		fmt.Println(err)
		panic(err)
	}
}

func closeFile(f *os.File) {
	_ = f.Close()
}
