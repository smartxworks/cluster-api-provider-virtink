# Kubernetes Cluster API Provider Virtink

[![build](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml/badge.svg)](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartxworks/cluster-api-provider-virtink)](https://goreportcard.com/report/github.com/smartxworks/cluster-api-provider-virtink)

Kubernetes-native declarative infrastructure for [Virtink](https://github.com/smartxworks/virtink).

## What is the Cluster API Provider Virtink

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) brings declarative, Kubernetes-style APIs to cluster creation, configuration and management. Cluster API Provider Virtink is a concrete implementation of Cluster API for Virtink.

The API itself is shared across multiple cloud providers allowing for true Virtink hybrid deployments of Kubernetes. It is built atop the lessons learned from previous cluster managers such as [kops](https://github.com/kubernetes/kops) and [kubicorn](http://kubicorn.io/).

## Deploy Cluster API Provider Virtink

Install clusterctl, see [Install clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start.html#install-clusterctl).

Customize clusterctl provider list, see [Provider repositories](https://cluster-api.sigs.k8s.io/clusterctl/configuration.html#provider-repositories) and add Virtink provider to clusterctl configuration file.

```yaml
providers:
  - name: "virtink"
    url: "https://github.com/smartxworks/cluster-api-provider-virtink/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

Initialize common providers.

```shell
clusterctl init --infrastructure virtink:v0.1.1
```

Create workload cluster template.

```shell
clusterctl generate cluster sample \
  --kubernetes-version v1.24.0 \
  --control-plane-machine-count=3 \
  --worker-machine-count=3 \
  > sample.yaml
```

Virtink cluster OS environment variables.

| Variable name                          | Note                                                       |
|----------------------------------------|------------------------------------------------------------|
| VIRTINK_POD_NETWORK_CIDR               | Virtink workload Kubernetes cluster Pod network CIDR       |
| VIRTINK_SERVICE_CIDR                   | Virtink workload Kubernetes cluster Service network CIDR   |
| VIRTINK_MACHINE_CPU_COUNT              | Virtink VM VCPU count                                      |
| VIRTINK_MACHINE_MEMORY_SIZE            | Virtink VM memory size                                     |
| VIRTINK_MACHINE_KERNEL_IMAGE           | Virtink VM kernel image                                    |
| VIRTINK_MACHINE_ROOTFS_IMAGE           | Virtink VM rootfs image                                    |
| VIRTINK_MACHINE_ROOTFS_SIZE            | Virtink VM rootfs size                                     |
| VIRTINK_INFRA_CLUSTER_SECRET_NAME      | Virtink Infrastructure cluster kubeconfig Secret name      |
| VIRTINK_INFRA_CLUSTER_SECRET_NAMESPACE | Virtink Infrastructure cluster kubeconfig Secret namespace |

Cluster API common variables, see [Common variables](https://cluster-api.sigs.k8s.io/clusterctl/provider-contract.html#common-variables).

## License

This project is distributed under the [Apache License, Version 2.0](LICENSE).
