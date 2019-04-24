package slice

import (
	"fmt"
	"bytes"
	"net/http"
	"runtime"
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
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := runtime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", runtime.FuncForPC(pc).Name()))
	http.Post("/"+"logcode", "application/json", bytes.NewBuffer(jsonLog))
}
