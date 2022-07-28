package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
)

// Test suite constants for e2e config variables.
const (
	KubernetesVersionManagement = "KUBERNETES_VERSION_MANAGEMENT"
	KubernetesVersion           = "KUBERNETES_VERSION"
	CNIPath                     = "CNI"
	CNIResources                = "CNI_RESOURCES"
)

func Byf(format string, a ...interface{}) {
	By(fmt.Sprintf(format, a...))
}
