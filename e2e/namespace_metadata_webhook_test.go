//go:build e2e

// Copyright 2020-2023 Project Capsule Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/projectcapsule/capsule/pkg/api"
)

var _ = Describe("creating a Namespace for a Tenant with additional metadata", func() {
	tnt := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-metadata",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "cap",
					Kind:       "dummy",
					Name:       "tenant-metadata",
					UID:        "tenant-metadata",
				},
			},
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: capsulev1beta2.OwnerListSpec{
				{
					Name: "gatsby",
					Kind: "User",
				},
			},
			NodeSelector: map[string]string{
				"node": "foobar",
			},
			NamespaceOptions: &capsulev1beta2.NamespaceOptions{
				AdditionalMetadata: &api.AdditionalMetadataSpec{
					Labels: map[string]string{
						"k8s.io/custom-label":     "foo",
						"clastix.io/custom-label": "bar",
					},
					Annotations: map[string]string{
						"k8s.io/custom-annotation":     "bizz",
						"clastix.io/custom-annotation": "buzz",
					},
				},
			},
		},
	}

	JustBeforeEach(func() {
		EventuallyCreation(func() error {
			return k8sClient.Create(context.TODO(), tnt)
		}).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), tnt)).Should(Succeed())
	})

	It("should contain additional Namespace metadata", func() {
		ns := NewNamespace("")
		NamespaceCreation(ns, tnt.Spec.Owners[0], defaultTimeoutInterval).Should(Succeed())
		TenantNamespaceList(tnt, defaultTimeoutInterval).Should(ContainElement(ns.GetName()))

		By("checking additional labels", func() {
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, ns)).Should(Succeed())

				for k, v := range tnt.Spec.NamespaceOptions.AdditionalMetadata.Labels {
					Expect(ns.Labels).To(HaveKeyWithValue(k, v))
				}
				return
		})
		By("checking additional annotations", func() {
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, ns)).Should(Succeed())

			for k, v := range tnt.Spec.NamespaceOptions.AdditionalMetadata.Annotations {
				Expect(ns.Annotations).To(HaveKeyWithValue(k, v))
			}
			return
		})
		By("checking namespace node-selector annotation", func() {
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, ns)).Should(Succeed())
			Expect(ns.Annotations).Should(HaveKeyWithValue("scheduler.alpha.kubernetes.io/node-selector", "node=foobar"))
			return
		})
	})
})
