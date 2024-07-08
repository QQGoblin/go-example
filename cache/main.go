package main

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultKubeConfig = "/etc/kubernetes/admin.conf"

func main() {

	restConfig, err := clientcmd.BuildConfigFromFlags("", defaultKubeConfig)
	if err != nil {
		klog.Fatalf("Build config from flags failed: %v", err)
	}

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		klog.Fatalf("AddToScheme scheme failed: %v", err)

	}

	c, err := cache.New(restConfig, cache.Options{
		Scheme:          scheme,
		DefaultSelector: cache.ObjectSelector{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errChan := make(chan error)

	go func() {
		if startErr := c.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	if !c.WaitForCacheSync(context.TODO()) {
		klog.Fatalf("Sync cache failed")
	}

	object := &v1.Node{}
	objKey := client.ObjectKey{
		Namespace: "",
		Name:      "node1"}
	if err := c.Get(context.TODO(), objKey, object); err != nil {
		klog.Fatalf("Get object %s failed: %v", objKey.String(), err)
	}

	klog.Infof("Get object %s<%v>", object.Name, object.GetObjectMeta().GetUID())

}
