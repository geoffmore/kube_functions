package main

import (
	//"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/clientcmd"
	//"path/filepath"
)

// Convert to a consistent unit, please. Otherwise, values will be garbage
// cpuQuota has a real and a imaginary portion
// the first element is the real sum for a kind of cpu quota
// the second element is the count of undefined elements

// all cpu units should be in millicores to avoid working with floating point
// numbers
// type cpuQuota [2]int64
// How is reddis used to cache?
// How can I rearchitect the kubectl touch command?
// if
// c1 = cpuQuota{70, 0}
// c2 = cpuQuota{0, 2}
// c1 has a real count of 70 cores and 0 non-defined cores
// c2 does not have a defined core count for 2 namespaces

// Function names don't make sense yet, but I'm working on that after I clean
// the code

// It was pointed out to me that the value I display as the cluster capacity is
// actually the maximum of defined quotas. If I wanted more accurate metrics of
// what was _actually_ being used, I should calculated based on quota.Status
// rather than quota.Spec. If I wanted to get super fancy with this, I could use
// the metric-server api endpoint and perform calculations such as "how much cpu
// is being used outside of quotas and what percentage of cpu lies outside of
// quotas"

type quotaSum struct {
	cpuLimits                [2]int64
	cpuRequests              [2]int64
	memLimits                [2]int64
	memRequests              [2]int64
	storageRequests          [2]int64
	ephemeralStorageRequests [2]int64
	ephemeralStorageLimits   [2]int64

	quotaCount     int
	namespaceCount int
}

type nodeSum struct {
	cpuLimits       int64
	cpuRequests     int64
	memLimits       int64
	memRequests     int64
	storageRequests int64
}

func main() {
	// Compares current quotas against cluster capacity based on node resources.
	// Intended to be a small piece in measuring cluster capacity
	clientset, err := genClient()
	if err != nil {
		panic(err.Error())
	}
	nodes := getNodes()
	quotaas := getQuotas()
	nodeMetrics := getNodeMetrics(clientset)
	clusterAllocation(nodes)
	clusterCapacity(nodes)
	clusterUtilization(nodeMetrics)
	quotaSpec(quotaas)
	quotaStatus(quotaas)

	// Most of this code is not deprecated in favor of the structure of new.go
	quotas, err := clientset.CoreV1().ResourceQuotas("").List(metav1.ListOptions{})
	quotaList := quotas.Items[:]
	if err != nil {
		panic(err.Error())
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	var clusterQuota quotaSum

	for _, quota := range quotaList {
		quotaSpec := quota.Spec.Hard

		isNil(&clusterQuota.cpuRequests, quotaSpec, v1.ResourceRequestsCPU)
		isNil(&clusterQuota.cpuLimits, quotaSpec, v1.ResourceLimitsCPU)

		isNil(&clusterQuota.memLimits, quotaSpec, v1.ResourceRequestsMemory)
		isNil(&clusterQuota.memRequests, quotaSpec, v1.ResourceLimitsMemory)

		isNil(&clusterQuota.storageRequests, quotaSpec, v1.ResourceRequestsStorage)

		isNil(&clusterQuota.ephemeralStorageRequests, quotaSpec, v1.ResourceRequestsEphemeralStorage)
		isNil(&clusterQuota.ephemeralStorageLimits, quotaSpec, v1.ResourceLimitsEphemeralStorage)

	}

	// https://github.com/kubernetes/client-go/blob/master/kubernetes/typed/core/v1/resourcequota.go
	// https://godoc.org/k8s.io/api/core/v1#ResourceQuota
	clusterQuota.quotaCount = len(quotas.Items)
	clusterQuota.namespaceCount = len(namespaces.Items)

	// At the end here, I should add logic to bring in non-defined values to
	// clusterQuota. Considering that the quotas present already account for what
	// exists for defined quotas, I should be safe adding a constant to each item
	// once that is the difference between namespaceCount and quotaCount
	// also, I have to convert to int64 here
	nonDefinedQuotaNamespaceCount := int64(clusterQuota.namespaceCount - clusterQuota.quotaCount)
	clusterQuota.cpuRequests[1] += nonDefinedQuotaNamespaceCount
	clusterQuota.cpuLimits[1] += nonDefinedQuotaNamespaceCount

	clusterQuota.memLimits[1] += nonDefinedQuotaNamespaceCount
	clusterQuota.memRequests[1] += nonDefinedQuotaNamespaceCount

	clusterQuota.storageRequests[1] += nonDefinedQuotaNamespaceCount

	clusterQuota.ephemeralStorageRequests[1] += nonDefinedQuotaNamespaceCount
	clusterQuota.ephemeralStorageLimits[1] += nonDefinedQuotaNamespaceCount

	//PrettyPrint(clusterQuota)
}

func isNil(vals *[2]int64, l v1.ResourceList, name v1.ResourceName) {
	val, exists := l[name]
	if exists {
		vals[0] += quantityToInt64(val)
	} else {
		vals[1]++
	}
}

func quantityToInt64(c resource.Quantity) (val int64) {
	// c.i and c.d are not accesible because they aren't in the same package
	// How do I tell the difference between a non-defined value and a
	// zero value in golang?
	// -- Since I have a map, I can check if the key exists and get a boolean back.
	//  that logic shouldn't be handled this far down

	var canConvert bool
	val, canConvert = c.AsInt64()
	// Not 100% confident with this logic, but I want to try a conversion to
	// int64, then try a conversion to decimal if that fails. If both fail, I'm
	// probably justified in panicing
	if canConvert == false {
		val, canConvert = c.AsDec().Unscaled()
		if canConvert == false {
			panic("Ya dun goofed")
		}
	}
	return val
}

func allOrInt(num1 [2]int64, num2 int64) (some string) {
	if num1[1] == num2 {
		some = "No defined value in all namespaces"
	} else {
		some = fmt.Sprintf("%v with no defined value in %v namespaces", num1[0], num1[1])
	}
	return some
}
func PrettyPrint(q quotaSum) {

	nsCount := int64(q.namespaceCount)

	fmt.Println("PrettyPrint is halfway decent")
	fmt.Printf("Cluster Stats:\n")
	fmt.Printf(
		"There are %d quotas and %d namespaces\n\n",
		q.quotaCount,
		q.namespaceCount,
	)
	fmt.Printf(
		"CPU Requests: %v\nCPU Limits: %v\nCPU Used: %v\n\n",
		allOrInt(q.cpuRequests, nsCount),
		allOrInt(q.cpuLimits, nsCount),
		"WIP",
	)
	fmt.Printf(
		"Memory Requests: %v\nMemory Limits: %v\nMemory Used: %v\n\n",
		allOrInt(q.memRequests, nsCount),
		allOrInt(q.memLimits, nsCount),
		"WIP",
	)
	fmt.Printf(
		"Ephemeral Storage Requests: %v\nEphemeral Storage Limits %v\nEphemeral Storage Used: %v\n\n",
		allOrInt(q.ephemeralStorageRequests, nsCount),
		allOrInt(q.ephemeralStorageLimits, nsCount),
		"WIP",
	)

	fmt.Printf(
		"Storage Requests: %v\n",
		allOrInt(q.storageRequests, nsCount),
	)
}
