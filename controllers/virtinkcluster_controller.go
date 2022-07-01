package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiutil "sigs.k8s.io/cluster-api/util"
	capipatch "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1beta1 "github.com/smartxworks/cluster-api-provider-virtink/api/v1beta1"
)

// VirtinkClusterReconciler reconciles a VirtinkCluster object
type VirtinkClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=virtinkclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters/status,verbs=get
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtinkCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *VirtinkClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cluster infrastructurev1beta1.VirtinkCluster
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

func (r *VirtinkClusterReconciler) reconcile(ctx context.Context, cluster *infrastructurev1beta1.VirtinkCluster) error {
	if !controllerutil.ContainsFinalizer(cluster, infrastructurev1beta1.ClusterFinalizer) {
		controllerutil.AddFinalizer(cluster, infrastructurev1beta1.ClusterFinalizer)
		return nil
	}

	infraClusterClient, err := buildInfrastructureClusterClient(ctx, r.Client, cluster.Spec.InfrastructureClusterSecretRef)
	if err != nil {
		return fmt.Errorf("build infrastructure Cluster client from kubeconfig Secret reference: %s", err)
	}

	if !cluster.DeletionTimestamp.IsZero() {
		if err := r.reconcileDeleteWithClient(ctx, cluster, infraClusterClient); err != nil {
			return fmt.Errorf("reconcile delete with client: %s", err)
		}
		return nil
	}

	if err := r.reconcileWithClient(ctx, cluster, infraClusterClient); err != nil {
		return fmt.Errorf("reconcile with client: %s", err)
	}

	return nil
}

func (r *VirtinkClusterReconciler) reconcileDeleteWithClient(ctx context.Context, cluster *infrastructurev1beta1.VirtinkCluster, infraClusterClient client.Client) error {
	var controlPlaneService corev1.Service
	controlPlaneServiceKey := types.NamespacedName{
		Name:      cluster.Name,
		Namespace: cluster.Namespace,
	}
	controlPlaneServiceNotFound := false
	if err := infraClusterClient.Get(ctx, controlPlaneServiceKey, &controlPlaneService); err != nil {
		if apierrors.IsNotFound(err) {
			controlPlaneServiceNotFound = true
		} else {
			return fmt.Errorf("get control plane Service from infrastructure Cluster: %s", err)
		}
	}

	if !controlPlaneServiceNotFound {
		if err := infraClusterClient.Delete(ctx, &controlPlaneService); err != nil {
			return fmt.Errorf("delete control plane Service in infrastructure Cluster: %s", err)
		}
		r.Recorder.Eventf(cluster, corev1.EventTypeNormal, "DeletedControlPlaneService", "Delete control plane Service %q", controlPlaneService.Name)
	}

	controllerutil.RemoveFinalizer(cluster, infrastructurev1beta1.ClusterFinalizer)

	return nil
}

func (r *VirtinkClusterReconciler) reconcileWithClient(ctx context.Context, cluster *infrastructurev1beta1.VirtinkCluster, infraClusterClient client.Client) error {
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
	if err := infraClusterClient.Get(ctx, controlPlaneServiceKey, &controlPlaneService); err != nil {
		if apierrors.IsNotFound(err) {
			controlPlaneServiceNotFound = true
		} else {
			return fmt.Errorf("get control plane Service from infrastructure Cluster: %s", err)
		}
	}

	if controlPlaneServiceNotFound {
		controlPlaneService, err := r.buildControlPlaneService(ctx, cluster, ownerCluster)
		if err != nil {
			return fmt.Errorf("build control plane Service: %s", err)
		}

		controlPlaneService.Name = controlPlaneServiceKey.Name
		controlPlaneService.Namespace = controlPlaneServiceKey.Namespace
		if err := infraClusterClient.Create(ctx, controlPlaneService); err != nil {
			return fmt.Errorf("create control plane Service in infrastructure Cluster: %s", err)
		}
		r.Recorder.Eventf(cluster, corev1.EventTypeNormal, "CreatedControlPlaneService", "Create control plane Service %q", controlPlaneService.Name)
	}

	if cluster.Spec.ControlPlaneServiceType != nil && *cluster.Spec.ControlPlaneServiceType == corev1.ServiceTypeLoadBalancer {
		if len(controlPlaneService.Status.LoadBalancer.Ingress) == 0 {
			return fmt.Errorf("control plane load balancer is not ready")
		}
		cluster.Spec.ControlPlaneEndpoint = capiv1beta1.APIEndpoint{
			Host: controlPlaneService.Status.LoadBalancer.Ingress[0].IP,
			Port: 6443,
		}
	} else {
		cluster.Spec.ControlPlaneEndpoint = capiv1beta1.APIEndpoint{
			Host: controlPlaneService.Spec.ClusterIP,
			Port: 6443,
		}
	}

	cluster.Status.Ready = true
	return nil
}

func (r *VirtinkClusterReconciler) buildControlPlaneService(ctx context.Context, cluster *infrastructurev1beta1.VirtinkCluster, ownerCluster *capiv1beta1.Cluster) (*corev1.Service, error) {
	serviceType := corev1.ServiceTypeNodePort
	if cluster.Spec.ControlPlaneServiceType != nil {
		serviceType = *cluster.Spec.ControlPlaneServiceType
	}
	return &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type: serviceType,
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

func buildInfrastructureClusterClient(ctx context.Context, c client.Client, secretRef *corev1.ObjectReference) (client.Client, error) {
	if secretRef == nil {
		return c, nil
	}

	kubeconfigSecret := corev1.Secret{}
	kubeconfigSecretKey := client.ObjectKey{
		Namespace: secretRef.Namespace,
		Name:      secretRef.Name,
	}
	if err := c.Get(ctx, kubeconfigSecretKey, &kubeconfigSecret); err != nil {
		return nil, fmt.Errorf("get kubeconfig Secret: %s", err)
	}

	kubeConfig, ok := kubeconfigSecret.Data["kubeconfig"]
	if !ok {
		return nil, fmt.Errorf("retrieve kubeconfig from Secret: 'kubeconfig' key is missing")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("create REST config from kubeconfig: %s", err)
	}

	infraClusterClient, err := client.New(restConfig, client.Options{Scheme: c.Scheme()})
	if err != nil {
		return nil, fmt.Errorf("create infrastructure Cluster client: %s", err)
	}
	return infraClusterClient, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtinkClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.VirtinkCluster{}).
		Complete(r)
}
