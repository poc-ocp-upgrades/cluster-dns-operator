package controller

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	controllerDaemonSetLabel = "dns.operator.openshift.io/daemonset-dns"
)

func DNSDaemonSetName(dns *operatorv1.DNS) types.NamespacedName {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return types.NamespacedName{Namespace: "openshift-dns", Name: "dns-" + dns.Name}
}
func DNSDaemonSetLabel(dns *operatorv1.DNS) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return dns.Name
}
func DNSDaemonSetPodSelector(dns *operatorv1.DNS) *metav1.LabelSelector {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &metav1.LabelSelector{MatchLabels: map[string]string{controllerDaemonSetLabel: DNSDaemonSetLabel(dns)}}
}
func DNSServiceName(dns *operatorv1.DNS) types.NamespacedName {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return types.NamespacedName{Namespace: "openshift-dns", Name: "dns-" + dns.Name}
}
func DNSConfigMapName(dns *operatorv1.DNS) types.NamespacedName {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return types.NamespacedName{Namespace: "openshift-dns", Name: "dns-" + dns.Name}
}
