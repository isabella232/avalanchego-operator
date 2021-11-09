package controllers

import (
	"context"
	"time"

	chainv1alpha1 "github.com/ava-labs/avalanchego-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Avalanchego controller", func() {
	const (
		AvalanchegoValidatorName           = "avalanchego-test-validator"
		AvalanchegoNamespace               = "default"
		AvalanchegoValidatorDeploymentName = "test-validator"

		AvalanchegoKind       = "Avalanchego"
		AvalanchegoAPIVersion = "chain.avax.network/v1alpha1"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Empty bootstrapperURL, genesis and certificates", func() {
		It("Should handle new chain creation", func() {

			spec := chainv1alpha1.AvalanchegoSpec{
				Tag:            "v1.6.3",
				DeploymentName: AvalanchegoValidatorDeploymentName,
				NodeCount:      5,
				Env: []corev1.EnvVar{
					{
						Name:  "AVAGO_LOG_LEVEL",
						Value: "debug",
					},
				},
			}
			key := types.NamespacedName{
				Name:      AvalanchegoValidatorName,
				Namespace: AvalanchegoNamespace,
			}
			toCreate := &chainv1alpha1.Avalanchego{
				TypeMeta: metav1.TypeMeta{
					Kind:       AvalanchegoKind,
					APIVersion: AvalanchegoAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}

			By("Creating Avalanchego chain successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			Eventually(func() bool {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), key, f)
				return f.Status.Error == ""
			}, timeout, interval).Should(BeTrue())

			By("Checking if amount of services created equals nodeCount")

			fetched := &chainv1alpha1.Avalanchego{}

			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), key, fetched)
				return fetched.Spec.NodeCount == len(fetched.Status.NetworkMembersURI)
			}, timeout, interval).Should(BeTrue())

			By("Checking, if genesis was generated")

			Expect(fetched.Status.Genesis).ShouldNot(Equal(""))

			By("Deleting the scope")
			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &chainv1alpha1.Avalanchego{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
