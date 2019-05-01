package controller

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-dns-operator/pkg/manifests"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	DNSClusterOperatorName = "dns"
)

func (r *reconciler) syncOperatorStatus() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	co := &configv1.ClusterOperator{ObjectMeta: metav1.ObjectMeta{Name: DNSClusterOperatorName}}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: co.Name}, co); err != nil {
		if errors.IsNotFound(err) {
			if err := r.client.Create(context.TODO(), co); err != nil {
				return fmt.Errorf("failed to create clusteroperator %s: %v", co.Name, err)
			}
			logrus.Infof("created clusteroperator %s", co.Name)
		} else {
			return fmt.Errorf("failed to get clusteroperator %s: %v", co.Name, err)
		}
	}
	ns, dnses, daemonsets, err := r.getOperatorState()
	if err != nil {
		return fmt.Errorf("failed to get operator state: %v", err)
	}
	oldStatus := co.Status.DeepCopy()
	co.Status.Conditions = computeStatusConditions(oldStatus.Conditions, ns, dnses, daemonsets)
	co.Status.RelatedObjects = []configv1.ObjectReference{{Resource: "namespaces", Name: "openshift-dns-operator"}, {Resource: "namespaces", Name: ns.Name}}
	if len(r.OperatorReleaseVersion) > 0 {
		for _, condition := range co.Status.Conditions {
			if condition.Type == configv1.OperatorAvailable && condition.Status == configv1.ConditionTrue {
				co.Status.Versions = []configv1.OperandVersion{{Name: "operator", Version: r.OperatorReleaseVersion}, {Name: "coredns", Version: r.CoreDNSImage}, {Name: "openshift-cli", Version: r.OpenshiftCLIImage}}
			}
		}
	}
	if !statusesEqual(*oldStatus, co.Status) {
		if err := r.client.Status().Update(context.TODO(), co); err != nil {
			return fmt.Errorf("failed to update clusteroperator %s: %v", co.Name, err)
		}
	}
	return nil
}
func (r *reconciler) getOperatorState() (*corev1.Namespace, []operatorv1.DNS, []appsv1.DaemonSet, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ns := manifests.DNSNamespace()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, ns); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("failed to get namespace %s: %v", ns.Name, err)
	}
	dnsList := &operatorv1.DNSList{}
	if err := r.client.List(context.TODO(), dnsList); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list dnses: %v", err)
	}
	daemonsetList := &appsv1.DaemonSetList{}
	if err := r.client.List(context.TODO(), daemonsetList, client.InNamespace(ns.Name)); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list daemonsets: %v", err)
	}
	return ns, dnsList.Items, daemonsetList.Items, nil
}
func computeStatusConditions(conditions []configv1.ClusterOperatorStatusCondition, ns *corev1.Namespace, dnses []operatorv1.DNS, daemonsets []appsv1.DaemonSet) []configv1.ClusterOperatorStatusCondition {
	_logClusterCodePath()
	defer _logClusterCodePath()
	failingCondition := &configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorFailing, Status: configv1.ConditionUnknown}
	if ns == nil {
		failingCondition.Status = configv1.ConditionTrue
		failingCondition.Reason = "NoNamespace"
		failingCondition.Message = "DNS namespace does not exist"
	} else {
		failingCondition.Status = configv1.ConditionFalse
	}
	conditions = setStatusCondition(conditions, failingCondition)
	progressingCondition := &configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorProgressing, Status: configv1.ConditionUnknown}
	numDNSes := len(dnses)
	numDaemonSets := len(daemonsets)
	if numDNSes == numDaemonSets {
		progressingCondition.Status = configv1.ConditionFalse
	} else {
		progressingCondition.Status = configv1.ConditionTrue
		progressingCondition.Reason = "Reconciling"
		progressingCondition.Message = fmt.Sprintf("have %d daemonsets, want %d", numDaemonSets, numDNSes)
	}
	conditions = setStatusCondition(conditions, progressingCondition)
	availableCondition := &configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorAvailable, Status: configv1.ConditionUnknown}
	daemonsetsAvailable := map[string]bool{}
	for _, d := range daemonsets {
		daemonsetsAvailable[d.Name] = d.Status.NumberAvailable > 0
	}
	unavailable := []string{}
	for _, dns := range dnses {
		name := DNSDaemonSetName(&dns).Name
		if available, exists := daemonsetsAvailable[name]; !exists {
			msg := fmt.Sprintf("no daemonset for dns %q", dns.Name)
			unavailable = append(unavailable, msg)
		} else if !available {
			msg := fmt.Sprintf("daemonset %q is not available", name)
			unavailable = append(unavailable, msg)
		}
	}
	if len(unavailable) == 0 {
		availableCondition.Status = configv1.ConditionTrue
	} else {
		availableCondition.Status = configv1.ConditionFalse
		availableCondition.Reason = "DaemonSetNotAvailable"
		availableCondition.Message = strings.Join(unavailable, "\n")
	}
	conditions = setStatusCondition(conditions, availableCondition)
	return conditions
}
func setStatusCondition(oldConditions []configv1.ClusterOperatorStatusCondition, condition *configv1.ClusterOperatorStatusCondition) []configv1.ClusterOperatorStatusCondition {
	_logClusterCodePath()
	defer _logClusterCodePath()
	condition.LastTransitionTime = metav1.Now()
	newConditions := []configv1.ClusterOperatorStatusCondition{}
	found := false
	for _, c := range oldConditions {
		if condition.Type == c.Type {
			if condition.Status == c.Status && condition.Reason == c.Reason && condition.Message == c.Message {
				return oldConditions
			}
			found = true
			newConditions = append(newConditions, *condition)
		} else {
			newConditions = append(newConditions, c)
		}
	}
	if !found {
		newConditions = append(newConditions, *condition)
	}
	return newConditions
}
func statusesEqual(a, b configv1.ClusterOperatorStatus) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	conditionCmpOpts := []cmp.Option{cmpopts.IgnoreFields(configv1.ClusterOperatorStatusCondition{}, "LastTransitionTime"), cmpopts.EquateEmpty(), cmpopts.SortSlices(func(a, b configv1.ClusterOperatorStatusCondition) bool {
		return a.Type < b.Type
	})}
	if !cmp.Equal(a.Conditions, b.Conditions, conditionCmpOpts...) {
		return false
	}
	relatedCmpOpts := []cmp.Option{cmpopts.EquateEmpty(), cmpopts.SortSlices(func(a, b configv1.ObjectReference) bool {
		return a.Name < b.Name
	})}
	if !cmp.Equal(a.RelatedObjects, b.RelatedObjects, relatedCmpOpts...) {
		return false
	}
	versionsCmpOpts := []cmp.Option{cmpopts.EquateEmpty(), cmpopts.SortSlices(func(a, b configv1.OperandVersion) bool {
		return a.Name < b.Name
	})}
	if !cmp.Equal(a.Versions, b.Versions, versionsCmpOpts...) {
		return false
	}
	return true
}
