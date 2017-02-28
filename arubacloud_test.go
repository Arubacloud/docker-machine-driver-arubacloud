package main

import (
    "testing"
    "github.com/docker/machine/libmachine/drivers"
)

func TestConfigFlags(t *testing.T) {
	driver := new(Driver)

	createFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"ac_username": "ARU-xxxx",
			"ac_password": "xxxx",
			"ac_admin_password": "xxxx",
			"ac_endpoint": "dc1",
			"ac_template": "ubuntu_xx_x",
			"ac_size": "Large",
		},
		CreateFlags: driver.GetCreateFlags(),
	}	
	
	if err := driver.SetConfigFromFlags(createFlags); err != nil {
		t.Errorf("Setting driver create flags failed. Error: " + err.Error())
	}


}

func TestDefaultConfigFlags(t *testing.T) {
	driver := new(Driver)

	createFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"ac_username": "ARU-xxxx",
			"ac_password": "xxxx",
			"ac_admin_password": "xxxx",
		},
		CreateFlags: driver.GetCreateFlags(),
	}	
	
	if err := driver.SetConfigFromFlags(createFlags); err != nil {
		t.Errorf("Setting driver create flags failed. Error: " + err.Error())
	}


}
