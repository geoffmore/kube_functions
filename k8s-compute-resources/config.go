package main

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	// auth/gcp necessary for gke
	// What does the _ syntax do?
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

var clientset *kubernetes.Clientset

func genClient() (*kubernetes.Clientset, error) {
	// - Boilerplate - //
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	// Doesn't currently work if you keep your context, users, and clusters in
	// different files. (I could implement something similar to a minify here,
	// but that would require reading a bunch of files
	// I _could_ look up the ':' delimited paths in the KUBECONFIG variable...
	// Or I _could_ use the output of kubectl config current-context --minify
	// The latter option feels fairly sloppy

	// Will declaring err multiple times cause a compilation error?
	// If so, how can I define one variable and simply assign to another?
	// err is redeclared already, so I don't think this will be a big deal like I
	// thought initially
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	panic(err.Error())
	//}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err.Error())
	//}
	return clientset, err
}
