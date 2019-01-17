package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	// "time"

	// "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type quotaSum struct {
	cpuLimits   int64
	cpuRequests int64
	memLimits   int64
	memRequests int64
}

type nodeSum struct {
	cpuLimits   int64
	cpuRequests int64
	memLimits   int64
	memRequests int64
}

func main() {
	// Compares current quotas against cluster capacity based on node resources.
	// Intended to be a small piece in measuring cluster capacity

	// - Boilerplate - //
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	quotas, err := clientset.CoreV1().ResourceQuotas("").List(metav1.ListOptions{})
	quotaList := quotas.Items[:]
	if err != nil {
		panic(err.Error())
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	//var cpuLimitSum int64 = 0
	var clusterQuota quotaSum

	//var limitNum int64
	//var canConvert bool
	//fmt.Println(quotaList)
	for _, quota := range quotaList {
		// I really don't want to import core/v1 just for a few structs
		// I also don't want to use the raw string that is the value
		//quotaSpec[v1.ResourceLimitsCPU.String()],
		//quotaSpec[v1.ResourceRequestsCPU.String()],
		quotaSpec := quota.Spec.Hard
		clusterQuota.cpuLimits += quantityToInt64(quotaSpec[v1.ResourceLimitsCPU])
		clusterQuota.cpuRequests += quantityToInt64(quotaSpec[v1.ResourceRequestsCPU])
		clusterQuota.memLimits += quantityToInt64(quotaSpec[v1.ResourceLimitsMemory])
		clusterQuota.memRequests += quantityToInt64(quotaSpec[v1.ResourceRequestsMemory])
		//cpuLimits := quotaSpec[v1.ResourceLimitsCPU]

		//limitNum, canConvert = cpuLimits.AsInt64()

		//if canConvert == false {
		//	fmt.Println("ya dun goofed. Use AsDec instead")
		//} else {
		//	// Untested default case should return an int64 and a bool
		//	limitNum, canConvert = cpuLimits.AsDec().Unscaled()
		//	if canConvert == false {
		//		panic("Ya dun goofed")
		//	}
		//}

		//cpuLimitSum += limitNum

		//fmt.Printf(
		//	"Cluster CPU Limits: %v\n",
		//	cpuLimitSum,
		//)
		//fmt.Println(quota.String())
	}
	//for i, quota := range quotas.Size() {
	//	fmt.Printf("%T", quota.Spec)
	//}
	// https://github.com/kubernetes/client-go/blob/master/kubernetes/typed/core/v1/resourcequota.go
	// https://godoc.org/k8s.io/api/core/v1#ResourceQuota
	quotaCount := len(quotas.Items)
	namespaceCount := len(namespaces.Items)

	// fmt.Printf("%+v\n", quotas)
	fmt.Printf(
		"There are %d quotas on the cluster and %d namespaces\n",
		quotaCount,
		namespaceCount,
	)
	fmt.Printf(
		"Cluster CPU Requests: %v\nCluster CPU Limits: %v\n",
		clusterQuota.cpuRequests,
		clusterQuota.cpuLimits,
	)
	fmt.Printf(
		"Cluster Mem Requests: %v\nCluster Mem Limits: %v\n",
		clusterQuota.memRequests,
		clusterQuota.memLimits,
	)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func quantityToInt64(c resource.Quantity) int64 {
	// Logic here is poor
	var val int64
	var canConvert bool
	val, canConvert = c.AsInt64()
	if canConvert == false {
		fmt.Println("ya dun goofed. Use AsDec instead")
	} else {
		// Untested default case should return an int64 and a bool
		val, canConvert = c.AsDec().Unscaled()
		if canConvert == false {
			panic("Ya dun goofed")
		}
	}
	return val
}
