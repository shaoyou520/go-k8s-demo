package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	//create config
	configfile := "/Users/qintao/tls/admin_user_kubeconfig.conf"
	config, err := clientcmd.BuildConfigFromFlags("", configfile)
	if err != nil {
		panic(err)
	}

	//create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deployments, err := clientset.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	for _, deployment := range deployments.Items {
		fmt.Println(deployment.Name)
	}

}
