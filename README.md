System Management - Pharos Node
=======================================

This is intented to be installed in each of Edge devices, which communicates with centralized service deployment management, called Pharos Anchor, running in a management server. Once Pharos Anchor sends a request of service deployment to a certain Edge device, the corresponding Pharos Node performs one of Docker operations to pull, run, and stop containers as requested.

## Prerequisites ##
- docker-ce
    - Version: 17.09
    - [How to install](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)
- go compiler
    - Version: 1.8 or above
    - [How to install](https://golang.org/dl/)

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
    - ANCHOR_ADDRESS='...'
    - NODE_ADDRESS='...'
- volume
    - "host folder"/data/db:/data/db (Note that you should replace "host folder" to a desired folder on your host machine)

You can execute it with a Docker image as follows:
```shell
$ docker run -it -p 48098:48098 -e ANCHOR_ADDRESS='...' -e NODE_ADDRESS='...' -v /data/db:/data/db -v /var/run/docker.sock:/var/run/docker.sock system-pharos-node-go-ubuntu
```
If it succeeds, you can see log messages on your screen as follows:
```shell
$ docker run -it -p 48098:48098 -v /data/db:/data/db system-pharos-node-go-ubuntu
[DEBUG][MA]2018/01/17 10:20:31 controller/health registration.go register : 66 [IN]
[DEBUG][MA]2018/01/17 10:20:31 controller/configuration.Executor configuration.go GetConfiguration : 76 [Configuration file is not found.]
[ERROR][MA]2018/01/17 10:20:31 controller/health registration.go register : 71 [not find target : ./configuration.json]
[DEBUG][MA]2018/01/17 10:20:31 controller/health registration.go register : 72 [OUT]
[ERROR][MA]2018/01/17 10:20:31 controller/health.init registration.go 0 : 57 [not find target : ./configuration.json]
[DEBUG][MA]2018/01/17 10:20:31 main main.go main : 25 [Start Pharos Node]
[DEBUG][MA]2018/01/17 10:20:31 api restapi.go RunNodeWebServer : 37 [Start Pharos Node Web Server]
[DEBUG][MA]2018/01/17 10:20:31 api restapi.go RunNodeWebServer : 38 [Listening 0.0.0.0:48098]
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] MongoDB starting : pid=7 port=27017 dbpath=/data/db 64-bit host=0c083c53cb38
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] db version v3.4.4
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] git version: 888390515874a9debd1b6c5d36559ca86b44babd
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] OpenSSL version: LibreSSL 2.5.5
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] allocator: system
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] modules: none
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] build environment:
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten]     distarch: x86_64
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten]     target_arch: x86_64
2018-01-17T10:20:31.902+0000 I CONTROL  [initandlisten] options: { storage: { journal: { enabled: true }, mmapv1: { smallFiles: true } } }
2018-01-17T10:20:31.902+0000 W -        [initandlisten] Detected unclean shutdown - /data/db/mongod.lock is not empty.
2018-01-17T10:20:31.911+0000 I -        [initandlisten] Detected data files in /data/db created by the 'wiredTiger' storage engine, so setting the active storage engine to 'wiredTiger'.
2018-01-17T10:20:31.911+0000 W STORAGE  [initandlisten] Recovering data from the last clean checkpoint.
2018-01-17T10:20:31.911+0000 I STORAGE  [initandlisten] 
2018-01-17T10:20:31.911+0000 I STORAGE  [initandlisten] ** WARNING: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine
2018-01-17T10:20:31.911+0000 I STORAGE  [initandlisten] **          See http://dochub.mongodb.org/core/prodnotes-filesystem
2018-01-17T10:20:31.911+0000 I STORAGE  [initandlisten] wiredtiger_open config: create,cache_size=11515M,session_max=20000,eviction=(threads_min=4,threads_max=4),config_base=false,statistics=(fast),log=(enabled=true,archive=true,path=journal,compressor=snappy),file_manager=(close_idle_time=100000),checkpoint=(wait=60,log_size=2GB),statistics_log=(wait=0),
2018-01-17T10:20:32.359+0000 W STORAGE  [initandlisten] Detected configuration for non-active storage engine mmapv1 when current storage engine is wiredTiger
2018-01-17T10:20:32.359+0000 I CONTROL  [initandlisten] 
2018-01-17T10:20:32.359+0000 I CONTROL  [initandlisten] ** WARNING: Access control is not enabled for the database.
2018-01-17T10:20:32.359+0000 I CONTROL  [initandlisten] **          Read and write access to data and configuration is unrestricted.
2018-01-17T10:20:32.359+0000 I CONTROL  [initandlisten] ** WARNING: You are running this process as the root user, which is not recommended.
2018-01-17T10:20:32.359+0000 I CONTROL  [initandlisten] 
2018-01-17T10:20:32.363+0000 I FTDC     [initandlisten] Initializing full-time diagnostic data capture with directory '/data/db/diagnostic.data'
2018-01-17T10:20:32.363+0000 I NETWORK  [thread1] waiting for connections on port 27017
2018-01-17T10:20:33.017+0000 I FTDC     [ftdc] Unclean full-time diagnostic data capture shutdown detected, found interim file, some metrics may have been lost. OK
^C2018-01-17T10:20:39.716+0000 I CONTROL  [signalProcessingThread] got signal 2 (Interrupt), will terminate after current cmd ends
2018-01-17T10:20:39.716+0000 I NETWORK  [signalProcessingThread] shutdown: going to close listening sockets...
2018-01-17T10:20:39.717+0000 I NETWORK  [signalProcessingThread] closing listening socket: 6
2018-01-17T10:20:39.717+0000 I NETWORK  [signalProcessingThread] closing listening socket: 7
2018-01-17T10:20:39.717+0000 I NETWORK  [signalProcessingThread] removing socket file: /tmp/mongodb-27017.sock
2018-01-17T10:20:39.717+0000 I NETWORK  [signalProcessingThread] shutdown: going to flush diaglog...
2018-01-17T10:20:39.717+0000 I FTDC     [signalProcessingThread] Shutting down full-time diagnostic data capture

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
