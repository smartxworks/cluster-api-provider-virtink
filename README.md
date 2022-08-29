# Kubernetes Cluster API Provider Virtink

[![build](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml/badge.svg)](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartxworks/cluster-api-provider-virtink)](https://goreportcard.com/report/github.com/smartxworks/cluster-api-provider-virtink)

Kubernetes-native declarative infrastructure for [Virtink](https://github.com/smartxworks/virtink).

## What is the Cluster API Provider Virtink

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) brings declarative, Kubernetes-style APIs to cluster creation, configuration and management. Cluster API Provider Virtink is a concrete implementation of Cluster API for Virtink.

The API itself is shared across multiple cloud providers allowing for true Virtink hybrid deployments of Kubernetes. It is built atop the lessons learned from previous cluster managers such as [kops](https://github.com/kubernetes/kops) and [kubicorn](http://kubicorn.io/).

## Launching a Kubernetes cluster on Virtink

Check out the [getting started guide](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/main/docs/getting_started.md) for launching a cluster on Virtink.

> **Note**: For `clusterctl` versions prior to v1.2.1, you'll need to add this provider manually to the `clusterctl` configuration file (`$HOME/.cluster-api/clusterctl.yaml`), as shown below:
>
> ```yaml
> providers:
>   - name: "virtink"
>     url: "https://github.com/smartxworks/cluster-api-provider-virtink/releases/latest/infrastructure-components.yaml"
>     type: "InfrastructureProvider"
> ```

## Environment Variables

Except for the [common variables](https://cluster-api.sigs.k8s.io/clusterctl/provider-contract.html#common-variables) provided by Cluster API, you can further customize your workload cluster on Virtink with following environment variables:

| Variable name                              | Note                                                                                                                 |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| VIRTINK_INFRA_CLUSTER_SECRET_NAME          | The name of secret in the management cluster that contains the kubeconfig of the Virtink infrastructure cluster      |
| VIRTINK_INFRA_CLUSTER_SECRET_NAMESPACE     | The namespace of secret in the management cluster that contains the kubeconfig of the Virtink infrastructure cluster |
| POD_NETWORK_CIDR                           | Range of IP addresses for the pod network (default `192.168.0.0/16`)                                                 |
| SERVICE_CIDR                               | Range of IP address for service VIPs (default `10.96.0.0/12`)                                                        |
| VIRTINK_CONTROL_PLANE_SERVICE_TYPE         | The type of control plane service (default `NodePort`)                                                               |
| VIRTINK_CONTROL_PLANE_MACHINE_CPU_CORES    | The CPU cores of each control plane machine (default `2`)                                                            |
| VIRTINK_CONTROL_PLANE_MACHINE_MEMORY_SIZE  | The memory size of each control plane machine (default `4Gi`)                                                        |
| VIRTINK_CONTROL_PLANE_MACHINE_KERNEL_IMAGE | The kernel image of control plane machine (default `smartxworks/capch-kernel-5.15.12`)                               |
| VIRTINK_CONTROL_PLANE_MACHINE_ROOTFS_IMAGE | The rootfs image of control plane machine (default `smartxworks/capch-rootfs-1.24.0`)                                |
| VIRTINK_CONTROL_PLANE_MACHINE_ROOTFS_SIZE  | The rootfs size of each control plane machine (default `4Gi`)                                                        |
| VIRTINK_WORKER_MACHINE_CPU_CORES           | The CPU cores of each worker machine (default `2`)                                                                   |
| VIRTINK_WORKER_MACHINE_MEMORY_SIZE         | The memory size of each worker machine (default `4Gi`)                                                               |
| VIRTINK_WORKER_MACHINE_KERNEL_IMAGE        | The kernel image of worker machine (default `smartxworks/capch-kernel-5.15.12`)                                      |
| VIRTINK_WORKER_MACHINE_ROOTFS_IMAGE        | The rootfs image of worker machine (default `smartxworks/capch-rootfs-1.24.0`)                                       |
| VIRTINK_WORKER_MACHINE_ROOTFS_SIZE         | The rootfs size of each worker machine (default `4Gi`)                                                               |

## License

This project is distributed under the [Apache License, Version 2.0](LICENSE).
