/*
 * Copyright (c) 2026, Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * http://oss.oracle.com/licenses/upl.
 */

package patching

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestPatchFailureMessageIsDigestOnly(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	patchBody := []byte(`{"metadata":{"annotations":{"coherence.oracle.com/last-error":"large patch body"}}}`)
	message := patchFailureMessage("StatefulSet", "test", patchBody)

	g.Expect(message).To(gomega.ContainSubstring("StatefulSet/test"))
	g.Expect(message).To(gomega.ContainSubstring("patch size="))
	g.Expect(message).To(gomega.ContainSubstring("sha256="))
	g.Expect(message).NotTo(gomega.ContainSubstring(string(patchBody)))
	g.Expect(strings.Count(message, "sha256=")).To(gomega.Equal(1))
}
