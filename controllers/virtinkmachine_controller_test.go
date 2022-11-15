package controllers

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	virtv1alpha1 "github.com/smartxworks/virtink/pkg/apis/virt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-virtink/api/v1beta1"
)

var _ = Describe("VirtinkMachine controller", func() {
	Context("for a pending VirtinkMachine", func() {
		var clusterKey types.NamespacedName
		var machineKey types.NamespacedName
		var virtinkMachineKey types.NamespacedName
		BeforeEach(func() {
			By("creating a new VirtinkCluster")
			clusterKey = types.NamespacedName{
				Name:      "cluster-" + uuid.New().String(),
				Namespace: "default",
			}

			cluster := capiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterKey.Name,
					Namespace: clusterKey.Namespace,
				},
				Spec: capiv1beta1.ClusterSpec{
					InfrastructureRef: &corev1.ObjectReference{
						Name:      clusterKey.Name,
						Namespace: clusterKey.Namespace,
					},
				},
			}
			Expect(k8sClient.Create(ctx, &cluster)).To(Succeed())

			virtinkCluster := infrastructurev1beta1.VirtinkCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterKey.Name,
					Namespace: clusterKey.Namespace,
				},
				Spec: infrastructurev1beta1.VirtinkClusterSpec{},
			}
			Expect(controllerutil.SetOwnerReference(&cluster, &virtinkCluster, k8sClient.Scheme())).To(Succeed())
			Expect(k8sClient.Create(ctx, &virtinkCluster)).To(Succeed())

			machineKey = types.NamespacedName{
				Name:      "machine-" + uuid.New().String(),
				Namespace: "default",
			}

			machine := capiv1beta1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineKey.Name,
					Namespace: machineKey.Namespace,
				},
				Spec: capiv1beta1.MachineSpec{
					ClusterName: clusterKey.Name,
				},
			}
			Expect(k8sClient.Create(ctx, &machine)).To(Succeed())

			virtinkMachineKey = types.NamespacedName{
				Namespace: "default",
				Name:      "virtink-machine-" + uuid.New().String(),
			}
			virtinkMachine := infrastructurev1beta1.VirtinkMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      virtinkMachineKey.Name,
					Namespace: virtinkMachineKey.Namespace,
					Labels: map[string]string{
						"cluster.x-k8s.io/cluster-name": cluster.Name,
					},
				},
				Spec: infrastructurev1beta1.VirtinkMachineSpec{
					VMSpec: virtv1alpha1.VirtualMachineSpec{
						Instance: virtv1alpha1.Instance{
							CPU: virtv1alpha1.CPU{
								Sockets:        uint32(1),
								CoresPerSocket: uint32(2),
							},
						},
					},
				},
			}
			Expect(controllerutil.SetOwnerReference(&machine, &virtinkMachine, k8sClient.Scheme())).To(Succeed())
			Expect(k8sClient.Create(ctx, &virtinkMachine)).To(Succeed())
		})

		It("should add finalizer", func() {
			var virtinkMachine infrastructurev1beta1.VirtinkMachine
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, virtinkMachineKey, &virtinkMachine)).To(Succeed())
				return controllerutil.ContainsFinalizer(&virtinkMachine, finalizer)
			}).Should(BeTrue())
		})

		Context("when infrastructure cluster is not ready", func() {
			It("should not create VM", func() {
				var vm virtv1alpha1.VirtualMachine
				Consistently(func() bool {
					return apierrors.IsNotFound(k8sClient.Get(ctx, virtinkMachineKey, &vm))
				}).Should(BeTrue())
			})
		})

		Context("when infrastructure cluster is ready", func() {
			BeforeEach(func() {
				var cluster capiv1beta1.Cluster
				Eventually(func() error {
					Expect(k8sClient.Get(ctx, clusterKey, &cluster)).To(Succeed())
					cluster.Status.InfrastructureReady = true
					return k8sClient.Status().Update(ctx, &cluster)
				}).Should(Succeed())
			})

			Context("when bootstrap data secret is not set", func() {
				It("should not create VM", func() {
					var vm virtv1alpha1.VirtualMachine
					Consistently(func() bool {
						return apierrors.IsNotFound(k8sClient.Get(ctx, virtinkMachineKey, &vm))
					}).Should(BeTrue())
				})
			})

			Context("when bootstrap data secret is set", func() {
				BeforeEach(func() {
					var machine capiv1beta1.Machine
					Expect(k8sClient.Get(ctx, machineKey, &machine)).To(Succeed())
					secretName := machine.Name + "-" + "secret"
					secret := corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      secretName,
							Namespace: machine.Namespace,
						},
						StringData: map[string]string{
							"value": "#cloud-init",
						},
					}
					Expect(k8sClient.Create(ctx, &secret)).To(Succeed())

					machine.Spec.Bootstrap.DataSecretName = &secretName
					Expect(k8sClient.Update(ctx, &machine)).To(Succeed())
				})

				It("should create a virtink VM", func() {
					var vm virtv1alpha1.VirtualMachine
					Eventually(func() error {
						return k8sClient.Get(ctx, virtinkMachineKey, &vm)
					}, "10s").Should(Succeed())
					Expect(vm.Spec.Volumes[0].CloudInit.UserDataBase64).To(Equal(base64.StdEncoding.EncodeToString([]byte("#cloud-init"))))

					var virtinkMachine infrastructurev1beta1.VirtinkMachine
					Eventually(func() bool {
						Expect(k8sClient.Get(ctx, machineKey, &virtinkMachine)).To(Succeed())
						return *virtinkMachine.Spec.ProviderID == fmt.Sprintf("virtink://%s", vm.UID) &&
							virtinkMachine.Status.Ready
					})
				})

				Context("when deleting VirtinkMachine", func() {
					It("should delete virtink VM and remove finalizer", func() {
						var vm virtv1alpha1.VirtualMachine
						Eventually(func() error {
							return k8sClient.Get(ctx, virtinkMachineKey, &vm)
						}, "10s").Should(Succeed())

						virtinkMachine := infrastructurev1beta1.VirtinkMachine{
							ObjectMeta: metav1.ObjectMeta{
								Name:      virtinkMachineKey.Name,
								Namespace: virtinkMachineKey.Namespace,
							},
						}
						Expect(k8sClient.Delete(ctx, &virtinkMachine)).To(Succeed())

						Eventually(func() bool {
							return apierrors.IsNotFound(k8sClient.Get(ctx, virtinkMachineKey, &vm))
						}).Should(BeTrue())
					})
				})
			})
		})
	})
})
