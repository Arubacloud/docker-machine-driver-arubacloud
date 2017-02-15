package main

import (
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
	plugin.RegisterDriver(new(Driver))
}
