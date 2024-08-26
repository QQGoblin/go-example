package main

import (
	"context"
	"flag"
	"github.com/QQGoblin/go-sdk/pkg/kubeutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var (
	namespace string
	podname   string
)

func init() {
	flag.StringVar(&namespace, "namespace", "default", "pod namespace")
	flag.StringVar(&podname, "name", "", "pod name")

}

func main() {

	flag.Parse()

	kubeconfig := "/etc/kubernetes/admin.conf"
	master := ""

	config := kubeutils.GetConfigOrDie(kubeconfig, master)

	kubeCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	pod, err := kubeCli.CoreV1().Pods(namespace).Get(context.Background(), podname, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
		Type:    "CustomCond",
		Reason:  "CustomCondReason",
		Status:  "CustomCondStatus",
		Message: "CustomCondMessage",
	})

	if _, err = kubeCli.CoreV1().Pods(namespace).UpdateStatus(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		klog.Fatal(err)
	}

}
