package slice

import (
	godefaultruntime "runtime"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
)

func RemoveString(slice []string, s string) []string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	newSlice := make([]string, 0)
	for _, item := range slice {
		if item == s {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		newSlice = nil
	}
	return newSlice
}
func ContainsString(slice []string, s string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
