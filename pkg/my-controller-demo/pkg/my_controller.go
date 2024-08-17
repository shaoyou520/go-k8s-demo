package controller

import (
	"context"
	json2 "encoding/json"
	apicorev1 "k8s.io/api/core/v1"
	netcorev1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
	v13 "k8s.io/client-go/informers/core/v1"
	v14 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	netv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	workNum        = 5
	maxRetry       = 3
	annotationsKey = "ingress/http"
	templateFile   = "deploy/IngressTemplate.yml"
)

type MyController struct {
	restClient    kubernetes.Interface
	serviceLister v1.ServiceLister
	ingressLister netv1.IngressLister
	queue         workqueue.TypedRateLimitingInterface[any]
}

func NewMyController(clientSet kubernetes.Interface, serviceInformer v13.ServiceInformer,
	ingressInformer v14.IngressInformer) *MyController {
	c := &MyController{
		restClient:    clientSet,
		serviceLister: serviceInformer.Lister(),
		ingressLister: ingressInformer.Lister(),
		queue:         workqueue.NewTypedRateLimitingQueue[any](workqueue.DefaultTypedControllerRateLimiter[any]()),
	}
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
	})

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})

	return c
}

func (c MyController) Run(stopCh chan struct{}) {
	for i := 0; i < workNum; i++ {
		//启动一个协程, 每隔一段时间运行函数, 直到关闭
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func (c *MyController) worker() {
	for c.processNextItem() {
	}
}

func (c *MyController) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)
	key := item.(string)
	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

func (c MyController) syncService(key string) error {
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	//service被删除
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	//新增和删除
	_, ok := service.GetAnnotations()[annotationsKey]
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if ok && errors.IsNotFound(err) {
		// key存在, 并且ingress不存在, 创建ingress
		ig := c.constructIngress(service)
		_, err := c.restClient.NetworkingV1().Ingresses(namespaceKey).
			Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		// key不存在, ingress 存在, 删除ingress
		err := c.restClient.NetworkingV1().Ingresses(namespaceKey).
			Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// 出现错误, 放回queue
func (c MyController) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) <= maxRetry {
		c.queue.AddRateLimited(key)
		return
	}

	runtime.HandleError(err)
	c.queue.Forget(key)
}

func (c MyController) constructIngress(service *apicorev1.Service) *netcorev1.Ingress {
	yaml, err := os.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}
	values := map[string]string{
		"id":        string(service.GetUID()),
		"name":      service.Name,
		"namespace": service.Namespace,
		"port":      strconv.FormatInt(int64(service.Spec.Ports[0].Port), 10),
	}
	templateStr := string(yaml)
	for key, val := range values {
		templateStr = strings.ReplaceAll(templateStr, "${"+key+"}", val)
	}

	templateStr1, err := yaml2.ToJSON([]byte(templateStr))

	if err != nil {
		panic(err)
	}
	ingress := &netcorev1.Ingress{}
	err = json2.Unmarshal(templateStr1, ingress)
	if err != nil {
		panic(err)
	}
	return ingress
}

func (c MyController) addService(obj interface{}) {
	c.enqueue(obj)
}

func (c MyController) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}

	c.queue.Add(key)
}

func (c MyController) updateService(oldObj interface{}, newObj interface{}) {
	//todo 比较annotation
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	c.enqueue(newObj)
}

func (c MyController) deleteIngress(obj interface{}) {
	ingress := obj.(*netcorev1.Ingress)
	ownerReference := metav1.GetControllerOf(ingress)
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}

	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}
