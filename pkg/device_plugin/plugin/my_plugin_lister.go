package plugin

import (
	"device_plugin/manage"
	"github.com/golang/glog"
	"strings"
)

var (
	ResourceNamespace = "plugin-test"
)

type QtTestLister struct {
	Real_devices []string
}

func (qtl QtTestLister) GetResourceNamespace() string {
	return ResourceNamespace
}

func (qtl QtTestLister) Discover(pluginListCh chan manage.PluginNameList) {
	var plugins = manage.PluginNameList(qtl.Real_devices)
	pluginListCh <- plugins
}

func (qtl QtTestLister) NewPlugin(kind string) manage.PluginInterface {
	glog.V(3).Infof("Creating device plugin %s", kind)
	return &QtTestDevicePlugin{
		assignmentCh: make(chan *Assignment),
		device_name:  strings.TrimPrefix(kind, "qt-test-"),
	}
}
