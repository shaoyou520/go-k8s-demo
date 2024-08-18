package main

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	clientset "my-crd/client/clientset/versioned"
	"my-crd/client/informers/externalversions"
)

func main() {

	configfile := "/Users/qintao/tls/admin_user_kubeconfig.conf"
	config, err := clientcmd.BuildConfigFromFlags("", configfile)
	if err != nil {
		panic(err)
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	list, err := clientset.MycontrollerV1().Foos(
		"default").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	for _, foo := range list.Items {
		println(foo.Name)
	}

	factory := externalversions.NewSharedInformerFactoryWithOptions(clientset,
		0, externalversions.WithNamespace("default"))
	factory.Mycontroller().V1().Foos().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("add foo")
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("delete foo")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("update foo")
		},
	})
	stopch := make(chan struct{})
	factory.Start(stopch)
	factory.WaitForCacheSync(stopch)
	<-stopch
}
