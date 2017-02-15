# Aruba Cloud Driver for Docker Machine

## Table of Contents
* [Overview](#overview)
* [Requirements](#requirements)
* [Installation](#installation)
  * [From a Release](#from-a-release)
  * [From Source](#from-source)
* [Usage](#usage)
  * [Available Options](#available-options)
  * [Example](#example)
* [License](#license)

## Overview

The Aruba Cloud Driver is a plugin for Docker Machine which allows you to automate the provisioning of Docker hosts on Aruba Cloud Servers. The plugin is based on the [Go Aruba Cloud SDK](https://github.com/Arubacloud/goarubacloud) and [Cloud API](http://kb.cloud.it/en/api.aspx). 

To acquire Aruba Cloud Cloud API credentials visit https://www.cloud.it.

## Requirements

  * [Docker Machine](https://docs.docker.com/machine/install-machine/) 0.9.0 or a newer version

Windows and Mac OS X users may install [Docker Toolbox](https://www.docker.com/products/docker-toolbox) package that includes the latest version of the Docker Machine.

## Installation

### From a Release

The latest version of the `docker-machine-driver-arubacloud` binary is available on the [GithHub Releases](https://github.com/Arubacloud/docker-machine-driver-arubacloud/releases) page.
Download the `tar` archive and extract it into a directory residing in your PATH. Select the binary that corresponds to your OS and according to the file name prefix:

* Linux: docker-machine-driver-arubacloud-linux
* Mac OS X: docker-machine-driver-arubacloud-darwin
* Windows: docker-machine-driver-arubacloud-windows

To extract and install the binary, Linux and Mac users can use the Terminal and the following commands:

```bash
sudo tar -C /usr/local/bin -xvzf docker-machine-driver-arubacloud*.tar.gz
```

If required, modify the permissions to make the plugin executable:

```bash
sudo chmod +x /usr/local/bin/docker-machine-driver-arubacloud
```

Windows users may run the above commands without `sudo` in Docker Quickstart Terminal that is installed with [Docker Toolbox](https://www.docker.com/products/docker-toolbox).

### From Source

Make sure you have installed [Go](http://www.golang.org) and configured [GOPATH](http://golang.org/doc/code.html#GOPATH) properly.

To download the repository and build the driver run the following:

```bash
go get -d -u github.com/Arubacloud/docker-machine-driver-arubacloud
cd $GOPATH/src/github.com/Arubacloud/docker-machine-driver-arubacloud
make build
```

To use the driver run:

```bash
make install
```

This command will install the driver into `/usr/local/bin`. 

Otherwise, set your PATH environment variable correctly. For example:

```bash
export PATH=$GOPATH/src/github.com/Arubacloud/docker-machine-driver-arubacloud/bin:$PATH
```

If you are running Windows, you may also need to install GNU Make, Bash shell and a few other Bash utilities available with [Cygwin](https://www.cygwin.com).

## Usage

You may want to refer to the Docker Machine [official documentation](https://docs.docker.com/machine/) before using the driver.

Verify that Docker Machine can see the Aruba Cloud driver:

```bash
docker-machine create -d arubacloud --help
```


### Available Options

  * `--ac_username`:  Aruba Cloud username.
  * `--ac_password`: Aruba Cloud password.
  * `--ac_admin_password`: Virtual machine admin password.
  * `--ac_endpoint`: Aruba Cloud Data Center (dc1,dc2,dc3,etc.).
  * `--ac_template`: Virtual machine template.
  * `--ac_size`: Size of the virtual machine.


|          CLI Option             |Default Value 	| Environment Variable           | Required |
| --------------------------------|--------------------| ------------------------------ | -------- |
| `--ac_username`	          |			     | `AC_USERNAME`            		| yes      |
| `--ac_password`       	   |		    | `AC_PASSWORD`         		 	| yes      |
| `--ac_admin_password`        	   |			| `AC_ADMIN_PASSWORD`            | no      |
| `--ac_endpoint`                  |`dc1`	| `AC_ENDPOINT`               	| yes      |
| `--ac_template`         	|`ubuntu1604_x64_1_0`	   | `AC_TEMPLATE`         		 	| yes      |
| `--ac_size`    		|`Large`   		   | `AC_SIZE`       				| yes      |

Valid values for `--ac_size` are `Small`, `Medium`, `Large`, `Extra Large`.

Available parameters for `--ac_endpoint` are shown in the next table.

| Parameter |                 Data Center Location                 |
|-----------|------------------------------------------------------|
| `dc1`     | Italy 1                                              |
| `dc2`     | Italy 2                                              |
| `dc3`     | Czech republic 									   |
| `dc4`     | France                                               |
| `dc5`     | Deutschland                             			   |
| `dc6`     | United Kingdom                            		|

Supported values for `--ac_template` are listed below.

|              Parameter                |					OS						|
|---------------------------------------|-------------------------------------------|
| `centos7_x64_1_0`                     | `CentOS 7.x 64bit`						|
| `debian8_x64_1_0`                     | `Debian 8 64bit`							|
| `ubuntu1604_x64_1_0`                	| `Ubuntu Server 16.04 LTS 64bit`			|
| `BSD-001-freebsd_x64_1_0`      		| `FreeBSD 10.x 64bit`						|
| `LO12-002_OpenSuse_12_x64_1_0`    	| `openSuse 12.1 64bit`						|
| `WS12-002_W2K12R2_1_0`                | `Windows 2012 R2 64bit`					|

 
### Examples

#### Create using defaults:

```
docker-machine --debug create --driver arubacloud \
 --ac_username		              "ARU-XXXX" \
 --ac_password			          "xxxxxxx" \
MyDockerHostName
```

#### Create specifying template, endpoint and size:

```
docker-machine --debug create --driver arubacloud \
 --ac_username		              "ARU-XXXX" \
 --ac_password			          "xxxxxxx" \
 --ac_endpoint			          "dc1" \
 --ac_template	                  "ubuntu1404_x64_1_0" \
 --ac_size				          "Large" \
 --ac_admin_password		      "yyyyyyyy" \  
MyDockerHostName
```
####View new instance

Go to Aruba Cloud dashboard to view new instance. 
Dashboard url is different depending on the selected endpoint:

|              Dashboard                					|
|-----------------------------------------------------------|
|[DC1](https://admin.dc1.computing.cloud.it/Login.aspx)		|
|[DC2](https://admin.dc2.computing.cloud.it/Login.aspx)		|
|[DC3](https://admin.dc3.computing.cloud.it/Login.aspx)		|
|[DC4](https://admin.dc4.computing.cloud.it/Login.aspx)		|
|[DC5](https://admin.dc5.computing.cloud.it/Login.aspx)		|
|[DC6](https://admin.dc6.computing.cloud.it/Login.aspx)		|

####View instance list

```
docker-machine ls
NAME    			ACTIVE   	DRIVER       STATE   	URL   				SWARM   	DOCKER    ERRORS
MyDockerHostName   	*        	arubacloud   Running    tcp://10.254.4.232             	v1.10.0   
default   			-       	arubacloud   Running    tcp://10.254.4.232             	v1.10.0   

```



## License

This code is released under the Apache 2.0 License.

Copyright (c) 2017 Aruba Cloud
