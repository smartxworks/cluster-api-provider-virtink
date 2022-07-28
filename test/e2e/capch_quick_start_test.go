package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	"k8s.io/utils/pointer"
	capi_e2e "sigs.k8s.io/cluster-api/test/e2e"
)

var _ = Describe("Cluster Creation using Cluster API quick-start test [PR-Blocking]", func() {
	By("Creating single-node control plane with one worker node")
	capi_e2e.QuickStartSpec(context.TODO(), func() capi_e2e.QuickStartSpecInput {
		return capi_e2e.QuickStartSpecInput{
			E2EConfig:             e2eConfig,
			ClusterctlConfigPath:  clusterctlConfigPath,
			BootstrapClusterProxy: bootstrapClusterProxy,
			ArtifactFolder:        artifactFolder,
			SkipCleanup:           skipCleanup,
			Flavor:                pointer.String("internal"),
		}
	})
})
