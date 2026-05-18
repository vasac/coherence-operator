/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package metadatapatch

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
)

func TestCreateAnnotationsMergePatchIsMetadataOnly(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	data, changed, err := CreateAnnotationsMergePatch(map[string]string{
		"coherence.oracle.com/last-error": "bounded",
	})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(changed).To(gomega.BeTrue())

	var patch map[string]interface{}
	g.Expect(json.Unmarshal(data, &patch)).To(gomega.Succeed())
	g.Expect(patch).To(gomega.HaveKey("metadata"))
	g.Expect(patch).NotTo(gomega.HaveKey("status"))
}
