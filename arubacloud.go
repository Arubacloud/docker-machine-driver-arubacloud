package main

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/arubacloud/goarubacloud"
    "github.com/arubacloud/goarubacloud/models"
	"time"
	"fmt"
	"io/ioutil"
	"net"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/ssh"
	"path/filepath"
	"os"
)

const (
	statusTimeout = 200
)

type Driver struct {
	drivers.BaseDriver

	TemplateID    int
	TemplateName string
	Size string
	PackageID     int
	AdminPassword string
	Username      string
	Password      string
	Endpoint      string
	ConfigureIPv6 bool

	// internal ids
	ServerId      int
	ServerName    string
	SSHKey   string
	
	Action   string

	// internal
	client        *goarubacloud.API
}

const(
	defaultTemplate = "ubuntu1604_x64_1_0"
	defaultEndpoint = "dc1"
	defaultSize = "Large"
	machineType = "NewSmart"
)

// GetCreateFlags registers the "machine create" flags recognized by this driver, including
// their help text and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "AC_USERNAME",
			Name: "ac_username",
			Usage: "ArubaCloud Username",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "AC_PASSWORD",
			Name: "ac_password",
			Usage: "ArubaCloud Password",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "AC_ADMIN_PASSWORD",
			Name: "ac_admin_password",
			Usage: "ArubaCloud Machine root password",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "AC_ENDPOINT",
			Name: "ac_endpoint",
			Usage: "ArubaCloud Endpoint name (dc1,dc2,dc3 etc.)",
			Value: defaultEndpoint,
		},
		mcnflag.StringFlag{
			EnvVar: "AC_TEMPLATE",
			Name: "ac_template",
			Usage: "ArubaCloud VM Template",
			Value: defaultTemplate,
		},
		mcnflag.StringFlag{
			EnvVar: "AC_SIZE",
			Name: "ac_size",
			Usage: "ArubaCloud Machine Size",
			Value: defaultSize,
		},
		mcnflag.StringFlag{
			EnvVar: "AC_ACTION",
			Name: "ac_action",
			Usage: "ArubaCloud Action type",
			Value: machineType,
		},
		mcnflag.StringFlag{
			EnvVar: "AC_IP",
			Name: "ac_ip",
			Usage: "Set this to use an already purchased Ip Address",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "AC_SSH_KEY",
			Name: "ac_ssh_key",
			Usage: "Absolute path of the ssh private key",
			Value: "",
		},
		mcnflag.BoolFlag{
			EnvVar: "AC_IPV6",
			Name: "ac_ipv6",
			Usage: "Configure an IPv6 address for the ArubaCloud VM",
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "arubacloud"
}

func (d *Driver) PreCreateCheck() error {

	return nil
}

// getClient returns an ArubaCloud API client pointing to dc1
func (d *Driver) getClient() (api *goarubacloud.API) {
	if d.client == nil {
		client, err := goarubacloud.NewAPI(d.Endpoint, d.Username, d.Password)
		if err != nil {
			return nil
		}
		d.client = client
	}

	return d.client
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Username = flags.String("ac_username")
	d.Password = flags.String("ac_password")
	d.AdminPassword = flags.String("ac_admin_password")
	d.PackageID = flags.Int("ac_package_id")
	d.TemplateName = flags.String("ac_template")
	d.Size = flags.String("ac_size")
	d.Endpoint = flags.String("ac_endpoint")
	d.SSHKey = flags.String("ac_ssh_key")
	d.Action = flags.String("ac_action")
	d.IPAddress = flags.String("ac_ip")
	d.SSHUser = "root"
	d.ConfigureIPv6 = flags.Bool("ac_ipv6")

	return nil
}

func (d *Driver) waitForServerStatus(status int) (server *models.Server, err error) {
	//func WaitForSpecificOrError(f func() (bool, error), maxAttempts int, waitInterval time.Duration) error
	return server, mcnutils.WaitForSpecificOrError(func() (bool, error) {
		server, err = d.client.GetServer(d.ServerId)
		if err != nil {
			return true, err
		}
		log.Debugf("Machine", map[string]interface{}{
			"Name":  d.ServerName,
			"State": server.ServerStatus,
		})

		if server.ServerStatus == 4 {
			return true, fmt.Errorf("Instance creation failed. Instance is in ERROR state")
		}

		if server.ServerStatus == status {
			return true, nil
		}
		return false, nil
	}, 10, 60 * time.Second)
}

func (d *Driver) CreateSmart() error {
	
	log.Debug("Create ", d.TemplateName)
	client := d.getClient()

	key, err := d.createKeyPair()
	if err != nil {
		return err
	}

	

	log.Debug("Get Template ", d.TemplateName)
	template, err := client.GetTemplate(d.TemplateName, 4)
	if err != nil {
		return err
	} else {
		log.Debug("Template found with Id: ", template.Id)
	}
	
	log.Debug("Get Package ", d.TemplateName)
	cloudpackage, err := client.GetPreconfiguredPackage(d.Size)
	if err != nil {
		return err
	} else {
		log.Debug("Package found with Id: ", cloudpackage.PackageID)
	}
	
	// Create instance
	log.Debug("Creating ArubaCloud server... with packageID: ", cloudpackage.PackageID)

	instance, err := client.CreateServerSmart(
		d.MachineName,
		d.AdminPassword,
		cloudpackage.PackageID,
		template.Id,
		key,
		d.ConfigureIPv6,
	)

	if err != nil {
		log.Debug(err)
		return err
	}

	log.Debug("Waiting for the server to be ready...")
	servers, err := client.GetServers()
	if err != nil {
		return err
	}

	// Retrieving ServerID from server list
	for _, server := range servers {
		log.Debugf("Iterating server name: %s", server.Name)
		if server.Name == d.MachineName {
			d.ServerId = server.ServerId
			log.Debugf("Setting Driver ServerId to: %d", d.ServerId)
		}
	}

	if d.ServerId == 0 {
		return fmt.Errorf("No Server found with Name: %s", d.MachineName)
	}

	// Retrieve ServerDetails for the given ServerID
	detailed_server_response, err := client.GetServer(d.ServerId)
	if err != nil {
		return err
	}

	// Override instance object with the new unmarshaled detailed server response
	instance = detailed_server_response

	// Wait until instance is ACTIVE
	log.Debugf("Waiting for ArubaCloud Server...", map[string]interface{}{"MachineID": d.ServerId})
	instance, err = d.waitForServerStatus(3)
	if err != nil {
		return err
	}

	// In order to obtain the IP address we have to get the server detail

	// Save Ip address that should be available at this point
	d.IPAddress = ""
	d.IPAddress = instance.EasyCloudIPAddress.Value

	if d.IPAddress == "" {
		return fmt.Errorf("No IP found for instance %s", instance.ServerId)
	}

	log.Debugf("IP address found", map[string]interface{}{
		"MachineID": d.ServerId,
		"IP":        d.IPAddress,
	})
	
	return nil
}

func (d *Driver) CreatePro() error {
	
	log.Debug("Create ", d.TemplateName)
	client := d.getClient()

	key, err := d.createKeyPair()
	if err != nil {
		return err
	}

	

	log.Debug("Get Template ", d.TemplateName)
	template, err := client.GetTemplate(d.TemplateName, 2)
	if err != nil {
		return err
	} else {
		log.Debug("Template found with Id: ", template.Id)
	}
	
	ipID := 0
	ipAddressValue := ""
	
	if len(d.IPAddress) > 0 {
		log.Debug("Get IpAddress ", d.IPAddress)
		ipAddress, err := client.GetPurchasedIpAddress(d.IPAddress)
		ipID = ipAddress.ResourceId
		ipAddressValue = d.IPAddress
		if err != nil {
			return err
		} else {
			log.Debug("IpAddress found with Id: ", ipAddress.ResourceId)
		}
	} else {
		log.Debug("Purchasing IpAddress ", d.IPAddress)
		ipAddress, err := client.PurchaseIpAddress()
		ipID = ipAddress.ResourceId
		ipAddressValue = ipAddress.Value
		if err != nil {
			return err
		} else {
			log.Debug("IpAddress purchased with Id: ", ipAddress.ResourceId)
		}
	}
	
	
	// Create instance
	log.Debug("Creating ArubaCloud server...")
	
	diskSize := 20
	cpuQuantity := 1
	ramQuantity := 1
	
	switch d.Size{
		case "Small":
			diskSize = 20
			cpuQuantity = 1
			ramQuantity = 1
		case "Medium":
			cpuQuantity = 1
			ramQuantity = 2
			diskSize = 40
		case "Large":
			cpuQuantity = 2
			ramQuantity = 4
			diskSize = 80
		case "Extra Large":
			cpuQuantity = 4
			ramQuantity = 8
			diskSize = 160
			
	}
	
	
	instance, err := client.CreateServerPro(
		d.MachineName,
		d.AdminPassword,
		template.Id,
		key,
		ipID,
		diskSize,
		cpuQuantity,
		ramQuantity,
		d.ConfigureIPv6,
	)

	if err != nil {
		log.Debug(err)
		return err
	}

	log.Debug("Waiting for the server to be ready...")
	servers, err := client.GetServers()
	if err != nil {
		return err
	}

	// Retrieving ServerID from server list
	for _, server := range servers {
		log.Debugf("Iterating server name: %s", server.Name)
		if server.Name == d.MachineName {
			d.ServerId = server.ServerId
			log.Debugf("Setting Driver ServerId to: %d", d.ServerId)
		}
	}

	if d.ServerId == 0 {
		return fmt.Errorf("No Server found with Name: %s", d.MachineName)
	}

	// Retrieve ServerDetails for the given ServerID
	detailed_server_response, err := client.GetServer(d.ServerId)
	if err != nil {
		return err
	}

	// Override instance object with the new unmarshaled detailed server response
	instance = detailed_server_response

	// Wait until instance is ACTIVE
	log.Debugf("Waiting for ArubaCloud Server...", map[string]interface{}{"MachineID": d.ServerId})
	instance, err = d.waitForServerStatus(3)
	if err != nil {
		return err
	}

	// In order to obtain the IP address we have to get the server detail

	// Save Ip address that should be available at this point
	d.IPAddress = ""
	d.IPAddress = ipAddressValue

	if d.IPAddress == "" {
		return fmt.Errorf("No IP found for instance %s", instance.ServerId)
	}

	log.Debugf("IP address found", map[string]interface{}{
		"MachineID": d.ServerId,
		"IP":        d.IPAddress,
	})
	
	return nil
}

func (d *Driver) Attach() error {
	
	log.Debug("Attaching machine %s at %s", d.MachineName, d.IPAddress)
	client := d.getClient()

	_, err := d.createKeyPair()
	if err != nil {
		return err
	}

	log.Debug("Waiting for the server to be ready...")
	servers, err := client.GetServers()
	if err != nil {
		return err
	}

	// Retrieving ServerID from server list
	for _, server := range servers {
		log.Debugf("Iterating server name: %s", server.Name)
		if server.Name == d.MachineName {
			d.ServerId = server.ServerId
			log.Debugf("Setting Driver ServerId to: %d", d.ServerId)
		}
	}

	if d.ServerId == 0 {
		return fmt.Errorf("No Server found with Name: %s", d.MachineName)
	}

	// Retrieve ServerDetails for the given ServerID
	_, err = client.GetServer(d.ServerId)
	if err != nil {
		return err
	}

	// Wait until instance is ACTIVE
	log.Debugf("Waiting for ArubaCloud Server...", map[string]interface{}{"MachineID": d.ServerId})
	_, err = d.waitForServerStatus(3)
	if err != nil {
		return err
	}


	log.Debugf("IP address found", map[string]interface{}{
		"MachineID": d.ServerId,
		"IP":        d.IPAddress,
	})
	
	return nil
}

// Create a new docker machine instance on ArubaCloud Cloud
func (d *Driver) Create() error {
	switch d.Action{
		case "NewSmart":
		err := d.CreateSmart()
		if err != nil {
			return err
		}
		case "NewPro":
		err := d.CreatePro()
		if err != nil {
			return err
		}
		case "Attach":
		err := d.Attach()
		if err != nil {
			return err
		}
	}
	
	

	// All done !
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func copySSHKey(src, dst string) error {
	if err := mcnutils.CopyFile(src, dst); err != nil {
		return fmt.Errorf("unable to copy ssh key: %s", err)
	}

	if err := os.Chmod(dst, 0600); err != nil {
		return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
	}

	return nil
}

func (d *Driver) createKeyPair() (string, error) {
	if len(d.SSHKey) > 0 {
		log.Debug("Importing Key Pair...", map[string]interface{}{"Path": d.SSHKey})
		keyfile := d.GetSSHKeyPath()
		log.Debug("keyfile: ", keyfile)
		keypath := filepath.Dir(keyfile)
		log.Debug("keypath: ", keypath)
		
		err := os.MkdirAll(keypath, 0700)
		if err != nil {
			return "", err
		}
		
		
		if err := copySSHKey(d.SSHKey, d.SSHKeyPath); err != nil {
			return "",err
		}
		if err := copySSHKey(d.SSHKey+".pub", d.SSHKeyPath+".pub"); err != nil {
			log.Infof("Couldn't copy SSH public key : %s", err)
			return "",err
		}
			
	} else {
		log.Debug("Creating Key Pair...")
		keyfile := d.GetSSHKeyPath()
		log.Debug("keyfile: ", keyfile)
		keypath := filepath.Dir(keyfile)
		log.Debug("keypath: ", keypath)
		err := os.MkdirAll(keypath, 0700)
		if err != nil {
			return "", err
		}
		
		err = ssh.GenerateSSHKey(d.GetSSHKeyPath())
		if err != nil {
			return "", err
		}

	}
	
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	log.Debugf("GetSSHHostname: ", d.IPAddress)
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	log.Debugf("Get status for ArubaCloud Server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	instance, err := client.GetServer(d.ServerId)
	if err != nil {
		return state.None, err
	}

	log.Debugf("ArubaCloud Server", map[string]interface{}{
		"MachineID": d.ServerId,
		"State":     instance.ServerStatus,
	})

	switch instance.ServerStatus {
	case 3:
		return state.Running, nil
	case 4:
		return state.Saved, nil
	case 2:
		return state.Stopped, nil
	case 1:
		return state.Starting, nil
	}

	return state.None, nil
}

func (d *Driver) Remove() error {
	log.Debugf("deleting server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	// Check the state of the Virtual Machine
	s, err := d.GetState()
	if err != nil { return err }
	if s == state.Running {
		client.StopServer(d.ServerId)
		_, err := d.waitForServerStatus(2)//2 = Stop --3 = Running --5 = Deleted
		if err != nil { return err }
	}

	// Deletes instance
	err = client.DeleteServer(d.ServerId)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	log.Debugf("starting server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	err := client.StartServer(d.ServerId)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() (err error) {
	log.Debugf("Stopping server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	// Check the state of the virtual machine
	s, err := d.GetState()
	if err != nil {
		return err
	}

	// Poweroff VM in case it's running
	if s == state.Running {
		client.StopServer(d.ServerId)
		_, err := d.waitForServerStatus(3)
		if err != nil { return err }
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if d.IPAddress == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(d.IPAddress, "2376")), nil
}

func (d *Driver) Restart() error {
	log.Debugf("restarting server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	// Poweroff the VM
	client.StopServer(d.ServerId)
	_, err := d.waitForServerStatus(2)
	if err != nil {
		return err
	}
	// Poweron the VM
	client.StartServer(d.ServerId)
	_, err = d.waitForServerStatus(3)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() (err error) {
	log.Debugf("Killing server...", map[string]interface{}{"MachineID": d.ServerId})

	client := d.getClient()

	// Check the state of the virtual machine
	s, err := d.GetState()
	if err != nil {
		return err
	}

	// Poweroff VM in case it's running
	if s == state.Running {
		client.KillServer(d.ServerId)
		_, err := d.waitForServerStatus(3)
		if err != nil { return err }
	}

	return nil
}
