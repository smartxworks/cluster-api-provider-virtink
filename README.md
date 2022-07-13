# Kubernetes Cluster API Provider Virtink

[![build](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml/badge.svg)](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartxworks/cluster-api-provider-virtink)](https://goreportcard.com/report/github.com/smartxworks/cluster-api-provider-virtink)

Kubernetes-native declarative infrastructure for [Virtink](https://github.com/smartxworks/virtink).

## What is the Cluster API Provider Virtink

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) brings declarative, Kubernetes-style APIs to cluster creation, configuration and management. Cluster API Provider Virtink is a concrete implementation of Cluster API for Virtink.

The API itself is shared across multiple cloud providers allowing for true Virtink hybrid deployments of Kubernetes. It is built atop the lessons learned from previous cluster managers such as [kops](https://github.com/kubernetes/kops) and [kubicorn](http://kubicorn.io/).

## Launching a Kubernetes cluster on Virtink

Check out the [getting started guide](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/main/docs/getting_started.md) for launching a cluster on Virtink. One thing to be noted is that since this project hasn't made to the official Cluster API provider list yet, you'll need to add it manually to the `clusterctl` configuration file (`$HOME/.cluster-api/clusterctl.yaml`), as shown below:

```yaml
providers:
  - name: "virtink"
    url: "https://github.com/smartxworks/cluster-api-provider-virtink/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

## Environment Variables

Except for the [common variables](https://cluster-api.sigs.k8s.io/clusterctl/provider-contract.html#common-variables) provided by Cluster API, you can further customize your workload cluster on Virtink with following environment variables:

| Variable name                          | Note                                                                                                                 |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| VIRTINK_INFRA_CLUSTER_SECRET_NAME      | The name of secret in the management cluster that contains the kubeconfig of the Virtink infrastructure cluster      |
| VIRTINK_INFRA_CLUSTER_SECRET_NAMESPACE | The namespace of secret in the management cluster that contains the kubeconfig of the Virtink infrastructure cluster |
| VIRTINK_POD_NETWORK_CIDR               | The pod network CIDR to use for the workload cluster (default `10.17.0.0/16`)                                        |
| VIRTINK_SERVICE_CIDR                   | The service network CIDR to use for the workload cluster (default `10.112.0.0/12`)                                   |
| VIRTINK_MACHINE_CPU_COUNT              | The number of VM vCPUs (default `2`)                                                                                 |
| VIRTINK_MACHINE_MEMORY_SIZE            | The memory size of VM (default `2Gi`)                                                                                |
| VIRTINK_MACHINE_KERNEL_IMAGE           | The kernel image of VM (default `smartxworks/capch-kernel-5.15.12`)                                                  |
| VIRTINK_MACHINE_ROOTFS_IMAGE           | The rootfs image of VM (default `smartxworks/capch-rootfs-1.24.0`)                                                   |
| VIRTINK_MACHINE_ROOTFS_SIZE            | The rootfs size of VM (default `4Gi`)                                                                                |

## License

This project is distributed under the [Apache License, Version 2.0](LICENSE).
