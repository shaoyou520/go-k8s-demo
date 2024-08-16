package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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

	//deployments, err := clientset.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	//for _, deployment := range deployments.Items {
	//	fmt.Println(deployment.Name)
	//}

	//factory := informers.NewSharedInformerFactory(clientset, 0)
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))
	informer := factory.Core().V1().Pods().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Println("add event: ", pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*v1.Pod)
			newpod := newObj.(*v1.Pod)
			fmt.Println("update event: ", old.Name, "  --> ", newpod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Println("delete event:", pod.Name)
		},
	})
	stopch := make(chan struct{})
	factory.Start(stopch)
	factory.WaitForCacheSync(stopch)
	<-stopch

}
