# Kubernetes Cluster API Provider VirTink

[![build](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml/badge.svg)](https://github.com/smartxworks/cluster-api-provider-virtink/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartxworks/cluster-api-provider-virtink)](https://goreportcard.com/report/github.com/smartxworks/cluster-api-provider-virtink)

Kubernetes-native declarative infrastructure for [VirTink](https://github.com/smartxworks/virtink).

## What is the Cluster API Provider VirTink

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) brings declarative, Kubernetes-style APIs to cluster creation, configuration and management. Cluster API Provider VirTink is a concrete implementation of Cluster API for VirTink.

The API itself is shared across multiple cloud providers allowing for true VirTink hybrid deployments of Kubernetes. It is built atop the lessons learned from previous cluster managers such as [kops](https://github.com/kubernetes/kops) and [kubicorn](http://kubicorn.io/).

## License

This project is distributed under the [Apache License, Version 2.0](LICENSE).
