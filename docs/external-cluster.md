# Use External Virtink Infrastructure Cluster

External Virtink infrastructure cluster indicates that the Cluster API management and Virtink components are deployed in different Kubernetes clusters, which is the recommended mode. Before referring [quick started guide](https://cluster-api.sigs.k8s.io/user/quick-start.html) for launching a cluster on Virtink, the following requirements should be met.

## Access Virtink Cluster by Kubeconfig

There should be a secret in the management cluster that contains kubeconfig of the Virtink cluster, which will be used by the Cluster API to access the Virtink Cluster. An administor kubeconfig of Virtink cluster can be used for testing, but [Kubernetes RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) is recommended for API objects access control of the Virtink cluster.

Apply the following manifest to Virtink cluster.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: virtink-infra-cluster
rules:
- apiGroups:
  - virt.virtink.smartx.com
  resources:
  - virtualmachines
  verbs:
  - create
  - delete
  - get
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - create
  - delete
  - get
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: virtink-infra-cluster
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: virtink-infra-cluster
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: virtink-infra-cluster
subjects:
  - kind: ServiceAccount
    name: virtink-infra-cluster
    namespace: default
EOF
```

for creating a persistent cluster, should follow [Launching a Kubernetes cluster on Virtink with persistent storage](./../README.md) to make Virtink cluster meets the conditions, and add below rule to virtink-infra-cluster ClusterRole.

```yaml
rules:
- apiGroups:
  - cdi.kubevirt.io
  resources:
  - datavolumes
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```

Create a kubeconfig with API access control of the Virtink cluster.

```shell
# Prepare an administor kubeconfig of the Virtink cluster named virtink-infra-cluster.kubeconfig
kubectl config --kubeconfig virtink-infra-cluster.kubeconfig unset users.kubernetes-admin.client-certificate
kubectl config --kubeconfig virtink-infra-cluster.kubeconfig unset users.kubernetes-admin.client-key
SA_SECRET="$(kubectl get sa virtink-infra-cluster -o jsonpath='{.secrets[0].name}')"
SA_TOKEN="$(kubectl get secret "${SA_SECRET}" -o jsonpath='{.data.token}' | base64 -d)"
kubectl config --kubeconfig virtink-infra-cluster.kubeconfig set-credentials kubernetes-admin "--token=${SA_TOKEN}"
```

> **Note**: In more recent versions, including K8S v1.24, the long term API token will not be automatically created for the ServiceAccount, you may [manually create a token Secret for the ServiceAccount](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#manually-create-a-long-lived-api-token-for-a-serviceaccount), and set the enviroment variable `SA_SECRET` above.

Create a secret in managment cluster and set environment variables before generating workload cluster configuration.

```shell
kubectl create secret generic virtink-infra-cluster --from-file=kubeconfig=virtink-infra-cluster.kubeconfig
export VIRTINK_INFRA_CLUSTER_SECRET_NAME=virtink-infra-cluster
export VIRTINK_INFRA_CLUSTER_SECRET_NAMESPACE=default
```

## LoadBalancer Service Support in Virtink Cluster

The kubeadm control plane controller in management cluster will check nested Kubernetes cluster state by control plane service, a `LoadBalancer` control plane service is required in external Virtink cluster, and its external IP address should be reachable in management cluster. Set environment variable before generating workload cluster configuration.

```shell
export VIRTINK_CONTROL_PLANE_SERVICE_TYPE=LoadBalancer
```

> Without load balancers supports? Try [MetalLB](https://metallb.universe.tf/) for bare-metal clusters.
