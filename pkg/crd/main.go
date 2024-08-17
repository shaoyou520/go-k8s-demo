package main

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	clientset "my-crd/client/clientset/versioned"
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

}
