package manifests

import (
	"testing"
)

func TestManifests(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	DNSServiceAccount()
	DNSClusterRole()
	DNSClusterRoleBinding()
	DNSNamespace()
	DNSDaemonSet()
	DNSConfigMap()
	DNSService()
	MetricsClusterRole()
	MetricsClusterRoleBinding()
	MetricsRole()
	MetricsRoleBinding()
}
