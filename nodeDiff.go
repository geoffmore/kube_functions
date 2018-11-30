package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	k8s_core "k8s.io/api/core/v1"
)

const commandPrefix string = "kubectl get node"
const commandSuffix string = "-o json"
const headerPadding string = "---"

var specFlag bool
var condFlag bool
var labelFlag bool
var annotationFlag bool
var infoFlag bool
var allFlag bool

func initFlags() {
	flag.BoolVar(&specFlag, "spec", false, "Node Spec")
	flag.BoolVar(&condFlag, "cond", false, "Node Conditions")
	flag.BoolVar(&labelFlag, "label", false, "Node Labels")
	flag.BoolVar(&annotationFlag, "annotation", false, "Node Annotations")
	flag.BoolVar(&infoFlag, "info", false, "Node Info")
	flag.BoolVar(&allFlag, "all", false, "Enable all Flags")

	flag.Parse()
}
func stringInSlice(a string, list []string) bool {
	// https://stackoverflow.com/questions/15323767/does-go-have-if-x-in-construct-similar-to-python
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func runCommand(osCmdRaw string) []byte {
	// https://stackoverflow.com/questions/19238143/does-golang-support-variadic-function
	// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
	// https://stackoverflow.com/questions/32721066/pass-string-to-a-function-that-expects-a-variadic-parameter

	osCmdSplit := strings.Split(osCmdRaw, " ")
	osCmd := osCmdSplit[0]
	osCmdArgs := osCmdSplit[1:]

	osCmdStruct := exec.Command(osCmd, osCmdArgs...)
	stdOutErr, _ := osCmdStruct.CombinedOutput()
	// Throwing away err here because STDERR will tell us what went wrong with the command
	return stdOutErr
}

func genCommand(prefix, base, suffix string) string {
	return fmt.Sprintf("%s %s %s", prefix, base, suffix)
}

func conditionStatus2String(some k8s_core.ConditionStatus) string {
	var status string
	switch some {
	case k8s_core.ConditionTrue:
		status = "true"
	case k8s_core.ConditionFalse:
		status = "false"
	case k8s_core.ConditionUnknown:
		status = "unknown"
	default:
		status = "logical error"
	}
	return status
}

func genNodeInfo(node string) k8sNode {
	var nodeStructRaw k8s_core.Node
	var nodeStruct k8sNode

	cmdString := genCommand(commandPrefix, node, commandSuffix)
	rawJSON := runCommand(cmdString)

	// Marshal -> JSON
	// Unmarshal <- JSON
	json.Unmarshal(rawJSON, &nodeStructRaw)

	nodeStruct.arch = nodeStructRaw.Status.NodeInfo.Architecture
	nodeStruct.runtimeVer = nodeStructRaw.Status.NodeInfo.ContainerRuntimeVersion
	nodeStruct.kernelVer = nodeStructRaw.Status.NodeInfo.KernelVersion
	nodeStruct.kubeletVer = nodeStructRaw.Status.NodeInfo.KubeletVersion
	nodeStruct.osType = nodeStructRaw.Status.NodeInfo.OperatingSystem
	nodeStruct.osImage = nodeStructRaw.Status.NodeInfo.OSImage
	nodeStruct.nodeLabels = nodeStructRaw.ObjectMeta.Labels
	nodeStruct.capacity = nodeStructRaw.Status.Capacity
	nodeStruct.allocatable = nodeStructRaw.Status.Allocatable
	nodeStruct.nodeAnnotations = nodeStructRaw.ObjectMeta.Annotations
	nodeStruct.name = nodeStructRaw.Name

	nodeConditions := nodeStructRaw.Status.Conditions
	for _, condition := range nodeConditions {
		switch condition.Type {
		case "OutOfDisk":
			nodeStruct.isOutOfDisk = conditionStatus2String(condition.Status)
		case "MemoryPressure":
			nodeStruct.hasMemPressure = conditionStatus2String(condition.Status)
		case "DiskPressure":
			nodeStruct.hasDiskPressure = conditionStatus2String(condition.Status)
		case "PIDPressure":
			nodeStruct.hasPIDPressure = conditionStatus2String(condition.Status)
		case "Ready":
			nodeStruct.isReady = conditionStatus2String(condition.Status)
		}
	}
	return nodeStruct

}

type k8sNode struct {
	name string
	// Resources
	capacity    k8s_core.ResourceList
	allocatable k8s_core.ResourceList
	// Labels
	nodeLabels map[string]string // .ObjectMeta.Labels
	// Annotations
	nodeAnnotations map[string]string // .ObjectMeta.Annotations
	// NodeInfo
	runtimeVer string //  .Status.NodeInfo.ContainerRuntimeVersion
	kernelVer  string //  .Status.NodeInfo.KernelVersion
	kubeletVer string //  .Status.NodeInfo.KubeletVersion
	osType     string //  .Status.NodeInfo.OperatingSystem
	osImage    string //  .Status.NodeInfo.OSImage
	arch       string //  .Status.NodeInfo.Architecture
	// Conditions
	isOutOfDisk     string
	hasDiskPressure string
	hasMemPressure  string
	hasPIDPressure  string
	isReady         string
}

func nodeInfoDiff(node1, node2 k8sNode) []string {
	var nodeInfoDiffs []string
	nodeInfoDiffs = append(nodeInfoDiffs,
		fmt.Sprintf("%s,%s,%s", headerPadding, "NODE INFO", headerPadding),
		fmt.Sprintf("%s,%s,%s", "arch", node1.arch, node2.arch),
		fmt.Sprintf("%s,%s,%s", "os", node1.osType, node2.osType),
		fmt.Sprintf("%s,%s,%s", "os version", node1.osImage, node2.osImage),
		fmt.Sprintf("%s,%s,%s", "runtime", node1.runtimeVer, node2.runtimeVer),
		fmt.Sprintf("%s,%s,%s", "kernel", node1.kernelVer, node2.kernelVer),
		fmt.Sprintf("%s,%s,%s", "kubelet", node1.kubeletVer, node2.kubeletVer),
	)
	return nodeInfoDiffs
}

func nodeResourceDiff(node1, node2 k8sNode) []string {
	var nodeResourceDiffs []string
	nodeResourceDiffs = append(nodeResourceDiffs,
		fmt.Sprintf("%s,%s,%s", headerPadding, "RESOURCES", headerPadding),
		"Capacity,,",
		fmt.Sprintf("%s,%s,%s", "cpu", node1.capacity.Cpu(), node2.capacity.Cpu()),
		fmt.Sprintf("%s,%s,%s", "ephemeral-storage", node1.capacity.StorageEphemeral(), node2.capacity.StorageEphemeral()),
		fmt.Sprintf("%s,%s,%s", "memory", node1.capacity.Memory(), node2.capacity.Memory()),
		fmt.Sprintf("%s,%s,%s", "pods", node1.capacity.Pods(), node2.capacity.Pods()),

		"Allocated,,",
		fmt.Sprintf("%s,%s,%s", "cpu", node1.allocatable.Cpu(), node2.allocatable.Cpu()),
		fmt.Sprintf("%s,%s,%s", "ephemeral-storage", node1.allocatable.StorageEphemeral(), node2.allocatable.StorageEphemeral()),
		fmt.Sprintf("%s,%s,%s", "memory", node1.allocatable.Memory(), node2.allocatable.Memory()),
		fmt.Sprintf("%s,%s,%s", "pods", node1.allocatable.Pods(), node2.allocatable.Pods()),
	)
	return nodeResourceDiffs
}

func nodeConditionDiff(node1, node2 k8sNode) []string {
	var nodeConditionDiffs []string
	nodeConditionDiffs = append(nodeConditionDiffs,
		fmt.Sprintf("%s,%s,%s", headerPadding, "CONDITIONS", headerPadding),
		fmt.Sprintf("OutOfDisk,%s,%s", node1.isOutOfDisk, node2.isOutOfDisk),
		fmt.Sprintf("DiskPressure,%s,%s", node1.hasDiskPressure, node2.hasDiskPressure),
		fmt.Sprintf("MemoryPressure,%s,%s", node1.hasMemPressure, node2.hasMemPressure),
		fmt.Sprintf("PIDPresssure,%s,%s", node1.hasPIDPressure, node2.hasPIDPressure),
		fmt.Sprintf("Ready,%s,%s", node1.isReady, node2.isReady),
	)
	return nodeConditionDiffs
}
func nodeLabelDiff(node1, node2 k8sNode) []string {
	var nodeLabelDiffs []string
	nodeLabelDiffs = append(nodeLabelDiffs, fmt.Sprintf("%s,%s,%s", headerPadding, "LABELS", headerPadding))

	var usedLabels []string
	for label, val := range node1.nodeLabels {
		val2, exists := node2.nodeLabels[label]
		someFunc(&val, true)
		someFunc(&val2, exists)
		nodeLabelDiffs = append(nodeLabelDiffs, fmt.Sprintf("%s,%s,%s", label, val, val2))
		usedLabels = append(usedLabels, label)
	}
	for label2, val2 := range node2.nodeLabels {
		if !stringInSlice(label2, usedLabels) {
			val, exists := node1.nodeLabels[label2]
			someFunc(&val2, true)
			someFunc(&val, exists)
			nodeLabelDiffs = append(nodeLabelDiffs, fmt.Sprintf("%s,%s,%s", label2, val, val2))
		}
	}
	return nodeLabelDiffs
}

func nodeAnnotationDiff(node1, node2 k8sNode) []string {
	var nodeAnnotationDiffs []string

	nodeAnnotationDiffs = append(nodeAnnotationDiffs, fmt.Sprintf("%s,%s,%s", headerPadding, "ANNOTATIONS", headerPadding))
	var usedAnnotations []string
	for annotation, val := range node1.nodeAnnotations {
		val2, exists := node2.nodeAnnotations[annotation]
		someFunc(&val, true)
		someFunc(&val2, exists)
		nodeAnnotationDiffs = append(nodeAnnotationDiffs, fmt.Sprintf("%s,%s,%s", annotation, val, val2))
		usedAnnotations = append(usedAnnotations, annotation)
	}
	for annotation2, val2 := range node2.nodeAnnotations {
		if !stringInSlice(annotation2, usedAnnotations) {
			val, exists := node1.nodeAnnotations[annotation2]
			someFunc(&val2, true)
			someFunc(&val, exists)
			nodeAnnotationDiffs = append(nodeAnnotationDiffs, fmt.Sprintf("%s,%s,%s", annotation2, val, val2))
		}
	}
	return nodeAnnotationDiffs
}

func nodeDiff(node1, node2 k8sNode) []string {
	var nodeDiffCSV []string

	//Easy: Condiitons, Node Info
	//Medium: Capacity, Allocated
	//Hard: Labels, Annotations

	nodeDiffCSV = append(nodeDiffCSV, fmt.Sprintf("Name,%s,%s", node1.name, node2.name))

	if allFlag {
		nodeDiffCSV = append(nodeDiffCSV, nodeInfoDiff(node1, node2)...)
		nodeDiffCSV = append(nodeDiffCSV, nodeResourceDiff(node1, node2)...)
		nodeDiffCSV = append(nodeDiffCSV, nodeConditionDiff(node1, node2)...)
		nodeDiffCSV = append(nodeDiffCSV, nodeLabelDiff(node1, node2)...)
		nodeDiffCSV = append(nodeDiffCSV, nodeAnnotationDiff(node1, node2)...)
	} else {
		if infoFlag {
			nodeDiffCSV = append(nodeDiffCSV, nodeInfoDiff(node1, node2)...)
		}
		if specFlag {
			nodeDiffCSV = append(nodeDiffCSV, nodeResourceDiff(node1, node2)...)

		}
		if condFlag {
			nodeDiffCSV = append(nodeDiffCSV, nodeConditionDiff(node1, node2)...)
		}
		if labelFlag {
			nodeDiffCSV = append(nodeDiffCSV, nodeLabelDiff(node1, node2)...)
		}
		if annotationFlag {
			nodeDiffCSV = append(nodeDiffCSV, nodeAnnotationDiff(node1, node2)...)
		}
	}
	return nodeDiffCSV
}

func someFunc(origVal *string, present bool) {
	if !present {
		*origVal = "DNE"
	} else if *origVal == "" {
		*origVal = "BLANK"
	}
}
func main() {
	var node1string, node2string string

	// https://stackoverflow.com/questions/35809252/check-if-flag-was-provided-in-go
	// Boolean flags are true if set and false otherwise
	initFlags()

	// https://stackoverflow.com/questions/19617229/golang-flag-gets-interpreted-as-first-os-args-argument
	// Parse non-command line arguments
	nonFlagArgs := flag.Args()
	node1string = nonFlagArgs[0]
	node2string = nonFlagArgs[1]

	// Generate the struct we need to compare nodes
	node1 := genNodeInfo(node1string)
	node2 := genNodeInfo(node2string)

	// Compare nodes based on flags that were selected
	for _, line := range nodeDiff(node1, node2) {
		fmt.Println(line)
	}
	// TODO
	// Verify string input for nodes before running genNodeInfo
	// Convert CSV to a table using an existing go package
	// Other stuff

}
