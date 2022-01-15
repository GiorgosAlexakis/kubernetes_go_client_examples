package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var kubeconfig string

type KubeOptions struct {
	Namespace string
	Pod       string
}

func markRequiredFlags(opts *KubeOptions) {
	if len(opts.Pod) == 0 {
		fmt.Fprintf(os.Stderr, "Fatal error,missing some variables with empty default values:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func parseFlags(opts *KubeOptions) {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&opts.Namespace, "namespace", "default", "namespace of the pod")
	flag.StringVar(&opts.Pod, "pod", "", "pod name(required)")
	flag.Parse()
	markRequiredFlags(opts)
}

func checkifPodExists(client *kubernetes.Clientset, opts *KubeOptions) bool {
	pod, err := client.CoreV1().Pods(opts.Namespace).Get(context.TODO(), opts.Pod, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Pod %s not found in namespace %s, request failed:%v\n", opts.Pod, opts.Namespace, err)
		return false
	}
	if pod.Status.Phase == "Running" {
		return true
	}
	return false
}

func main() {
	opts := &KubeOptions{}
	parseFlags(opts)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// show all pods the default namespace
	pods, err := clientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}
	// print pods names
	for _, pod := range pods.Items {
		fmt.Println(pod.Name, pod.Status.Phase)
	}
	PodExists := checkifPodExists(clientSet, opts)
	if PodExists {
		fmt.Printf("Pod %s is running in namespace %s\n", opts.Pod, opts.Namespace)
	} else {
		fmt.Printf("Pod %s is not running in namespace %s\n", opts.Pod, opts.Namespace)
	}

}
