package main

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

const (
	DefaultSecurityGroup = "default"
	DefaultProjectName   = "docker-machine"
	DefaultTemplateID    = "401"
	ImageName            = "Ubuntu 14.04"
	SshUserName          = "root"
)

func main() {
	plugin.RegisterDriver(&Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: SshUserName,
			SSHPort: 22,
		}})
}
