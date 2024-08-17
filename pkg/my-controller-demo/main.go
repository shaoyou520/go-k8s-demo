package main

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"my-controller-demo/pkg"
)

// 需求:
//
//	1: 创建service时,添加 annotation=ingress/http:true 自动创建ingress
//	2: 删除service时,自动删除ingress; ownerReferences 从属自动删除
//	3: 更新时, 添加annotation=ingress/http:true 自动创建ingress, 去除则删除, 否则不处理
//	4: ingress 删除时自动创建
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

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))
	serviceInformer := factory.Core().V1().Services()
	ingressInformer := factory.Networking().V1().Ingresses()
	mycontroller := controller.NewMyController(clientset, serviceInformer, ingressInformer)
	stopch := make(chan struct{})
	// 初始化所有请求的通知者。它们在 goroutine 中处理
	// 运行直到停止通道关闭。
	// 警告：启动不会阻塞。当在 go 例程中运行时，它将与稍后的 WaitForCacheSync 竞争。
	factory.Start(stopch)
	//会阻塞，直到所有已启动的通知程序的缓存都已同步
	// 或者停止通道被关闭。
	factory.WaitForCacheSync(stopch)

	mycontroller.Run(stopch)

}
