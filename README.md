System Management - Pharos Node
=======================================

This is intented to be installed in each of Edge devices, which communicates with centralized service deployment management, called Pharos Anchor, running in a management server. Once Pharos Anchor sends a request of service deployment to a certain Edge device, the corresponding Pharos Node performs one of Docker operations to pull, run, and stop containers as requested.

## Quick start ##
This provides how to download and run pre-built Docker image without building project.

#### 1. Install docker-ce ####
- docker-ce
  - Version: 17.09
  - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)

#### 2. Download Docker image ####
Please visit [Downloads-ubuntu](https://github.sec.samsung.net/RS7-EdgeComputing/system-pharos-node-go/releases/download/alpha-1.1_rel/pharos_node_ubuntu_x86_64.tar)

#### 3. Load Docker image from tar file ####
```shell
$ docker load -i pharos_node_ubuntu_x86_64.tar
```
If it succeeds, you can see the Docker image as follows:
```shell
$ sudo docker images
REPOSITORY                                                                TAG      IMAGE ID        CREATED        SIZE
docker.sec.samsung.net:5000/edge/system-pharos-node-go/ubuntu_x86_64      alpha    534169f4035c    7 weeks ago    166MB
```
Note that, you can find other docker image, [Downloads-rpi_arm64](https://github.sec.samsung.net/RS7-EdgeComputing/system-pharos-node-go/releases/download/alpha-1.1_rel/pharos-node-rpi3-beluga.tar) and [Downloads-rpi-arm32](https://github.sec.samsung.net/RS7-EdgeComputing/system-pharos-node-go/releases/download/alpha-1.1_rel/pharos-node-artik530-beluga.tar)

#### 4. Run with Docker image ####
You can execute it with a Docker image as follows:
```shell
$ docker run -it \
	-p 48098:48098 \
	-e ANCHOR_ADDRESS='change_it_to_actual_anchor_address' \
	-e NODE_ADDRESS='change_it_to_actual_node_address' \
	-v /data/db:/data/db \
	-v /var/run/docker.sock:/var/run/docker.sock \
	docker.sec.samsung.net:5000/edge/system-pharos-node-go/ubuntu_x86_64:alpha
```

## Build Prerequisites ##
- docker-ce
  - Version: 17.09
  - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)
- go compiler
  - Version: 1.8 or above
  - [How to install](https://golang.org/dl/)
- Rasberry Pi3 (Optional)
  - [How to install RPi OS (Raspbian)](https://www.raspberrypi.org/documentation/installation/installing-images/)
  - [How to configure network settings on your RPi3 - Useful link](https://kr.mathworks.com/help/supportpkg/raspberrypi/ug/getting-the-raspberry_pi-ip-address.html?requestedDomain=true)
  - [How to install Docker to RPi3 (CPU arch: Armhf)](https://docs.docker.com/install/linux/docker-ce/debian/#prerequisites)
  - [How to access insecure registry for Docker image](https://docs.docker.com/registry/insecure/#deploy-a-plain-http-registry)

## How to build ##
This provides how to build sources codes to an excutable binary and dockerize it to create a Docker image.

#### 1. Executable binary ####
```shell
$ ./build.sh
```
If source codes are successfully built, you can find an output binary file, **main**, on a root of project folder.
Note that, you can find other build scripts, **build_arm.sh** and **build_arm64**, which can be used to build the codes for ARM and ARM64 machines, respectively.

#### 2. Docker Image  ####
Next, you can create it to a Docker image.
```shell
$ docker build -t system-pharos-node-go-ubuntu -f Dockerfile .
```
If it succeeds, you can see the built image as follows:
```shell
$ sudo docker images
REPOSITORY                         TAG        IMAGE ID        CREATED           SIZE
system-pharos-node-go-ubuntu       latest     fcbbd4c401c2    31 seconds ago    157MB
```
Note that, you can find other Dockerfiles, **Dockerfile_arm** and **Dockerfile_arm64**, which can be used to dockerize for ARM and ARM64 machines, respectively.

## How to run with Docker image ##
Required options to run Docker image
- port
    - 48098:48098
- environment variables
    - [Mandatory] ANCHOR_ADDRESS='...'
    - [Mandatory] NODE_ADDRESS='...'
    - [Optional] REVERSE_PROXY=true/false
    - [Optional] ANCHOR_REVERSE_PROXY=true/false
    - [Optional] DEVICE_ID='...'
    - [Optional] DEVICE_NAME='...'
- volume
    - "host folder"/data/db:/data/db (Note that you should replace "host folder" to a desired folder on your host machine)

You can execute it with a Docker image as follows:
```shell
$ docker run -it \
	-p 48098:48098 \
	-e ANCHOR_ADDRESS='...' \
	-e NODE_ADDRESS='...' \
	-e REVERSE_PROXY='true/false' \
	-e ANCHOR_REVERSE_PROXY='true/false' \
	-e DEVICE_ID='...' \
	-e DEVICE_NAME='...' \
	-v /data/db:/data/db \
	-v /var/run/docker.sock:/var/run/docker.sock \
	system-pharos-node-go-ubuntu
```
If it succeeds, you can see log messages on your screen as follows:
```shell
$ docker run -it -p 48088:48098 -e ANCHOR_ADDRESS=10.113.64.134 -e NODE_ADDRESS=10.113.64.134 -v /pharos-node/data/db:/data/db -v /var/run/docker.sock:/var/run/docker.sock system-pharos-node-go-ubuntu
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 135 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 140 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 135 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 140 [OUT]
[ERROR][NODE]2018/07/04 07:38:40 controller/configuration configuration.go getProxyInfo : 216 [No reverse proxy environment]
[ERROR][NODE]2018/07/04 07:38:40 controller/configuration configuration.go initConfiguration : 109 [unknown error : No reverse proxy environment]
[DEBUG][NODE]2018/07/04 07:38:40 controller/dockercontroller.dockerExecutorImpl dockerexecutor.go Info : 151 []
[DEBUG][NODE]2018/07/04 07:38:40 controller/dockercontroller.dockerExecutorImpl dockerexecutor.go Info : 173 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 135 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 140 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 113 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/deployment deploymentcontroller.go restoreAllAppsState : 986 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/service.Executor service.go GetAppList : 155 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/service.Executor service.go GetAppList : 171 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/deployment deploymentcontroller.go restoreAllAppsState : 1014 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health registration.go register : 96 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 controller/configuration.Executor configuration.go GetConfiguration : 148 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperties : 155 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperties : 171 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/configuration.Executor configuration.go GetConfiguration : 168 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/service.Executor service.go GetAppList : 155 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/service.Executor service.go GetAppList : 171 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health registration.go sendRegisterRequest : 182 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go MakeAnchorRequestUrl : 102 [http://10.113.64.134:48099/api/v1/management/nodes/register]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertMapToJson : 56 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertMapToJson : 63 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health registration.go sendRegisterRequest : 195 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertJsonToMap : 41 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertJsonToMap : 49 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 135 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 148 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 89 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go SetProperty : 128 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go startHealthCheck : 31 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 controller/configuration.Executor configuration.go GetConfiguration : 148 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperties : 155 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperties : 171 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/configuration.Executor configuration.go GetConfiguration : 168 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go startHealthCheck : 73 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health registration.go register : 150 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 main main.go main : 25 [Start Pharos Node]
[DEBUG][NODE]2018/07/04 07:38:40 api restapi.go RunNodeWebServer : 39 [Start Pharos Node Web Server]
[DEBUG][NODE]2018/07/04 07:38:40 api restapi.go RunNodeWebServer : 40 [Listening 0.0.0.0:48098]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go sendPingRequest : 81 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 135 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 db/bolt/configuration.Executor configuration.go GetProperty : 148 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertMapToJson : 56 [IN]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go ConvertMapToJson : 63 [OUT]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go sendPingRequest : 100 [try to send ping request]
[DEBUG][NODE]2018/07/04 07:38:40 commons/util util.go MakeAnchorRequestUrl : 102 [http://10.113.64.134:48099/api/v1/management/nodes/4bd8554a-c9d2-4b06-82d9-c4231fb326af/ping]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go sendPingRequest : 114 [receive pong response, code[200]]
[DEBUG][NODE]2018/07/04 07:38:40 controller/health checks.go sendPingRequest : 115 [OUT]

```

## (Optional) How to enable QEMU environment on your computer
QEMU could be useful if you want to test your implemetation on various CPU architectures(e.g. ARM, ARM64) but you have only Ubuntu PC. To enable QEMU on your machine, please do as follows.

Required packages for QEMU:
```shell
$ apt-get install -y qemu-user-static binfmt-support
```
For ARM 32bit:
```shell
$ echo ':arm:M::\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:/usr/bin/qemu-arm-static:' > /proc/sys/fs/binfmt_misc/register <br />
```
For ARM 64bit:
```shell
$ echo ':aarch64:M::\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7:\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff:/usr/bin/qemu-aarch64-static:' > /proc/sys/fs/binfmt_misc/register <br />
```

Now, you can build your implementation and dockerize it for ARM and ARM64 on your Ubuntu PC. The below is an example for ARM build.

```shell
$ ./build_arm.sh
$ docker build -t system-pharos-node-go-arm -f Dockerfile_arm .
```
