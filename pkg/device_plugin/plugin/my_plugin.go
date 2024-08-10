package plugin

import (
	"github.com/containerd/containerd"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type QtTestDevicePlugin struct {
	assignmentCh chan *Assignment
	device_name  string
	device_paths map[string]*Assignment
	client       *containerd.Client
	ctx          context.Context
}

func (p *QtTestDevicePlugin) Start() error {
	glog.Info("starting QtTestDevicePlugin")
	return nil
}
