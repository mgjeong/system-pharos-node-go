#!/bin/bash
export GOPATH=$PWD
export ANCHOR_ADDRESS=127.0.0.1
export NODE_ADDRESS=127.0.0.1

go get github.com/golang/mock/gomock
go get github.com/ghodss/yaml
go get -d docker.io/go-docker
go get -d github.com/docker/libcompose
go get golang.org/x/sys/unix
go get github.com/shirou/gopsutil
go get github.com/boltdb/bolt

rm -rf $GOPATH/src/github.com/docker/distribution/vendor/github.com/opencontainers

pkg_list=("api" "api/common" "api/deployment" "api/health" "api/monitoring/resource" "api/configuration" "api/notification" "api/notification/apps" "commons/errors" "commons/logger" "commons/url" "commons/util" "controller/deployment" "controller/dockercontroller" "controller/health" "controller/monitoring/resource" "controller/monitoring/apps" "controller/configuration" "controller/shellcommand" "controller/monitoring/apps" "controller/notification/apps" "db/bolt/event" "db/bolt/configuration" "db/bolt/service" "messenger")

function func_cleanup(){
    rm *.out *.test
    rm -rf $GOPATH/pkg
    rm -rf $GOPATH/src/docker.io
    rm -rf $GOPATH/src/golang.org
    rm -rf $GOPATH/src/github.com
    unset ANCHOR_ADDRESS
    unset NODE_ADDRESS
}

count=0
for pkg in "${pkg_list[@]}"; do
 go test -c -v -gcflags "-N -l" $pkg
 go test -coverprofile=$count.cover.out $pkg
 if [ $? -ne 0 ]; then
    echo "Unittest is failed."
    func_cleanup
    exit 1
 fi
 count=$count.0
done

echo "mode: set" > coverage.out && cat *.cover.out | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> coverage.out

go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverall.html

func_cleanup
