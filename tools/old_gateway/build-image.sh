#!/bin/sh
#直接在本地编译,也可以直接将二进制文件放到镜像中运行
export GO111MODULE=auto && export GOPROXY=https://goproxy.cn && go mod tidy
GOOS=linux GOARCH=amd64 go build -o ./bin/gateway
docker build -f Dockerfile-market -t gateway-market .
docker build -f Dockerfile-proxy -t gateway-proxy .
