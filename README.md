# Go Kubernetes 开发示例集

本项目是一组使用 Go 语言进行 Kubernetes 开发的示例代码，涵盖了从基础客户端使用到高级控制器开发、设备插件、Prometheus Exporter 以及 Webhook 等多种场景。

## 项目结构

```
├── doc/                        # 文档
│   ├── code-generator.md       # code-generator 使用说明
│   └── RuntimeService.proto    # CRI RuntimeService 协议定义
├── fsnotify.md                 # fsnotify 文件监听库使用说明
├── pkg/
│   ├── rest_client/            # RESTClient 示例
│   ├── clientset/              # ClientSet + Informer 示例
│   ├── crd/                    # 自定义 CRD 示例（使用 code-generator）
│   ├── my-controller-demo/     # 自定义控制器示例
│   ├── device_plugin/          # Kubernetes 设备插件示例
│   ├── exporter/               # Prometheus Exporter 示例
│   ├── kubebuilder/            # Kubebuilder 示例（CRD + Webhook）
│   └── webhook/                # Webhook e2e 示例
```

## 模块说明

### 1. RESTClient 示例 (`pkg/rest_client/`)

使用 Kubernetes 最底层的 `rest.RESTClient` 直接与 API Server 交互，获取 Pod 列表。

**核心流程：**
- 通过 `clientcmd.BuildConfigFromFlags` 加载 kubeconfig 配置
- 配置 `GroupVersion`、`NegotiatedSerializer`、`APIPath`
- 创建 `rest.RESTClient` 并调用 `.Get().Namespace("default").Resource("pods")` 获取 Pod 列表

**运行方式：**
```bash
cd pkg/rest_client
go run main.go
```

### 2. ClientSet + Informer 示例 (`pkg/clientset/`)

展示如何使用 `kubernetes.Clientset` 结合 `SharedInformerFactory` 监听 Kubernetes 资源事件。

**核心流程：**
- 创建 `kubernetes.Clientset`
- 使用 `informers.NewSharedInformerFactoryWithOptions` 创建 Informer 工厂（限定 `default` 命名空间）
- 注册 `AddFunc`、`UpdateFunc`、`DeleteFunc` 事件处理函数
- 启动 Informer 并等待缓存同步

**运行方式：**
```bash
cd pkg/clientset
go run clientset_demo.go
```

### 3. 自定义 CRD 示例 (`pkg/crd/`)

使用 `code-generator` 为自定义资源 `Foo` 生成 clientset、informer、lister 代码。

**CRD 定义：**
- Group: `mycontroller.k8s.io`
- Version: `v1`
- Kind: `Foo`
- Spec 包含: `deploymentName` (string), `replicas` (int32)
- Status 包含: `availableReplicas` (int32)

**目录结构：**
- `api/mycontroller/v1/` - 类型定义 (`types.go`, `doc.go`, `register.go`)
- `client/` - code-generator 生成的客户端代码
- `manifest/my-crd.yaml` - CRD YAML 定义文件
- `hack/` - 代码生成脚本

**使用步骤：**
1. 安装 CRD: `kubectl apply -f manifest/my-crd.yaml`
2. 运行程序: `cd pkg/crd && go run main.go`

详细的 code-generator 使用说明请参考 `doc/code-generator.md`。

### 4. 自定义控制器示例 (`pkg/my-controller-demo/`)

实现了一个基于 Informer + WorkQueue 的自定义控制器，自动管理 Service 与 Ingress 的关联。

**核心需求：**
1. 创建 Service 时，如果包含 annotation `ingress/http: true`，自动创建对应的 Ingress
2. 删除 Service 时，通过 `ownerReferences` 自动删除关联的 Ingress
3. 更新 Service 时，根据 annotation 的变化自动创建或删除 Ingress
4. Ingress 被意外删除时，自动重新创建

**架构设计：**
- `MyController` 结构体包含 `restClient`、`serviceLister`、`ingressLister` 和 `workqueue`
- 监听 Service 的 Add/Update 事件和 Ingress 的 Delete 事件
- 使用 5 个 worker 协程并发处理队列中的事件
- 最大重试次数为 3 次
- 通过 YAML 模板 (`deploy/IngressTemplate.yml`) 构建 Ingress 对象

**运行方式：**
```bash
cd pkg/my-controller-demo
go run main.go
```

### 5. Kubernetes 设备插件示例 (`pkg/device_plugin/`)

实现了一个完整的 Kubernetes Device Plugin，用于向 kubelet 注册和管理自定义硬件设备。

**架构设计：**
- `manage/` - 插件管理框架
  - `Manager`: 管理所有设备插件的生命周期，监听系统信号和 kubelet socket 变化
  - `ListerInterface`: 设备发现接口，负责发现和创建插件
  - `PluginInterface`: 设备插件接口，封装 `pluginapi.DevicePluginServer`
  - `devicePlugin`: gRPC 服务端实现，负责向 kubelet 注册和提供设备
- `plugin/` - 具体设备插件实现
  - `QtTestDevicePlugin`: 测试设备插件，实现了 `ListAndWatch`、`Allocate` 等方法
  - `QtTestLister`: 设备列表发现器，资源命名空间为 `plugin-test`

**设备插件 gRPC 接口：**
- `ListAndWatch`: 返回设备列表，当设备状态变化时推送更新
- `Allocate`: 容器创建时调用，返回设备挂载信息
- `GetDevicePluginOptions`: 返回插件选项
- `PreStartContainer`: 容器启动前的预处理
- `GetPreferredAllocation`: 返回优选设备分配方案

**运行方式：**
```bash
cd pkg/device_plugin
# 通过环境变量设置设备列表
export devices="device1,device2"
go run main.go
```

### 6. Prometheus Exporter 示例 (`pkg/exporter/`)

展示如何使用 `prometheus/client_golang` 库创建自定义 Prometheus 指标导出器。

**定义的指标类型：**

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `http_request_count` | Counter | HTTP 请求计数，按 endpoint 和 code 标签分类 |
| `order_num` | Gauge | 订单数量，可增可减 |
| `http_request_duration` | Summary | HTTP 请求耗时分布，按 endpoint 标签分类 |
| `example_histogram` | Histogram | 直方图示例，桶宽度为 10 |

**服务端点：**
- `/metrics` - Prometheus 抓取端点
- `/hello/` - 业务请求端点（会更新上述指标）
- 监听端口: `8888`

**运行方式：**
```bash
cd pkg/exporter
go run main.go
```

**配套 Prometheus 配置：** `my-prometheus-server.yaml` 提供了完整的 Prometheus Server ConfigMap 配置，包含对 K8s API Server、Node、Pod、Service 等的自动发现和抓取规则。

### 7. Kubebuilder 示例 (`pkg/kubebuilder/`)

使用 Kubebuilder 框架创建 CRD 和控制器。

**快速开始：**
```bash
# 安装 kubebuilder
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder

# 创建项目
mkdir -p crd && cd crd
kubebuilder init --domain qt.domain --repo qt.domain/App

# 创建 API
kubebuilder create api --group k8s.qt --version v1 --kind App

# 安装 CRD
make install

# 构建并推送镜像
IMG=qtdocker/app-controller make docker-build
IMG=qtdocker/app-controller make docker-push

# 部署
IMG=qtdocker/app-controller make deploy
```

**Webhook 支持：**
```bash
kubebuilder create webhook --group k8s.qt --version v1 --kind App --defaulting --conversion --programmatic-validation
```

支持三种 Webhook 类型：
- **Defaulting Webhook** (MutatingAdmissionWebhook): 为资源设置默认值
- **Validating Webhook** (ValidatingAdmissionWebhook): 验证资源合法性
- **Conversion Webhook**: 多版本 CRD 之间的转换

详细说明请参考:
- `pkg/kubebuilder/kubebuilder.md`
- `pkg/kubebuilder/webhook.md`

### 8. Webhook 示例 (`pkg/webhook/e2e_example/`)

提供了 Kubernetes 准入控制 Webhook 的端到端示例，包含多种 Webhook handler:
- `addlabel.go` - 添加标签的 Mutating Webhook
- `alwaysallow.go` - 总是允许的 Validating Webhook
- `alwaysdeny.go` - 总是拒绝的 Validating Webhook
- `pods.go` - Pod 相关的 Webhook
- `crd.go` - CRD 相关的 Webhook
- `customresource.go` - 自定义资源的 Webhook
- `convert.go` - 资源转换 Webhook

## 前置条件

- Go 1.21+
- Kubernetes 集群（用于运行示例）
- kubectl 命令行工具
- kubeconfig 配置文件

## 相关文档

- `doc/code-generator.md` - code-generator 工具使用指南
- `pkg/kubebuilder/kubebuilder.md` - Kubebuilder 使用指南
- `pkg/kubebuilder/webhook.md` - Webhook 准入控制说明
- `fsnotify.md` - fsnotify 文件监听库使用说明
