package plugin

import (
	"device_plugin/common"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/golang/glog"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/context"
	"time"
)

type QtTestDevicePlugin struct {
	assignmentCh chan *Assignment
	device_name  string
	device_paths map[string]*Assignment
	client       *containerd.Client
	ctx          context.Context
}

func (p *QtTestDevicePlugin) Start() error {
	go p.interfaceCreator()
	return nil
}

func (p *QtTestDevicePlugin) interfaceCreator() {
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
func (p *QtTestDevicePlugin) tryAllocatingDevices() {
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

				err = p.moveSocketcanIntoPod(assignment.Name, int(pids[0].Pid))
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
func (nbdp *QtTestDevicePlugin) moveSocketcanIntoPod(ifname string, containerPid int) error {
	link, err := netlink.LinkByName(ifname)
	if err != nil {
		return err
	}
	return netlink.LinkSetNsPid(link, containerPid)
}
