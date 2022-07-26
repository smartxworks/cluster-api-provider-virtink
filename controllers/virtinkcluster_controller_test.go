package controllers

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-virtink/api/v1beta1"
)

var _ = Describe("VirtinkCluster controller", func() {
	Context("for a pending VirtinkCluster", func() {
		var virtinkClusterKey types.NamespacedName
		BeforeEach(func() {
			By("creating a new VirtinkCluster")
			virtinkClusterKey = types.NamespacedName{
				Name:      "cluster-" + uuid.New().String(),
				Namespace: "default",
			}

			cluster := capiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      virtinkClusterKey.Name,
					Namespace: virtinkClusterKey.Namespace,
				},
				Spec: capiv1beta1.ClusterSpec{},
			}
			Expect(k8sClient.Create(ctx, &cluster)).To(Succeed())

			virtinkCluster := infrastructurev1beta1.VirtinkCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      virtinkClusterKey.Name,
					Namespace: virtinkClusterKey.Namespace,
				},
				Spec: infrastructurev1beta1.VirtinkClusterSpec{},
			}
			Expect(k8sClient.Create(ctx, &virtinkCluster)).To(Succeed())
		})

		It("should add finalizer", func() {
			var virtinkCluster infrastructurev1beta1.VirtinkCluster
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, virtinkClusterKey, &virtinkCluster)).To(Succeed())
				return controllerutil.ContainsFinalizer(&virtinkCluster, finalizer)
			}).Should(BeTrue())
		})

		Context("when owner cluster is set", func() {
			BeforeEach(func() {
				var cluster capiv1beta1.Cluster
				Expect(k8sClient.Get(ctx, virtinkClusterKey, &cluster)).To(Succeed())
				var virtinkCluster infrastructurev1beta1.VirtinkCluster
				Eventually(func() error {
					Expect(k8sClient.Get(ctx, virtinkClusterKey, &virtinkCluster)).To(Succeed())
					Expect(controllerutil.SetOwnerReference(&cluster, &virtinkCluster, k8sClient.Scheme())).To(Succeed())
					return k8sClient.Update(ctx, &virtinkCluster)
				}).Should(Succeed())
			})

			It("should create control plane service", func() {
				var virtinkCluster infrastructurev1beta1.VirtinkCluster
				Eventually(func() bool {
					Expect(k8sClient.Get(ctx, virtinkClusterKey, &virtinkCluster)).To(Succeed())
					return controllerutil.ContainsFinalizer(&virtinkCluster, finalizer)
				}).Should(BeTrue())

				var svcKey = types.NamespacedName{Namespace: virtinkCluster.Namespace, Name: virtinkClusterKey.Name}
				var svc corev1.Service
				Eventually(func() error {
					return k8sClient.Get(ctx, svcKey, &svc)
				}).Should(Succeed())
			})
		})
	})

	Context("for a deleting VirtinkCluster", func() {
		var virtinkClusterKey types.NamespacedName
		BeforeEach(func() {
			By("creating a new VirtinkCluster")
			virtinkClusterKey = types.NamespacedName{
				Name:      "cluster-" + uuid.New().String(),
				Namespace: "default",
			}

			cluster := capiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      virtinkClusterKey.Name,
					Namespace: virtinkClusterKey.Namespace,
				},
				Spec: capiv1beta1.ClusterSpec{},
			}
			Expect(k8sClient.Create(ctx, &cluster)).To(Succeed())

			virtinkCluster := infrastructurev1beta1.VirtinkCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      virtinkClusterKey.Name,
					Namespace: virtinkClusterKey.Namespace,
				},
				Spec: infrastructurev1beta1.VirtinkClusterSpec{},
			}
			Expect(controllerutil.SetOwnerReference(&cluster, &virtinkCluster, k8sClient.Scheme())).To(Succeed())
			Expect(k8sClient.Create(ctx, &virtinkCluster)).To(Succeed())

			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, virtinkClusterKey, &virtinkCluster)).To(Succeed())
				return controllerutil.ContainsFinalizer(&virtinkCluster, finalizer)
			}).Should(BeTrue())

			var svcKey = types.NamespacedName{Namespace: "default", Name: virtinkClusterKey.Name}
			var svc corev1.Service
			Eventually(func() error {
				return k8sClient.Get(ctx, svcKey, &svc)
			}).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &virtinkCluster)).To(Succeed())
		})

		It("should delete control plane service and remove finalizer", func() {
			var virtinkCluster infrastructurev1beta1.VirtinkCluster
			Eventually(func() bool {
				return apierrors.IsNotFound(k8sClient.Get(ctx, virtinkClusterKey, &virtinkCluster))
			}).Should(BeTrue())

			var svcKey = types.NamespacedName{Namespace: "default", Name: virtinkClusterKey.Name}
			var svc corev1.Service
			Eventually(func() bool {
				return apierrors.IsNotFound(k8sClient.Get(ctx, svcKey, &svc))
			}).Should(BeTrue())
		})
	})
})
