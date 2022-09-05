package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	virtv1alpha1 "github.com/smartxworks/virtink/pkg/apis/virt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiutil "sigs.k8s.io/cluster-api/util"
	capipatch "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-virtink/api/v1beta1"
)

// VirtinkMachineReconciler reconciles a VirtinkMachine object
type VirtinkMachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines/status,verbs=get
//+kubebuilder:rbac:groups=virt.virtink.smartx.com,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtinkMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *VirtinkMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var machine infrastructurev1beta1.VirtinkMachine
	if err := r.Get(ctx, req.NamespacedName, &machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patchHelper, err := capipatch.NewHelper(&machine, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("create Machine patch helper: %s", err)
	}

	if err := r.reconcile(ctx, &machine); err != nil {
		return ctrl.Result{}, err
	}

	if err := patchHelper.Patch(ctx, &machine); err != nil {
		return ctrl.Result{}, fmt.Errorf("patch Machine: %s", err)
	}
	return ctrl.Result{}, nil
}

func (r *VirtinkMachineReconciler) reconcile(ctx context.Context, machine *infrastructurev1beta1.VirtinkMachine) error {
	infraClusterClient := r.Client
	var ownerMachine *capiv1beta1.Machine
	var ownerCluster *capiv1beta1.Cluster
	if controllerutil.ContainsFinalizer(machine, finalizer) {
		m, err := capiutil.GetOwnerMachine(ctx, r.Client, machine.ObjectMeta)
		if err != nil {
			return fmt.Errorf("get owner Machine: %s", err)
		}
		if m == nil {
			return fmt.Errorf("owner Machine is nil")
		}
		ownerMachine = m

		c, err := capiutil.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
		if err != nil {
			return fmt.Errorf("get owner Cluster: %s", err)
		}
		if c == nil {
			return fmt.Errorf("owner Cluster is nil")
		}
		ownerCluster = c

		var cluster infrastructurev1beta1.VirtinkCluster
		clusterKey := types.NamespacedName{
			Name:      ownerCluster.Spec.InfrastructureRef.Name,
			Namespace: ownerCluster.Spec.InfrastructureRef.Namespace,
		}
		if err := r.Get(ctx, clusterKey, &cluster); err != nil {
			return fmt.Errorf("get Cluster: %s", err)
		}

		if cluster.Spec.InfraClusterSecretRef != nil {
			c, err := buildInfraClusterClient(ctx, r.Client, cluster.Spec.InfraClusterSecretRef)
			if err != nil {
				return fmt.Errorf("build infra cluster client: %s", err)
			}
			infraClusterClient = c
		}
	}

	if !machine.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(machine, finalizer) {
			var vm virtv1alpha1.VirtualMachine
			vmKey := types.NamespacedName{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			}
			vmNotFound := false
			if err := infraClusterClient.Get(ctx, vmKey, &vm); err != nil {
				if apierrors.IsNotFound(err) {
					vmNotFound = true
				} else {
					return fmt.Errorf("get VM: %s", err)
				}
			}

			if !vmNotFound {
				if err := infraClusterClient.Delete(ctx, &vm); err != nil {
					return fmt.Errorf("delete VM: %s", err)
				}
				r.Recorder.Eventf(machine, corev1.EventTypeNormal, "DeletedVM", "Deleted VM %q", vm.Name)
			}

			controllerutil.RemoveFinalizer(machine, finalizer)
		}
	} else {
		if !controllerutil.ContainsFinalizer(machine, finalizer) {
			controllerutil.AddFinalizer(machine, finalizer)
			return nil
		}

		if !ownerCluster.Status.InfrastructureReady {
			return fmt.Errorf("owner Cluster is not ready")
		}

		if ownerMachine.Spec.Bootstrap.DataSecretName == nil {
			return fmt.Errorf("bootstrap data is nil")
		}

		dataVolumes := r.buildDataVolumes(ctx, machine)
		for _, dataVolume := range dataVolumes {
			dataVolumeKey := types.NamespacedName{
				Namespace: dataVolume.Namespace,
				Name:      dataVolume.Name,
			}
			dataVolumeNotFound := false
			createdDataVolume := cdiv1beta1.DataVolume{}
			if err := r.Get(ctx, dataVolumeKey, &createdDataVolume); err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("get DataVolume: %s", err)
				}
				dataVolumeNotFound = true
			}
			if dataVolumeNotFound {
				if err := infraClusterClient.Create(ctx, dataVolume); err != nil {
					return fmt.Errorf("create DataVolume: %s", err)
				}
				r.Recorder.Eventf(machine, corev1.EventTypeNormal, "CreatedDataVolume", "Created DataVolume %q", dataVolume.Name)
			}
		}

		var vm virtv1alpha1.VirtualMachine
		vmKey := types.NamespacedName{
			Name:      machine.Name,
			Namespace: machine.Namespace,
		}
		vmNotFound := false
		if err := infraClusterClient.Get(ctx, vmKey, &vm); err != nil {
			if apierrors.IsNotFound(err) {
				vmNotFound = true
			} else {
				return fmt.Errorf("get VM: %s", err)
			}
		}

		vmUID := vm.UID
		if vmNotFound {
			vm, err := r.buildVM(ctx, machine, ownerMachine)
			if err != nil {
				return fmt.Errorf("build VM: %s", err)
			}

			vm.Name = vmKey.Name
			vm.Namespace = vmKey.Namespace
			if err := infraClusterClient.Create(ctx, vm); err != nil {
				return fmt.Errorf("create VM: %s", err)
			}
			r.Recorder.Eventf(machine, corev1.EventTypeNormal, "CreatedVM", "Created VM %q", vm.Name)
			vmUID = vm.UID
		}

		providerID := fmt.Sprintf("virtink://%s", vmUID)
		machine.Spec.ProviderID = &providerID
		machine.Status.Ready = true
	}

	return nil
}

func (r *VirtinkMachineReconciler) buildVM(ctx context.Context, machine *infrastructurev1beta1.VirtinkMachine, ownerMachine *capiv1beta1.Machine) (*virtv1alpha1.VirtualMachine, error) {
	vm := &virtv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      machine.Labels,
			Annotations: machine.Annotations,
		},
		Spec: machine.Spec.VMSpec,
	}

	for i := range vm.Spec.Volumes {
		if vm.Spec.Volumes[i].DataVolume != nil {
			vm.Spec.Volumes[i].DataVolume.VolumeName = fmt.Sprintf("%s-%s", machine.Name, vm.Spec.Volumes[i].DataVolume.VolumeName)
		}
	}

	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      *ownerMachine.Spec.Bootstrap.DataSecretName,
	}
	if err := r.Get(ctx, secretKey, &secret); err != nil {
		return nil, fmt.Errorf("get bootstrap Secret: %s", err)
	}

	vm.Spec.Instance.Disks = append(vm.Spec.Instance.Disks, virtv1alpha1.Disk{
		Name: "cloud-init",
	})
	vm.Spec.Volumes = append(vm.Spec.Volumes, virtv1alpha1.Volume{
		Name: "cloud-init",
		VolumeSource: virtv1alpha1.VolumeSource{
			CloudInit: &virtv1alpha1.CloudInitVolumeSource{
				UserDataBase64: base64.StdEncoding.EncodeToString(secret.Data["value"]),
			},
		},
	})

	var cluster infrastructurev1beta1.VirtinkCluster
	clusterKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      ownerMachine.Spec.ClusterName,
	}
	if err := r.Client.Get(ctx, clusterKey, &cluster); err != nil {
		return nil, err
	}

	if cluster.Spec.NodeAddressConfig == nil {
		return vm, nil
	}

	var machineList infrastructurev1beta1.VirtinkMachineList
	if err := r.Client.List(ctx, &machineList, client.InNamespace(machine.Namespace), client.MatchingLabels{"cluster.x-k8s.io/cluster-name": cluster.Name}); err != nil {
		return nil, err
	}

	var isAddressUsedByMachine = func(m *infrastructurev1beta1.VirtinkMachine, addr string) bool {
		for _, value := range m.Annotations {
			if strings.Contains(value, addr) {
				return true
			}
		}
		return false
	}

	var nextAddress string
	for _, addr := range cluster.Spec.NodeAddressConfig.Addresses {
		var used bool
		for _, virtinkMachine := range machineList.Items {
			if isAddressUsedByMachine(&virtinkMachine, addr) {
				used = true
				break
			}
		}
		if !used {
			nextAddress = addr
			break
		}
	}
	if nextAddress == "" {
		return nil, fmt.Errorf("no more available IP for %q", machine.Name)
	}

	macAddress, err := generateMAC()
	if err != nil {
		return nil, fmt.Errorf("generate MAC address: %s", err)
	}

	for _, anno := range cluster.Spec.NodeAddressConfig.Annotations {
		items := strings.SplitN(anno, "=", 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("invalid annotation template: %s", anno)
		}
		replacer := strings.NewReplacer("$IP_ADDRESS", nextAddress, "$MAC_ADDRESS", macAddress.String())
		vm.Annotations[items[0]] = replacer.Replace(items[1])
	}

	return vm, nil
}

func (r *VirtinkMachineReconciler) buildDataVolumes(ctx context.Context, machine *infrastructurev1beta1.VirtinkMachine) []*cdiv1beta1.DataVolume {
	dataVolumes := []*cdiv1beta1.DataVolume{}
	for _, volume := range machine.Spec.VolumeTemplates {
		switch {
		case volume.DataVolume != nil:
			dataVolume := cdiv1beta1.DataVolume{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: machine.Namespace,
					Name:      fmt.Sprintf("%s-%s", machine.Name, volume.DataVolume.Name),
				},
				Spec: *volume.DataVolume.Spec.DeepCopy(),
			}
			dataVolumes = append(dataVolumes, &dataVolume)
		}
	}
	return dataVolumes
}

func generateMAC() (net.HardwareAddr, error) {
	prefix := []byte{0x52, 0x54, 0x00}
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return nil, fmt.Errorf("rand: %s", err)
	}
	return net.HardwareAddr(append(prefix, suffix...)), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtinkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.VirtinkMachine{}).
		Complete(r)
}
