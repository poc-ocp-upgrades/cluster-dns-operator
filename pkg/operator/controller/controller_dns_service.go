package controller

import (
	"context"
	"fmt"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-dns-operator/pkg/manifests"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *reconciler) ensureDNSService(dns *operatorv1.DNS, clusterIP string, daemonsetRef metav1.OwnerReference) (*corev1.Service, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	current, err := r.currentDNSService(dns)
	if err != nil {
		return nil, err
	}
	if current != nil {
		return current, nil
	}
	desired := desiredDNSService(dns, clusterIP, daemonsetRef)
	if err := r.client.Create(context.TODO(), desired); err != nil {
		return nil, fmt.Errorf("failed to create dns service: %v", err)
	}
	logrus.Infof("created dns service: %s/%s", desired.Namespace, desired.Name)
	return desired, nil
}
func (r *reconciler) currentDNSService(dns *operatorv1.DNS) (*corev1.Service, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	current := &corev1.Service{}
	err := r.client.Get(context.TODO(), DNSServiceName(dns), current)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return current, nil
}
func desiredDNSService(dns *operatorv1.DNS, clusterIP string, daemonsetRef metav1.OwnerReference) *corev1.Service {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s := manifests.DNSService()
	name := DNSServiceName(dns)
	s.Namespace = name.Namespace
	s.Name = name.Name
	s.SetOwnerReferences([]metav1.OwnerReference{dnsOwnerRef(dns)})
	s.Labels = map[string]string{manifests.OwningDNSLabel: DNSDaemonSetLabel(dns)}
	s.Spec.Selector = DNSDaemonSetPodSelector(dns).MatchLabels
	if len(clusterIP) > 0 {
		s.Spec.ClusterIP = clusterIP
	}
	return s
}
