package plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"time"
)

// ListAndWatch 返回 Device 列表构成的数据流。
// 当 Device 状态发生变化或者 Device 消失时，ListAndWatch
// 会返回新的列表。
func (qtdp *QtTestDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	glog.Info("ListAndWatch called")
	devices := make([]*pluginapi.Device, 1)
	fsWatcher, _ := fsnotify.NewWatcher()
	fsWatcher.Events
	for i := range devices {
		devices[i] = &pluginapi.Device{
			ID:     qtdp.device_name,
			Health: pluginapi.Healthy,
		}
	}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: devices})

	for {
		time.Sleep(10 * time.Second)
	}
}

// Allocate 在容器创建期间调用，这样设备插件可以运行一些特定于设备的操作，
// 并告诉 kubelet 如何令 Device 可在容器中访问的所需执行的具体步骤
func (scdp *QtTestDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	glog.Info("Allocate")
	var response pluginapi.AllocateResponse

	for _, req := range r.ContainerRequests {
		var devices []*pluginapi.DeviceSpec
		for _, devid := range req.DevicesIDs {
			dev := new(pluginapi.DeviceSpec)
			containerPath := fmt.Sprintf("/tmp/k8s-socketcan/socketcan-%s", devid)
			dev.HostPath = fakeDevicePath
			dev.ContainerPath = containerPath
			dev.Permissions = "r"
			devices = append(devices, dev)

			scdp.assignmentCh <- &Assignment{
				containerPath,
				scdp.device_name,
			}
		}

		response.ContainerResponses = append(response.ContainerResponses,
			&pluginapi.ContainerAllocateResponse{
				Devices: devices,
			})

	}

	return &response, nil
}

// GetDevicePluginOptions 返回与设备管理器沟通的选项。
func (QtTestDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	glog.Info("GetDevicePluginOptions")
	return &pluginapi.DevicePluginOptions{}, nil
}

// PreStartContainer 在设备插件注册阶段根据需要被调用，调用发生在容器启动之前。
// 在将设备提供给容器使用之前，设备插件可以运行一些诸如重置设备之类的特定于
// 具体设备的操作，
func (QtTestDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	glog.Info("PreStartContainer")
	return nil, nil
}

// GetPreferredAllocation 从一组可用的设备中返回一些优选的设备用来分配，
// 所返回的优选分配结果不一定会是设备管理器的最终分配方案。
// 此接口的设计仅是为了让设备管理器能够在可能的情况下做出更有意义的决定。
func (QtTestDevicePlugin) GetPreferredAllocation(ctx context.Context, in *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	glog.Info("GetPreferredAllocation")
	return nil, nil
}
