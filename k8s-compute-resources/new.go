package main

// Next step is figuring out how to import the inf package

import (
	"fmt"
	//"gopkg.in/inf.v0"
	"encoding/json"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/rest"
	metricsv1b1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// This function should be in another file
func getNodes() (nodes *v1.NodeList) {
	// Get a list of nodes
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return nodes
}
func getQuotas() (quotas *v1.ResourceQuotaList) {
	quotas, err := clientset.CoreV1().ResourceQuotas("").List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return quotas
}

type cluster struct {
	cpu              *resource.Quantity
	mem              *resource.Quantity
	pods             *resource.Quantity
	storageEphemeral *resource.Quantity
}

func genClusterStruct() cluster {
	return cluster{
		cpu:              resource.NewMilliQuantity(0, resource.DecimalSI),
		mem:              resource.NewQuantity(0, resource.BinarySI),
		pods:             resource.NewQuantity(0, resource.DecimalSI),
		storageEphemeral: resource.NewQuantity(0, resource.DecimalSI),
	}
}

type quota struct {
	// These are the resources I think are important. This is not the complete list
	// of resources. Today, this mirrors the cluster resources in the cluster
	// struct
	cpuLim *resource.Quantity
	cpuReq *resource.Quantity
	memLim *resource.Quantity
	memReq *resource.Quantity
	pods   *resource.Quantity
	//services   *resource.Quantity
	storageLim *resource.Quantity
	storageReq *resource.Quantity
}

func genQuotaStruct() quota {
	return quota{
		cpuLim: resource.NewMilliQuantity(0, resource.DecimalSI),
		cpuReq: resource.NewMilliQuantity(0, resource.DecimalSI),
		memLim: resource.NewMilliQuantity(0, resource.BinarySI),
		memReq: resource.NewMilliQuantity(0, resource.BinarySI),
		pods:   resource.NewQuantity(0, resource.DecimalSI),
		//services:   resource.NewMilliQuantity(0, resource.DecimalSI),
		storageLim: resource.NewMilliQuantity(0, resource.DecimalSI),
		storageReq: resource.NewMilliQuantity(0, resource.DecimalSI),
	}
}

func quotaSpec(quotas *v1.ResourceQuotaList) {
	qt := genQuotaStruct()
	//ql := genQuotaStruct()

	var rl v1.ResourceList

	for _, quota := range quotas.Items {
		rl = quota.Spec.Hard
		//ql := rl[v1.ResourcePods]
		// Need a method to handle this conversion
		// It can be done at the end of summation (as a interface method on the
		// struct)
		// I suppose I should add a node flag and allow for individual nodes to be
		// listed. ql and qt should share a method (since they are the same
		// struct) that formats the fields within the struct
		// Also, add if not nil should be implemented
		// Actually, this metric probably makes sense being separated by
		// namespace instead of node. With that in mind, does it make sense to
		// have a generic verbose flag that includes information not aggregated to
		// the cluster level?
		qt.cpuLim.Add(rl[v1.ResourceLimitsCPU])
		qt.cpuReq.Add(rl[v1.ResourceRequestsCPU])
		qt.memLim.Add(rl[v1.ResourceLimitsMemory])
		qt.memReq.Add(rl[v1.ResourceRequestsMemory])
		qt.pods.Add(rl[v1.ResourcePods])
		qt.storageLim.Add(rl[v1.ResourceRequestsEphemeralStorage])
		qt.storageReq.Add(rl[v1.ResourceLimitsEphemeralStorage])

	}
	fmt.Printf("Quota Spec: \n")
	fmt.Printf("cpu lim: %v\n", qt.cpuLim)
	fmt.Printf("cpu req: %v\n", qt.cpuReq)
	fmt.Printf("mem lim: %v\n", qt.memLim)
	fmt.Printf("mem req: %v\n", qt.memReq)
	fmt.Printf("pods: %v\n", qt.pods)
	fmt.Printf("storage lim: %v\n", qt.storageLim)
	fmt.Printf("storage req: %v\n", qt.storageReq)
	fmt.Printf("\n")
}

func quotaStatus(quotas *v1.ResourceQuotaList) {
	qt := genQuotaStruct()
	//ql := genQuotaStruct()

	var rl v1.ResourceList

	for _, quota := range quotas.Items {
		// This may get us an accurate pod count
		rl = quota.Status.Used
		//ql := rl[v1.ResourcePods]
		// Need a method to handle this conversion
		// It can be done at the end of summation (as a interface method on the
		// struct)
		// I suppose I should add a node flag and allow for individual nodes to be
		// listed. ql and qt should share a method (since they are the same
		// struct) that formats the fields within the struct
		// Also, add if not nil should be implemented
		qt.cpuLim.Add(rl[v1.ResourceLimitsCPU])
		qt.cpuReq.Add(rl[v1.ResourceRequestsCPU])
		qt.memLim.Add(rl[v1.ResourceLimitsMemory])
		qt.memReq.Add(rl[v1.ResourceRequestsMemory])
		qt.pods.Add(rl[v1.ResourcePods])
		qt.storageLim.Add(rl[v1.ResourceRequestsEphemeralStorage])
		qt.storageReq.Add(rl[v1.ResourceLimitsEphemeralStorage])

	}
	fmt.Printf("Quota Status: \n")
	fmt.Printf("cpu lim: %v\n", qt.cpuLim)
	fmt.Printf("cpu req: %v\n", qt.cpuReq)
	fmt.Printf("mem lim: %v\n", qt.memLim)
	fmt.Printf("mem req: %v\n", qt.memReq)
	fmt.Printf("pods: %v\n", qt.pods)
	fmt.Printf("storage lim: %v\n", qt.storageLim)
	fmt.Printf("storage req: %v\n", qt.storageReq)
	fmt.Printf("\n\n")
}
func quotaPrint() {}

// What are the resources I need and do I use the native structs or generate my
// own?

// Node Allocation         (node definitions)
func nodeAllocation() {}

// Node Utilization        (metrics-server)
func nodeUtilization() {}

// Node Reservations       (pod quotas)

func nodeReservation() {}

// Cluster metrics will be the sum of respective node metrics. Should either of
// these be gated by a runtime flag?
// Also, how should these get to json?
// How will they be ECS compliant?
// Should I be able to limit the nodes listed by a set of labels?
// All of these functions need a client
// Don't forget that everything is int64
// Maybe a node list should be collected outside of these individual functions
// and functions should have this node list as a parameter of the signature?
// How do I deal with *resource.Quantity. Helper functions?
// cpu stuff will be in the form of cores as a floating (if possible)

// Cluster Allocatation    (node definitions)
func clusterAllocation(nodes *v1.NodeList) {
	// This part should probably be executed in parallel

	c := genClusterStruct()

	//var capacity v1.ResourceList
	var rl v1.ResourceList
	//var nodeName string
	var cpu, memory, pods, storageEphemeral *resource.Quantity
	// not sure what storage does, so it will stay commented for now
	//var cpu, memory, storage, storageEphemeral, pods *resource.Quantity

	// This should probably be contained within a struct
	// struct should be cluster and contain a list of resources?
	//var clusterCpu *resource.Quantity = resource.NewMilliQuantity(0, resource.DecimalSI)
	//var clusterMem *resource.Quantity = resource.NewQuantity(0, resource.BinarySI)
	//var clusterPods *resource.Quantity = resource.NewQuantity(0, resource.DecimalSI)
	//var clusterStorageEphemeral *resource.Quantity = resource.NewQuantity(0, resource.DecimalSI)

	fmt.Printf("Cluster Allocation: \n")
	fmt.Printf("Nodes: \n")
	for _, node := range nodes.Items {
		//capacity = node.Status.Capacity
		//nodeName = node.ObjectMeta.Name
		rl = node.Status.Allocatable
		cpu = rl.Cpu()
		memory = rl.Memory()
		pods = rl.Pods()
		storageEphemeral = rl.StorageEphemeral()
		// Print resources per node
		// Gating this behind a flag
		//fmt.Printf("name: %v\n", nodeName)
		//fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", cpu, memory, pods, storageEphemeral)
		//fmt.Printf("\n")

		// Add resources to current cluster total
		c.cpu.Add(*cpu)
		c.mem.Add(*memory)
		c.pods.Add(*pods)
		c.storageEphemeral.Add(*storageEphemeral)
	}
	// How can I use the value of cpu.Format rather than the resource.DecimalSI?
	//fmt.Printf("cpus: %v\n", clusterCpu)
	fmt.Printf("Total: \n")
	//memPrint, _ := clusterCpu.AsScale(resource.Giga)
	fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", c.cpu, c.mem, c.pods, c.storageEphemeral)
	fmt.Printf("\n\n")
	// Execute a function that adds node stats to a running total, prints node
	// info
}

// Cluster Capacity (node definitions)
func clusterCapacity(nodes *v1.NodeList) {
	// This part should probably be executed in parallel
	c := genClusterStruct()

	//var capacity v1.ResourceList
	var rl v1.ResourceList
	//var nodeName string
	var cpu, memory, pods, storageEphemeral *resource.Quantity
	// not sure what storage does, so it will stay commented for now
	//var cpu, memory, storage, storageEphemeral, pods *resource.Quantity

	// This should probably be contained within a struct
	// struct should be cluster and contain a list of resources?
	//var clusterCpu *resource.Quantity = resource.NewMilliQuantity(0, resource.DecimalSI)
	//var clusterMem *resource.Quantity = resource.NewQuantity(0, resource.BinarySI)
	//var clusterPods *resource.Quantity = resource.NewQuantity(0, resource.DecimalSI)
	//var clusterStorageEphemeral *resource.Quantity = resource.NewQuantity(0, resource.DecimalSI)

	fmt.Printf("Cluster Capacity: \n")
	fmt.Printf("Nodes: \n")
	for _, node := range nodes.Items {
		//capacity = node.Status.Capacity
		//nodeName = node.ObjectMeta.Name
		rl = node.Status.Capacity
		cpu = rl.Cpu()
		memory = rl.Memory()
		pods = rl.Pods()
		storageEphemeral = rl.StorageEphemeral()
		// Print resources per node
		// Gating this behind a flag
		//fmt.Printf("name: %v\n", nodeName)
		//fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", cpu, memory, pods, storageEphemeral)
		//fmt.Printf("\n")

		// Add resources to current cluster total
		c.cpu.Add(*cpu)
		c.mem.Add(*memory)
		c.pods.Add(*pods)
		c.storageEphemeral.Add(*storageEphemeral)
	}
	// How can I use the value of cpu.Format rather than the resource.DecimalSI?
	//fmt.Printf("cpus: %v\n", clusterCpu)
	fmt.Printf("Total: \n")
	//memPrint, _ := clusterCpu.AsScale(resource.Giga)
	fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", c.cpu, c.mem, c.pods, c.storageEphemeral)
	fmt.Printf("\n\n")
	// Execute a function that adds node stats to a running total, prints node
	// info
}

func capacityAllocationPrint() {
	// Similar to quotaPrint
}

// Cluster Utilization     (metrics-server)
func getNodeMetrics(clientset *kubernetes.Clientset) (nodeMetricList *metricsv1b1.NodeMetricsList) {
	//data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/nodes").Do()
	data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/nodes").DoRaw()
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &nodeMetricList); err != nil {
		panic(err)
	}
	//fmt.Printf("%+v", nodeMetricList.Items)
	return nodeMetricList
}

func clusterUtilization(nodeMetricList *metricsv1b1.NodeMetricsList) {
	// https://stackoverflow.com/questions/52029656/how-to-retrieve-kubernetes-metrics-via-client-go-and-golang
	// This part should probably be executed in parallel

	c := genClusterStruct()

	//var capacity v1.ResourceList
	var rl v1.ResourceList
	//var nodeName string
	var cpu, memory, pods, storageEphemeral *resource.Quantity
	// not sure what storage does, so it will stay commented for now
	//var cpu, memory, storage, storageEphemeral, pods *resource.Quantity

	fmt.Printf("Cluster Utilization: \n")
	//fmt.Printf("Nodes: \n")
	for _, node := range nodeMetricList.Items {
		//capacity = node.Status.Capacity
		//nodeName = node.ObjectMeta.Name
		rl = node.Usage
		cpu = rl.Cpu()
		memory = rl.Memory()
		pods = rl.Pods()
		storageEphemeral = rl.StorageEphemeral()
		// Print resources per node
		// Gating this behind a flag
		//fmt.Printf("name: %v\n", nodeName)
		//fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", cpu, memory, pods, storageEphemeral)
		//fmt.Printf("\n")

		// Add resources to current cluster total
		c.cpu.Add(*cpu)
		c.mem.Add(*memory)
		c.pods.Add(*pods)
		c.storageEphemeral.Add(*storageEphemeral)
	}
	// How can I use the value of cpu.Format rather than the resource.DecimalSI?
	//fmt.Printf("cpus: %v\n", clusterCpu)
	fmt.Printf("Total: \n")
	//memPrint, _ := clusterCpu.AsScale(resource.Giga)
	fmt.Printf("cpu: %v\nmem: %v\npods: %v\nephemeral-storage: %v\n", c.cpu, c.mem, "WIP", "WIP")
	fmt.Printf("\n\n")
	// Execute a function that adds node stats to a running total, prints node
	// info
}

// Cluster Reservations    (resourcequota)
func clusterReservation() {}

// cpu_lim
// cpu_req
