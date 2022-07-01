package controllers

import (
	"context"
	"fmt"

	virtv1alpha1 "github.com/smartxworks/virtink/pkg/apis/virt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
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
	if !controllerutil.ContainsFinalizer(machine, infrastructurev1beta1.MachineFinalizer) {
		controllerutil.AddFinalizer(machine, infrastructurev1beta1.MachineFinalizer)
		return nil
	}

	ownerMachine, err := capiutil.GetOwnerMachine(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		return fmt.Errorf("get owner Machine: %s", err)
	}
	if ownerMachine == nil {
		return fmt.Errorf("owner Machine is nil")
	}

	ownerCluster, err := capiutil.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		return fmt.Errorf("get owner Cluster: %s", err)
	}
	if ownerCluster == nil {
		return fmt.Errorf("owner Cluster is nil")
	}

	var virtinkCluster infrastructurev1beta1.VirtinkCluster
	virtinkClusterKey := types.NamespacedName{
		Namespace: ownerCluster.Spec.InfrastructureRef.Namespace,
		Name:      ownerCluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, virtinkClusterKey, &virtinkCluster); err != nil {
		return fmt.Errorf("get Virtink Cluster: %s", err)
	}

	infraClusterClient, err := buildInfrastructureClusterClient(ctx, r.Client, virtinkCluster.Spec.InfrastructureClusterSecretRef)
	if err != nil {
		return fmt.Errorf("build infrastructure Cluster client from kubeconfig Secret reference: %s", err)
	}

	if !machine.DeletionTimestamp.IsZero() {
		if err := r.reconcileDeleteWithClient(ctx, machine, infraClusterClient); err != nil {
			return fmt.Errorf("reconcile delete with client: %s", err)
		}
		return nil
	}

	if err := r.reconcileWithClient(ctx, machine, ownerMachine, ownerCluster, infraClusterClient); err != nil {
		return fmt.Errorf("reconcile with client: %s", err)
	}

	return nil
}

func (r *VirtinkMachineReconciler) reconcileDeleteWithClient(ctx context.Context, machine *infrastructurev1beta1.VirtinkMachine, infraClusterClient client.Client) error {
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
			return fmt.Errorf("get VM from infrastructure Cluster: %s", err)
		}
	}
	if !vmNotFound {
		if err := infraClusterClient.Delete(ctx, &vm); err != nil {
			return fmt.Errorf("delete VM in infrastructure Cluster: %s", err)
		}
		r.Recorder.Eventf(machine, corev1.EventTypeNormal, "DeletedVM", "Delete VM %q", vm.Name)
	}

	controllerutil.RemoveFinalizer(machine, infrastructurev1beta1.MachineFinalizer)

	return nil
}

func (r *VirtinkMachineReconciler) reconcileWithClient(ctx context.Context, machine *infrastructurev1beta1.VirtinkMachine, ownerMachine *capiv1beta1.Machine, ownerCluster *capiv1beta1.Cluster, infraClusterClient client.Client) error {
	if !ownerCluster.Status.InfrastructureReady {
		return fmt.Errorf("owner Cluster is not ready")
	}

	if ownerMachine.Spec.Bootstrap.DataSecretName == nil {
		return fmt.Errorf("bootstrap data is nil")
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
			return fmt.Errorf("get VM from infrastructure Cluster: %s", err)
		}
	}

	if vmNotFound {
		vm, err := r.buildVM(ctx, machine, ownerMachine)
		if err != nil {
			return fmt.Errorf("build VM: %s", err)
		}

		vm.Name = vmKey.Name
		vm.Namespace = vmKey.Namespace
		if err := infraClusterClient.Create(ctx, vm); err != nil {
			return fmt.Errorf("create VM in infrastructure Cluster: %s", err)
		}
		r.Recorder.Eventf(machine, corev1.EventTypeNormal, "CreatedVM", "Create VM %q", vm.Name)
		return fmt.Errorf("VM %q not found", vm.Name)
	} else {
		providerID := fmt.Sprintf("virtink://%s", vm.UID)
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
				UserData: string(secret.Data["value"]),
			},
		},
	})
	return vm, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtinkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.VirtinkMachine{}).
		Complete(r)
}
