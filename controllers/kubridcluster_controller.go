package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiutil "sigs.k8s.io/cluster-api/util"
	capipatch "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-kubrid/api/v1beta1"
)

// KubridClusterReconciler reconciles a KubridCluster object
type KubridClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kubridclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters/status,verbs=get
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubridCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *KubridClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cluster infrastructurev1beta1.KubridCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	patchHelper, err := capipatch.NewHelper(&cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("create Cluster patch helper: %s", err)
	}

	if err := r.reconcile(ctx, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := patchHelper.Patch(ctx, &cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("patch Cluster: %s", err)
	}
	return ctrl.Result{}, nil
}

func (r *KubridClusterReconciler) reconcile(ctx context.Context, cluster *infrastructurev1beta1.KubridCluster) error {
	if !cluster.DeletionTimestamp.IsZero() {
		return nil
	}

	ownerCluster, err := capiutil.GetOwnerCluster(ctx, r.Client, cluster.ObjectMeta)
	if err != nil {
		return fmt.Errorf("get owner Cluster: %s", err)
	}
	if ownerCluster == nil {
		return fmt.Errorf("owner Cluster is nil")
	}

	var controlPlaneService corev1.Service
	controlPlaneServiceKey := types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}
	controlPlaneServiceNotFound := false
	if err := r.Get(ctx, controlPlaneServiceKey, &controlPlaneService); err != nil {
		if apierrors.IsNotFound(err) {
			controlPlaneServiceNotFound = true
		} else {
			return fmt.Errorf("get control plane Service: %s", err)
		}
	}

	if !controlPlaneServiceNotFound && !metav1.IsControlledBy(&controlPlaneService, cluster) {
		controlPlaneServiceNotFound = true
	}

	if controlPlaneServiceNotFound {
		controlPlaneService, err := r.buildControlPlaneService(ctx, cluster, ownerCluster)
		if err != nil {
			return fmt.Errorf("build control plane Service: %s", err)
		}

		controlPlaneService.Name = controlPlaneServiceKey.Name
		controlPlaneService.Namespace = controlPlaneServiceKey.Namespace
		if err := controllerutil.SetControllerReference(cluster, controlPlaneService, r.Scheme); err != nil {
			return fmt.Errorf("set control plane Service controller reference: %s", err)
		}
		if err := r.Create(ctx, controlPlaneService); err != nil {
			return fmt.Errorf("create control plane Service: %s", err)
		}
		r.Recorder.Eventf(cluster, corev1.EventTypeNormal, "CreatedControlPlaneService", "Create control plane Service %q", controlPlaneService.Name)
	}

	cluster.Spec.ControlPlaneEndpoint = capiv1beta1.APIEndpoint{
		Host: controlPlaneService.Spec.ClusterIP,
		Port: 6443,
	}
	cluster.Status.Ready = true
	return nil
}

func (r *KubridClusterReconciler) buildControlPlaneService(ctx context.Context, cluster *infrastructurev1beta1.KubridCluster, ownerCluster *capiv1beta1.Cluster) (*corev1.Service, error) {
	return &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				capiv1beta1.ClusterLabelName:             ownerCluster.Name,
				capiv1beta1.MachineControlPlaneLabelName: "",
			},
			Ports: []corev1.ServicePort{{
				Port:       6443,
				TargetPort: intstr.FromInt(6443),
			}},
		},
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubridClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.KubridCluster{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
