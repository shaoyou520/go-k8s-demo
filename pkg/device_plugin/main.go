package main

import (
	"device_plugin/manage"
	"device_plugin/plugin"
	"flag"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	flag.Set("logtostderr", "true")

	hw_devices := []string{}

	device_list := os.Getenv("devices")
	if device_list != "" {
		hw_devices = strings.Split(device_list, ",")
	}

	manager := manage.NewManager(plugin.QtTestLister{hw_devices})
	manager.Run()
}
