package manage

import (
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type PluginInterface interface {
	pluginapi.DevicePluginServer
}

type PluginInterfaceStart interface {
	Start() error
}

type PluginInterfaceStop interface {
	Stop() error
}

// DevicePlugin represents a gRPC server client/server.
type devicePlugin struct {
	DevicePluginImpl PluginInterface
	ResourceName     string
	Name             string
	Socket           string
	Server           *grpc.Server
	Running          bool
	Starting         *sync.Mutex
}

func newDevicePlugin(resourceNamespace string, pluginName string, devicePluginImpl PluginInterface) devicePlugin {
	return devicePlugin{
		DevicePluginImpl: devicePluginImpl,
		Socket:           pluginapi.DevicePluginPath + resourceNamespace + "_" + pluginName,
		ResourceName:     resourceNamespace + "/" + pluginName,
		Name:             pluginName,
		Starting:         &sync.Mutex{},
	}
}

// 服务注册
func (dpi *devicePlugin) StartServer() error {
	glog.V(3).Infof("%s: Starting plugin server", dpi.Name)

	dpi.Starting.Lock()
	defer dpi.Starting.Unlock()

	if dpi.Running {
		return nil
	}

	err := dpi.serve()
	if err != nil {
		return err
	}

	err = dpi.register()
	if err != nil {
		dpi.StopServer()
		return err
	}
	dpi.Running = true

	return nil
}

// 启动服务
func (dpi *devicePlugin) serve() error {
	glog.V(3).Infof("%s: Starting the DPI gRPC server", dpi.Name)

	err := dpi.cleanup()
	if err != nil {
		glog.Errorf("%s: Failed to setup a DPI gRPC server: %s", dpi.Name, err)
		return err
	}

	sock, err := net.Listen("unix", dpi.Socket)
	if err != nil {
		glog.Errorf("%s: Failed to setup a DPI gRPC server: %s", dpi.Name, err)
		return err
	}

	dpi.Server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(dpi.Server, dpi.DevicePluginImpl)

	go dpi.Server.Serve(sock)
	glog.V(3).Infof("%s: Serving requests...", dpi.Name)
	// Wait till grpc server is ready.
	for i := 0; i < 10; i++ {
		services := dpi.Server.GetServiceInfo()
		if len(services) >= 1 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// 向k8s 注册服务
func (dpi *devicePlugin) register() error {
	glog.V(3).Infof("%s: Registering the DPI with Kubelet", dpi.Name)

	conn, err := grpc.NewClient("unix://"+pluginapi.KubeletSocket,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		glog.Errorf("%s: Could not dial gRPC: %s", dpi.Name, err)
		return err
	}
	client := pluginapi.NewRegistrationClient(conn)
	glog.Infof("%s: Registration for endpoint %s", dpi.Name, path.Base(dpi.Socket))

	options, err := dpi.DevicePluginImpl.GetDevicePluginOptions(context.Background(), &pluginapi.Empty{})
	if err != nil {
		glog.Errorf("%s: Failed to get device plugin options %s", dpi.Name, err)
		return err
	}

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(dpi.Socket),
		ResourceName: dpi.ResourceName,
		Options:      options,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		glog.Errorf("%s: Registration failed: %s", dpi.Name, err)
		glog.Errorf("%s: Make sure that the DevicePlugins feature gate is enabled and kubelet running", dpi.Name)
		return err
	}
	return nil
}

// 停止服务
func (dpi *devicePlugin) StopServer() error {
	// TODO: should this also be a critical section?
	// how do we prevent multiple stops? or start/stop race condition?
	glog.V(3).Infof("%s: Stopping plugin server", dpi.Name)

	if !dpi.Running {
		glog.V(3).Infof("%s: Tried to stop stopped DPI", dpi.Name)
		return nil
	}

	glog.V(3).Infof("%s: Stopping the DPI gRPC server", dpi.Name)
	dpi.Server.Stop()
	dpi.Running = false

	return dpi.cleanup()
}

// 清理socket
func (dpi *devicePlugin) cleanup() error {
	if err := os.Remove(dpi.Socket); err != nil && !os.IsNotExist(err) {
		glog.Errorf("%s: Could not clean up socket %s: %s", dpi.Name, dpi.Socket, err)
		return err
	}

	return nil
}
