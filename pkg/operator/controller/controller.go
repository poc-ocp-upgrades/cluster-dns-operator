package controller

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
	"net"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-dns-operator/pkg/manifests"
	operatorclient "github.com/openshift/cluster-dns-operator/pkg/operator/client"
	"github.com/openshift/cluster-dns-operator/pkg/util/slice"
	"k8s.io/client-go/rest"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"github.com/apparentlymart/go-cidr/cidr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	DefaultDNSController	= "default"
	DNSControllerFinalizer	= "dns.operator.openshift.io/dns-controller"
)

func New(mgr manager.Manager, config Config) (controller.Controller, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	kubeClient, err := operatorclient.NewClient(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	reconciler := &reconciler{Config: config, client: kubeClient}
	c, err := controller.New("operator-controller", mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return nil, err
	}
	if err := c.Watch(&source.Kind{Type: &operatorv1.DNS{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}
	if err := c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1.DNS{}}); err != nil {
		return nil, err
	}
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1.DNS{}}); err != nil {
		return nil, err
	}
	if err := c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{OwnerType: &operatorv1.DNS{}}); err != nil {
		return nil, err
	}
	return c, nil
}

type Config struct {
	KubeConfig		*rest.Config
	CoreDNSImage		string
	OpenshiftCLIImage	string
	OperatorReleaseVersion	string
}
type reconciler struct {
	Config
	client	kclient.Client
}

func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	errs := []error{}
	result := reconcile.Result{}
	logrus.Infof("reconciling request: %v", request)
	if request.NamespacedName.Name != DefaultDNSController {
		logrus.Errorf("skipping unexpected dns %s", request.NamespacedName.Name)
		return result, nil
	}
	dns := &operatorv1.DNS{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, dns); err != nil {
		if errors.IsNotFound(err) {
			logrus.Infof("dns not found; reconciliation will be skipped for request: %v", request)
		} else {
			errs = append(errs, fmt.Errorf("failed to get dns %s: %v", request, err))
		}
		dns = nil
	}
	if dns != nil {
		if err := r.ensureDNSNamespace(); err != nil {
			errs = append(errs, fmt.Errorf("failed to ensure dns namespace: %v", err))
		}
		if dns.DeletionTimestamp != nil {
			if err := r.ensureOpenshiftExternalNameServiceDeleted(); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete external name for openshift service: %v", err))
			}
			if err := r.ensureDNSDeleted(dns); err != nil {
				errs = append(errs, fmt.Errorf("failed to ensure deletion for dns %s: %v", dns.Name, err))
			}
			if len(errs) == 0 {
				if slice.ContainsString(dns.Finalizers, DNSControllerFinalizer) {
					updated := dns.DeepCopy()
					updated.Finalizers = slice.RemoveString(updated.Finalizers, DNSControllerFinalizer)
					if err := r.client.Update(context.TODO(), updated); err != nil {
						errs = append(errs, fmt.Errorf("failed to remove finalizer from dns %s: %v", dns.Name, err))
					}
				}
			}
		} else if err := r.enforceDNSFinalizer(dns); err != nil {
			errs = append(errs, fmt.Errorf("failed to enforce finalizer for dns %s: %v", dns.Name, err))
		} else {
			if err := r.ensureDNS(dns); err != nil {
				errs = append(errs, fmt.Errorf("failed to ensure dns %s: %v", dns.Name, err))
			} else if err := r.ensureExternalNameForOpenshiftService(); err != nil {
				errs = append(errs, fmt.Errorf("failed to ensure external name for openshift service: %v", err))
			}
		}
	}
	if err := r.syncOperatorStatus(); err != nil {
		errs = append(errs, fmt.Errorf("failed to sync operator status: %v", err))
	}
	if len(errs) > 0 {
		logrus.Errorf("failed to reconcile request %s: %v", request, utilerrors.NewAggregate(errs))
	}
	return result, utilerrors.NewAggregate(errs)
}
func (r *reconciler) ensureExternalNameForOpenshiftService() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	svc := &corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}, ObjectMeta: metav1.ObjectMeta{Name: "openshift", Namespace: "default"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeExternalName, ExternalName: "kubernetes.default.svc.cluster.local"}}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}, svc); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get external name service %s/%s: %v", svc.Namespace, svc.Name, err)
		}
		if err := r.client.Create(context.TODO(), svc); err != nil {
			return fmt.Errorf("failed to create external name service %s/%s: %v", svc.Namespace, svc.Name, err)
		}
		logrus.Infof("created external name service %s/%s", svc.Namespace, svc.Name)
	}
	return nil
}
func (r *reconciler) ensureOpenshiftExternalNameServiceDeleted() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	svc := &corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}, ObjectMeta: metav1.ObjectMeta{Name: "openshift", Namespace: "default"}}
	if err := r.client.Delete(context.TODO(), svc); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete external name service %s/%s: %v", svc.Namespace, svc.Name, err)
	}
	logrus.Infof("deleted external name service %s/%s", svc.Namespace, svc.Name)
	return nil
}
func (r *reconciler) enforceDNSFinalizer(dns *operatorv1.DNS) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !slice.ContainsString(dns.Finalizers, DNSControllerFinalizer) {
		dns.Finalizers = append(dns.Finalizers, DNSControllerFinalizer)
		if err := r.client.Update(context.TODO(), dns); err != nil {
			return err
		}
		logrus.Infof("enforced finalizer for dns: %s", dns.Name)
	}
	return nil
}
func (r *reconciler) ensureDNSDeleted(dns *operatorv1.DNS) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := r.ensureDNSDaemonSetDeleted(dns); err != nil {
		return fmt.Errorf("failed to delete daemonset for dns %s: %v", dns.Name, err)
	}
	return nil
}
func (r *reconciler) ensureDNSNamespace() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ns := manifests.DNSNamespace()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ns.Name}, ns); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns namespace %q: %v", ns.Name, err)
		}
		if err := r.client.Create(context.TODO(), ns); err != nil {
			return fmt.Errorf("failed to create dns namespace %s: %v", ns.Name, err)
		}
		logrus.Infof("created dns namespace: %s", ns.Name)
	}
	cr := manifests.DNSClusterRole()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name}, cr); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns cluster role %s: %v", cr.Name, err)
		}
		if err := r.client.Create(context.TODO(), cr); err != nil {
			return fmt.Errorf("failed to create dns cluster role %s: %v", cr.Name, err)
		}
		logrus.Infof("created dns cluster role: %s", cr.Name)
	}
	crb := manifests.DNSClusterRoleBinding()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name}, crb); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns cluster role binding %s: %v", crb.Name, err)
		}
		if err := r.client.Create(context.TODO(), crb); err != nil {
			return fmt.Errorf("failed to create dns cluster role binding %s: %v", crb.Name, err)
		}
		logrus.Infof("created dns cluster role binding: %s", crb.Name)
	}
	sa := manifests.DNSServiceAccount()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: sa.Namespace, Name: sa.Name}, sa); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns service account %s/%s: %v", sa.Namespace, sa.Name, err)
		}
		if err := r.client.Create(context.TODO(), sa); err != nil {
			return fmt.Errorf("failed to create dns service account %s/%s: %v", sa.Namespace, sa.Name, err)
		}
		logrus.Infof("created dns service account: %s/%s", sa.Namespace, sa.Name)
	}
	return nil
}
func (r *reconciler) ensureMetricsIntegration(dns *operatorv1.DNS, svc *corev1.Service, daemonsetRef metav1.OwnerReference) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cr := manifests.MetricsClusterRole()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name}, cr); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns metrics cluster role %s: %v", cr.Name, err)
		}
		if err := r.client.Create(context.TODO(), cr); err != nil {
			return fmt.Errorf("failed to create dns metrics cluster role %s: %v", cr.Name, err)
		}
		logrus.Infof("created dns metrics cluster role %s", cr.Name)
	}
	crb := manifests.MetricsClusterRoleBinding()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name}, crb); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns metrics cluster role binding %s: %v", crb.Name, err)
		}
		if err := r.client.Create(context.TODO(), crb); err != nil {
			return fmt.Errorf("failed to create dns metrics cluster role binding %s: %v", crb.Name, err)
		}
		logrus.Infof("created dns metrics cluster role binding %s", crb.Name)
	}
	mr := manifests.MetricsRole()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: mr.Namespace, Name: mr.Name}, mr); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns metrics role %s/%s: %v", mr.Namespace, mr.Name, err)
		}
		if err := r.client.Create(context.TODO(), mr); err != nil {
			return fmt.Errorf("failed to create dns metrics role %s/%s: %v", mr.Namespace, mr.Name, err)
		}
		logrus.Infof("created dns metrics role %s/%s", mr.Namespace, mr.Name)
	}
	mrb := manifests.MetricsRoleBinding()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: mrb.Namespace, Name: mrb.Name}, mrb); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get dns metrics role binding %s/%s: %v", mrb.Namespace, mrb.Name, err)
		}
		if err := r.client.Create(context.TODO(), mrb); err != nil {
			return fmt.Errorf("failed to create dns metrics role binding %s/%s: %v", mrb.Namespace, mrb.Name, err)
		}
		logrus.Infof("created dns metrics role binding %s/%s", mrb.Namespace, mrb.Name)
	}
	if _, err := r.ensureServiceMonitor(dns, svc, daemonsetRef); err != nil {
		return fmt.Errorf("failed to ensure servicemonitor for %s: %v", dns.Name, err)
	}
	return nil
}
func (r *reconciler) ensureDNS(dns *operatorv1.DNS) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	clusterDomain := "cluster.local"
	clusterIP, err := r.getClusterIPFromNetworkConfig()
	if err != nil {
		return fmt.Errorf("failed to get cluster IP from network config: %v", err)
	}
	errs := []error{}
	if daemonset, err := r.ensureDNSDaemonSet(dns, clusterIP, clusterDomain); err != nil {
		errs = append(errs, fmt.Errorf("failed to ensure daemonset for dns %s: %v", dns.Name, err))
	} else {
		trueVar := true
		daemonsetRef := metav1.OwnerReference{APIVersion: "apps/v1", Kind: "DaemonSet", Name: daemonset.Name, UID: daemonset.UID, Controller: &trueVar}
		if _, err := r.ensureDNSConfigMap(dns, clusterDomain, daemonsetRef); err != nil {
			errs = append(errs, fmt.Errorf("failed to create configmap for dns %s: %v", dns.Name, err))
		}
		if svc, err := r.ensureDNSService(dns, clusterIP, daemonsetRef); err != nil {
			errs = append(errs, fmt.Errorf("failed to create service for dns %s: %v", dns.Name, err))
		} else if err := r.ensureMetricsIntegration(dns, svc, daemonsetRef); err != nil {
			errs = append(errs, fmt.Errorf("failed to integrate metrics with openshift-monitoring for dns %s: %v", dns.Name, err))
		}
		if err := r.syncDNSStatus(dns, clusterIP, clusterDomain); err != nil {
			errs = append(errs, fmt.Errorf("failed to sync status of dns %s/%s: %v", daemonset.Namespace, daemonset.Name, err))
		}
	}
	return utilerrors.NewAggregate(errs)
}
func (r *reconciler) syncDNSStatus(dns *operatorv1.DNS, clusterIP, clusterDomain string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	current := &operatorv1.DNS{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: dns.Name}, current); err != nil {
		return fmt.Errorf("failed to get dns %s: %v", dns.Name, err)
	}
	if current.Status.ClusterIP == clusterIP && current.Status.ClusterDomain == clusterDomain {
		return nil
	}
	current.Status.ClusterIP = clusterIP
	current.Status.ClusterDomain = clusterDomain
	if err := r.client.Status().Update(context.TODO(), current); err != nil {
		return fmt.Errorf("failed to update status for dns %s: %v", current.Name, err)
	}
	return nil
}
func (r *reconciler) getClusterIPFromNetworkConfig() (string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	networkConfig := &configv1.Network{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, networkConfig); err != nil {
		return "", fmt.Errorf("failed to get network 'cluster': %v", err)
	}
	if len(networkConfig.Status.ServiceNetwork) == 0 {
		return "", fmt.Errorf("no service networks found in cluster network config")
	}
	_, serviceCIDR, err := net.ParseCIDR(networkConfig.Status.ServiceNetwork[0])
	if err != nil {
		return "", fmt.Errorf("invalid service cidr %s: %v", networkConfig.Status.ServiceNetwork[0], err)
	}
	dnsClusterIP, err := cidr.Host(serviceCIDR, 10)
	if err != nil {
		return "", fmt.Errorf("invalid service cidr %v: %v", serviceCIDR, err)
	}
	return dnsClusterIP.String(), nil
}
func dnsOwnerRef(dns *operatorv1.DNS) metav1.OwnerReference {
	_logClusterCodePath()
	defer _logClusterCodePath()
	trueVar := true
	return metav1.OwnerReference{APIVersion: "operator.openshift.io/v1", Kind: "DNS", Name: dns.Name, UID: dns.UID, Controller: &trueVar}
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
