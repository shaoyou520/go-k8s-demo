package plugin

import (
	"device_plugin/common"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/golang/glog"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"time"
)

type Assignment struct {
	ContainerPath string
	Name          string
}

type VCANDevicePlugin struct {
	assignmentCh chan *Assignment
	device_paths map[string]*Assignment
	client       *containerd.Client
	ctx          context.Context
}

func (p *VCANDevicePlugin) Start() error {
	go p.interfaceCreator()
	return nil
}

func (scdp *VCANDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devices := make([]*pluginapi.Device, 100)

	for i := range devices {
		devices[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("vcan-%d", i),
			Health: pluginapi.Healthy,
		}
	}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: devices})

	for {
		time.Sleep(10 * time.Second)
	}
}

func (scdp *VCANDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse

	for _, req := range r.ContainerRequests {
		var devices []*pluginapi.DeviceSpec
		for i, devid := range req.DevicesIDs {
			dev := new(pluginapi.DeviceSpec)
			containerPath := fmt.Sprintf("/tmp/k8s-socketcan/%s", devid)
			dev.HostPath = fakeDevicePath
			dev.ContainerPath = containerPath
			dev.Permissions = "r"
			devices = append(devices, dev)

			scdp.assignmentCh <- &Assignment{
				containerPath,
				fmt.Sprintf(common.VcanNameTemplate, i),
			}
		}

		response.ContainerResponses = append(response.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Devices: devices,
		})

	}

	return &response, nil
}

func (VCANDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (VCANDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return nil, nil
}

func (VCANDevicePlugin) GetPreferredAllocation(ctx context.Context, in *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return nil, nil
}

func (p *VCANDevicePlugin) interfaceCreator() {
	client, err := containerd.New("/var/run/k8s-socketcan/containerd.sock")
	if err != nil {
		glog.V(3).Info("Failed to connect to containerd")
		panic(err)
	}
	p.client = client

	context := context.Background()
	p.ctx = namespaces.WithNamespace(context, "k8s.io")

	p.device_paths = make(map[string]*Assignment)

	go func() {
		var retry *time.Timer = time.NewTimer(0)
		var waiting = false
		<-retry.C
		for {
			select {
			case alloc := <-p.assignmentCh:
				glog.V(3).Infof("New allocation request: %v", alloc)
				p.device_paths[alloc.ContainerPath] = alloc
			case <-retry.C:
				waiting = false
				glog.V(3).Infof("Trying to allocate: %v", p.device_paths)
				p.tryAllocatingDevices()
			}

			if !waiting && len(p.device_paths) > 0 {
				retry = time.NewTimer(common.ContainerWaitDelaySeconds * time.Second)
				waiting = true
			}
		}
	}()
}

// Searches through all containers for matching fake devices and creates the network interfaces.
func (p *VCANDevicePlugin) tryAllocatingDevices() {
	containers, err := p.client.Containers(p.ctx, "")
	if err != nil {
		glog.V(3).Infof("Failed to get container list: %v", err)
		return
	}

	for _, container := range containers {
		spec, err := container.Spec(p.ctx)
		if err != nil {
			glog.V(3).Infof("Failed to get fetch container spec: %v", err)
			return
		}
		for _, device := range spec.Linux.Devices {
			if assignment, ok := p.device_paths[device.Path]; ok {
				// we found a container we are looking for
				task, err := container.Task(p.ctx, nil)
				if err != nil {
					glog.Warningf("Failed to get the task: %v", err)
					return
				}

				pids, err := task.Pids(p.ctx)
				if err != nil {
					glog.Warningf("Failed to get task Pids: %v", err)
					return
				}

				err = p.createSocketcanInPod(assignment.Name, int(pids[0].Pid))
				if err != nil {
					glog.Warningf("Failed to create interface: %v: %v", assignment.Name, err)
					return
				}

				glog.V(3).Infof("Successfully created the vcan interface: %v", assignment)
				delete(p.device_paths, device.Path)
			}
		}
	}
}

// Creates the named vcan interface inside the pod namespace.
func (nbdp *VCANDevicePlugin) createSocketcanInPod(ifname string, containerPid int) error {
	la := netlink.NewLinkAttrs()
	la.Name = ifname
	la.Flags = net.FlagUp
	la.Namespace = netlink.NsPid(containerPid)

	return netlink.LinkAdd(&netlink.GenericLink{
		LinkAttrs: la,
		LinkType:  "vcan",
	})
}
