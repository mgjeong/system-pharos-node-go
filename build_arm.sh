###############################################################################
# Copyright 2018 Samsung Electronics All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###############################################################################
#!/bin/bash

echo -e "\n\033[33m"Start building of Pharos-Node"\033[0m"
export GOPATH=$PWD

function func_cleanup(){
    rm -rf $GOPATH/pkg
    rm -rf $GOPATH/src/docker.io
    rm -rf $GOPATH/src/golang.org
    rm -rf $GOPATH/src/github.com
}


function build(){
    CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o pharos-node -a -ldflags '-extldflags "-static"'  src/main/main.go
    if [ $? -ne 0 ]; then
        echo -e "\n\033[31m"build fail"\033[0m"
        func_cleanup
        exit 1
    fi
}

function download_pkgs(){
    pkg_list=(
        "gopkg.in/yaml.v2"
        "-d docker.io/go-docker"
        "-d github.com/docker/libcompose"
        "golang.org/x/sys/unix"
        "github.com/shirou/gopsutil"
        "github.com/boltdb/bolt"
        )


    idx=1
    for pkg in "${pkg_list[@]}"; do
        echo -ne "(${idx}/${#pkg_list[@]}) go get $pkg"
        go get $pkg
        if [ $? -ne 0 ]; then
            echo -e "\n\033[31m"download fail"\033[0m"
            func_cleanup
            exit 1
        fi
        echo ": Done"
        idx=$((idx+1))
    done
    rm -rf $GOPATH/src/github.com/docker/distribution/vendor/github.com/opencontainers
}

echo -e "\nDownload dependent go-pkgs"
download_pkgs

echo -ne "\nMaking executable file of Pharos-Node service"
build
echo ": Done"

echo -ne "\nPost processing"
func_cleanup
echo ": Done"

echo -e "\n\033[33m"Succeed build of Pharos-Node"\033[0m"
