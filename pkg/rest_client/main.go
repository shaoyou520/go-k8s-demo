package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	//RESTClient
	configfile := "/Users/qintao/tls/admin_user_kubeconfig.conf"
	config, err := clientcmd.BuildConfigFromFlags("", configfile)
	if err != nil {
		panic(err)
	}
	config.GroupVersion = &v1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"

	// client
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	// get data
	pods := v1.PodList{}
	err = restClient.Get().Namespace("default").Resource("pods").Do(context.TODO()).Into(&pods)
	if err != nil {
		println(err)
	} else {
		for i, pod := range pods.Items {
			fmt.Println(i, pod.Name)
		}

	}
}
