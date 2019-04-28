package manifests

import (
	"bytes"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	DNSNamespaceAsset		= "assets/dns/namespace.yaml"
	DNSServiceAccountAsset		= "assets/dns/service-account.yaml"
	DNSClusterRoleAsset		= "assets/dns/cluster-role.yaml"
	DNSClusterRoleBindingAsset	= "assets/dns/cluster-role-binding.yaml"
	DNSConfigMapAsset		= "assets/dns/configmap.yaml"
	DNSDaemonSetAsset		= "assets/dns/daemonset.yaml"
	DNSServiceAsset			= "assets/dns/service.yaml"
	MetricsClusterRoleAsset		= "assets/dns/metrics/cluster-role.yaml"
	MetricsClusterRoleBindingAsset	= "assets/dns/metrics/cluster-role-binding.yaml"
	MetricsRoleAsset		= "assets/dns/metrics/role.yaml"
	MetricsRoleBindingAsset		= "assets/dns/metrics/role-binding.yaml"
	OwningDNSLabel			= "dns.operator.openshift.io/owning-dns"
)

func MustAssetReader(asset string) io.Reader {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return bytes.NewReader(MustAsset(asset))
}
func DNSNamespace() *corev1.Namespace {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ns, err := NewNamespace(MustAssetReader(DNSNamespaceAsset))
	if err != nil {
		panic(err)
	}
	return ns
}
func DNSServiceAccount() *corev1.ServiceAccount {
	_logClusterCodePath()
	defer _logClusterCodePath()
	sa, err := NewServiceAccount(MustAssetReader(DNSServiceAccountAsset))
	if err != nil {
		panic(err)
	}
	return sa
}
func DNSClusterRole() *rbacv1.ClusterRole {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cr, err := NewClusterRole(MustAssetReader(DNSClusterRoleAsset))
	if err != nil {
		panic(err)
	}
	return cr
}
func DNSClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	_logClusterCodePath()
	defer _logClusterCodePath()
	crb, err := NewClusterRoleBinding(MustAssetReader(DNSClusterRoleBindingAsset))
	if err != nil {
		panic(err)
	}
	return crb
}
func DNSConfigMap() *corev1.ConfigMap {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cm, err := NewConfigMap(MustAssetReader(DNSConfigMapAsset))
	if err != nil {
		panic(err)
	}
	return cm
}
func DNSDaemonSet() *appsv1.DaemonSet {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ds, err := NewDaemonSet(MustAssetReader(DNSDaemonSetAsset))
	if err != nil {
		panic(err)
	}
	return ds
}
func DNSService() *corev1.Service {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s, err := NewService(MustAssetReader(DNSServiceAsset))
	if err != nil {
		panic(err)
	}
	return s
}
func MetricsClusterRole() *rbacv1.ClusterRole {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cr, err := NewClusterRole(MustAssetReader(MetricsClusterRoleAsset))
	if err != nil {
		panic(err)
	}
	return cr
}
func MetricsClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	_logClusterCodePath()
	defer _logClusterCodePath()
	crb, err := NewClusterRoleBinding(MustAssetReader(MetricsClusterRoleBindingAsset))
	if err != nil {
		panic(err)
	}
	return crb
}
func MetricsRole() *rbacv1.Role {
	_logClusterCodePath()
	defer _logClusterCodePath()
	r, err := NewRole(MustAssetReader(MetricsRoleAsset))
	if err != nil {
		panic(err)
	}
	return r
}
func MetricsRoleBinding() *rbacv1.RoleBinding {
	_logClusterCodePath()
	defer _logClusterCodePath()
	rb, err := NewRoleBinding(MustAssetReader(MetricsRoleBindingAsset))
	if err != nil {
		panic(err)
	}
	return rb
}
func NewServiceAccount(manifest io.Reader) (*corev1.ServiceAccount, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	sa := corev1.ServiceAccount{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&sa); err != nil {
		return nil, err
	}
	return &sa, nil
}
func NewClusterRole(manifest io.Reader) (*rbacv1.ClusterRole, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cr := rbacv1.ClusterRole{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}
func NewClusterRoleBinding(manifest io.Reader) (*rbacv1.ClusterRoleBinding, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	crb := rbacv1.ClusterRoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&crb); err != nil {
		return nil, err
	}
	return &crb, nil
}
func NewRole(manifest io.Reader) (*rbacv1.Role, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	r := rbacv1.Role{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
func NewRoleBinding(manifest io.Reader) (*rbacv1.RoleBinding, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	rb := rbacv1.RoleBinding{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&rb); err != nil {
		return nil, err
	}
	return &rb, nil
}
func NewConfigMap(manifest io.Reader) (*corev1.ConfigMap, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cm := corev1.ConfigMap{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&cm); err != nil {
		return nil, err
	}
	return &cm, nil
}
func NewDaemonSet(manifest io.Reader) (*appsv1.DaemonSet, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ds := appsv1.DaemonSet{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&ds); err != nil {
		return nil, err
	}
	return &ds, nil
}
func NewService(manifest io.Reader) (*corev1.Service, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s := corev1.Service{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
func NewNamespace(manifest io.Reader) (*corev1.Namespace, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ns := corev1.Namespace{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&ns); err != nil {
		return nil, err
	}
	return &ns, nil
}
