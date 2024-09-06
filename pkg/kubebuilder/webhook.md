### 准入控制

前面我们有学习到k8s提供了一系列的准入控制器，通过它们我们可以对api server的请求
进行处理。而对于我们自定义的需求，可以通过[MutatingAdmissionWebhook](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)
和[ValidatingAdmissionWebhook](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook)
进行处理。

### 创建webhook
1. 生成代码 创建默认、转换、验证 Webhook
``` 
kubebuilder create webhook --group k8s.qt --version v1 --kind App --defaulting --conversion --programmatic-validation
```

创建之后，在main.go中会添加以下代码:

```go
	if err = (&ingressv1beta1.App{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "App")
		os.Exit(1)
	}
```

同时会生成下列文件，主要有：

- api/v1beta1/app_webhook.go webhook对应的handler，我们添加业务逻辑的地方

- api/v1beta1/webhook_suite_test.go 测试

- config/certmanager 自动生成自签名的证书，用于webhook server提供https服务

- config/webhook 用于注册webhook到k8s中

- config/crd/patches 为conversion自动注入caBoundle

- config/default/manager_webhook_patch.yaml 让manager的deployment支持webhook请求
- config/default/webhookcainjection_patch.yaml 为webhook server注入caBoundle

注入caBoundle由cert-manager的[ca-injector](https://cert-manager.io/docs/concepts/ca-injector/#examples) 组件实现
