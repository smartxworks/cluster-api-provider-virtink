package controllers

import (
	"context"
	"fmt"

	kubridv1alpha1 "github.com/smartxworks/kubrid/pkg/apis/kubrid/v1alpha1"
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

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-kubrid/api/v1beta1"
)

// KubridMachineReconciler reconciles a KubridMachine object
type KubridMachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines/status,verbs=get
//+kubebuilder:rbac:groups=kubrid.smartx.com,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubridMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *KubridMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var machine infrastructurev1beta1.KubridMachine
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

func (r *KubridMachineReconciler) reconcile(ctx context.Context, machine *infrastructurev1beta1.KubridMachine) error {
	if !machine.DeletionTimestamp.IsZero() {
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
	if !ownerCluster.Status.InfrastructureReady {
		return fmt.Errorf("owner Cluster is not ready")
	}

	if ownerMachine.Spec.Bootstrap.DataSecretName == nil {
		return fmt.Errorf("bootstrap data is nil")
	}

	var vm kubridv1alpha1.VirtualMachine
	vmKey := types.NamespacedName{
		Name:      machine.Name,
		Namespace: machine.Namespace,
	}
	vmNotFound := false
	if err := r.Get(ctx, vmKey, &vm); err != nil {
		if apierrors.IsNotFound(err) {
			vmNotFound = true
		} else {
			return fmt.Errorf("get VM: %s", err)
		}
	}

	if !vmNotFound && !metav1.IsControlledBy(&vm, machine) {
		vmNotFound = true
	}

	if vmNotFound {
		vm, err := r.buildVM(ctx, machine, ownerMachine)
		if err != nil {
			return fmt.Errorf("build VM: %s", err)
		}

		vm.Name = vmKey.Name
		vm.Namespace = vmKey.Namespace
		if err := controllerutil.SetControllerReference(machine, vm, r.Scheme); err != nil {
			return fmt.Errorf("set VM controller reference: %s", err)
		}
		if err := r.Create(ctx, vm); err != nil {
			return fmt.Errorf("create VM: %s", err)
		}
		r.Recorder.Eventf(machine, corev1.EventTypeNormal, "CreatedVM", "Create VM %q", vm.Name)
	} else {
		providerID := fmt.Sprintf("kubrid://%s", vm.UID)
		machine.Spec.ProviderID = &providerID
		machine.Status.Ready = true
	}
	return nil
}

func (r *KubridMachineReconciler) buildVM(ctx context.Context, machine *infrastructurev1beta1.KubridMachine, ownerMachine *capiv1beta1.Machine) (*kubridv1alpha1.VirtualMachine, error) {
	vm := &kubridv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      machine.Labels,
			Annotations: machine.Annotations,
		},
		Spec: machine.Spec.VMSpec,
	}
	vm.Spec.Instance.Disks = append(vm.Spec.Instance.Disks, kubridv1alpha1.Disk{
		Name: "cloud-init",
	})
	vm.Spec.Volumes = append(vm.Spec.Volumes, kubridv1alpha1.Volume{
		Name: "cloud-init",
		VolumeSource: kubridv1alpha1.VolumeSource{
			CloudInit: &kubridv1alpha1.CloudInitVolumeSource{
				UserDataSecretName: *ownerMachine.Spec.Bootstrap.DataSecretName,
			},
		},
	})
	return vm, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubridMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.KubridMachine{}).
		Owns(&kubridv1alpha1.VirtualMachine{}).
		Complete(r)
}
