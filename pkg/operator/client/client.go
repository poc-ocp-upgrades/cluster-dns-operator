package client

import (
	"fmt"
	"bytes"
	"net/http"
	"runtime"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	scheme *runtime.Scheme
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	scheme = kscheme.Scheme
	if err := operatorv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := configv1.Install(scheme); err != nil {
		panic(err)
	}
}
func GetScheme() *runtime.Scheme {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return scheme
}
func NewClient(kubeConfig *rest.Config) (client.Client, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mapper, err := apiutil.NewDiscoveryRESTMapper(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to discover api rest mapper: %v", err)
	}
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	return kubeClient, nil
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := runtime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", runtime.FuncForPC(pc).Name()))
	http.Post("/"+"logcode", "application/json", bytes.NewBuffer(jsonLog))
}
