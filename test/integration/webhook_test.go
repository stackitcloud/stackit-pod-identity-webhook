package integration

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stackitcloud/stackit-pod-identity-webhook/pkg/webhook"
)

var _ = Describe("Pod Identity Webhook Injection", func() {
	var (
		namespace string
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = "default"
	})

	Context("When a Pod is being created", func() {
		It("should inject the environment variables and volumes correctly", func() {
			By("creating a ServiceAccount with workload identity annotation")
			sa := NewTestServiceAccount("test-sa-injection", namespace, map[string]string{
				webhook.AnnotationServiceAccountEmail:           "test-sa@stackit.cloud",
				webhook.AnnotationAudience:                      "custom-audience",
				webhook.AnnotationIDPTokenEndpoint:              "https://custom-idp.stackit.cloud",
				webhook.AnnotationServiceAccountTokenExpiration: "1200",
				webhook.AnnotationFederatedTokenFile:            "/custom/path/token",
			})
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("creating a Pod using that ServiceAccount")
			pod := NewTestPod("test-pod-injection", namespace, sa.Name, nil)
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("verifying the volumes were injected")
			Expect(pod.Spec.Volumes).To(HaveLen(1))
			Expect(pod.Spec.Volumes[0].Name).To(Equal(webhook.DefaultVolumeName))
			Expect(pod.Spec.Volumes[0].Projected).NotTo(BeNil())
			Expect(pod.Spec.Volumes[0].Projected.Sources[0].ServiceAccountToken.Audience).To(Equal("custom-audience"))
			Expect(*pod.Spec.Volumes[0].Projected.Sources[0].ServiceAccountToken.ExpirationSeconds).To(Equal(int64(1200)))

			By("verifying the container mutation")
			container := pod.Spec.Containers[0]
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/custom/path"))

			By("verifying environment variables were injected")
			envMap := make(map[string]string)
			for _, env := range container.Env {
				envMap[env.Name] = env.Value
			}
			Expect(envMap[webhook.EnvServiceAccountEmail]).To(Equal("test-sa@stackit.cloud"))
			Expect(envMap[webhook.EnvIDPTokenEndpoint]).To(Equal("https://custom-idp.stackit.cloud"))
		})

		It("should perform the default mutations only for the sa-email annotation", func() {
			By("creating a ServiceAccount with only the required email annotation")
			sa := NewTestServiceAccount("test-sa-bare-bones", namespace, map[string]string{
				webhook.AnnotationServiceAccountEmail: "test-sa@stackit.cloud",
			})
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("creating a Pod using that ServiceAccount")
			pod := NewTestPod("test-pod-bare-bones", namespace, sa.Name, nil)
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("verifying the volumes were injected with default values")
			Expect(pod.Spec.Volumes).To(HaveLen(1))
			Expect(pod.Spec.Volumes[0].Projected.Sources[0].ServiceAccountToken.Audience).To(Equal(webhook.DefaultAudience))
			Expect(pod.Spec.Volumes[0].Projected.Sources[0].ServiceAccountToken.Path).To(Equal("token"))

			By("verifying the container mutation with default path")
			container := pod.Spec.Containers[0]
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/var/run/secrets/stackit.cloud/serviceaccount"))

			By("verifying only required environment variables were injected")
			envMap := make(map[string]string)
			for _, env := range container.Env {
				envMap[env.Name] = env.Value
			}
			Expect(envMap[webhook.EnvServiceAccountEmail]).To(Equal("test-sa@stackit.cloud"))
			Expect(envMap).NotTo(HaveKey(webhook.EnvIDPTokenEndpoint))
			Expect(envMap).NotTo(HaveKey(webhook.EnvFederatedTokenFile))
		})

		It("should fail mutation when ServiceAccount has invalid token expiration", func() {
			By("creating a ServiceAccount with invalid expiration annotation")
			sa := NewTestServiceAccount("test-sa-invalid-expiration", namespace, map[string]string{
				webhook.AnnotationServiceAccountEmail:           "test@stackit.cloud",
				webhook.AnnotationServiceAccountTokenExpiration: "not-a-number",
			})
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("attempting to create a Pod using that ServiceAccount")
			pod := NewTestPod("test-pod-invalid-expiration", namespace, sa.Name, nil)
			err := k8sClient.Create(ctx, pod)

			By("verifying the creation failed due to webhook error")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not-a-number"))
		})

		It("should not mutate Pod when ServiceAccount has no workload identity annotation", func() {
			By("creating a ServiceAccount without workload identity annotation")
			sa := NewTestServiceAccount("test-sa-no-annotation", namespace, nil)
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("creating a Pod using that ServiceAccount")
			pod := NewTestPod("test-pod-no-annotation", namespace, sa.Name, nil)
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("verifying the Pod was NOT mutated")
			Expect(pod.Spec.Volumes).To(BeEmpty())
		})

		It("should not mutate Pod when skip annotation is set on the Pod", func() {
			By("creating a ServiceAccount with workload identity annotation")
			sa := NewTestServiceAccount("test-sa-skip", namespace, map[string]string{
				webhook.AnnotationServiceAccountEmail: "test@stackit.cloud",
			})
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("creating a Pod with skip annotation")
			pod := NewTestPod("test-pod-skip", namespace, sa.Name, map[string]string{
				webhook.AnnotationSkipWebhook: "true",
			})
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("verifying the Pod was NOT mutated")
			Expect(pod.Spec.Volumes).To(BeEmpty())
		})

		It("should not mutate Pod when skip annotation is set on the ServiceAccount", func() {
			By("creating a ServiceAccount with skip annotation")
			sa := NewTestServiceAccount("test-sa-skip-sa", namespace, map[string]string{
				webhook.AnnotationServiceAccountEmail: "test@stackit.cloud",
				webhook.AnnotationSkipWebhook:         "true",
			})
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			By("creating a Pod using that ServiceAccount")
			pod := NewTestPod("test-pod-skip-sa", namespace, sa.Name, nil)
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("verifying the Pod was NOT mutated")
			Expect(pod.Spec.Volumes).To(BeEmpty())
		})
	})
})
