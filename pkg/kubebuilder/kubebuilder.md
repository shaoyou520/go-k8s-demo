### 安装部署
```  
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder
```

### 创建项目
```  
mkdir -p crd && cd crd
./../kubebuilder init --domain qt.domain --repo qt.doamin/App
```
如果创建的项目不在GOPATH中，为了让 kubebuilder 和 Go 识别导入路径，
需要运行 go mod init <modulename>

### 创建一个 API
```  
./../kubebuilder create api --group k8s.qt --version v1 --kind App
```
Create Resource [y/n]：y 创建文件api/v1/guestbook_types.go，此文件用来定义 API

Create Controller [y/n]：y 创建文件 controller/guestbook_controller.go，此文件用来编写实现 Kind(CRD) 的业务逻辑

#### 安装crd

```shell
make install
```

#### 部署自定义controller

> 开发时可以直接在本地调试。

1. 构建镜像
```shell
IMG=qtdocker/app-controller make docker-build
```
2. push镜像
```shell
IMG=qtdocker/app-controller make docker-push
```

3. 部署
> fix: 部署之前需要修改一下controllers/app_controller.go的rbac
> ```yaml
>//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
>//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
>//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
> ```
```shell
IMG=wangtaotao2015/app-controller make deploy
```

#### 验证

1. 创建一个app

```shell
kubectl apply -f config/samples
```

2. 检查是否创建了deployment

3. 修改app，看service、ingress是否能被创建

4. 访问ingress，看是否可以访问到服务


