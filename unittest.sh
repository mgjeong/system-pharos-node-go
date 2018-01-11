#!/bin/bash

export GOPATH=$PWD

go get github.com/golang/mock/gomock
go get github.com/ghodss/yaml
go get -d docker.io/go-docker
go get -d github.com/docker/libcompose

rm -rf $GOPATH/src/github.com/docker/distribution/vendor/github.com/opencontainers

pkg_list=("api/common" "api/deployment" "api/health" "api/resource" "controller/deployment" "controller/deployment/dockercontroller" "controller/health" "controller/resource" "controller/shellcommand" "db/mongo/service" "messenger")

function func_cleanup(){
    rm *.out *.test
    rm -rf $GOPATH/pkg
    rm -rf $GOPATH/src/docker.io
    rm -rf $GOPATH/src/golang.org
    rm -rf $GOPATH/src/github.com
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
