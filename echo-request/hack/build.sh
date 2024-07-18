#!/usr/bin/env bash

set +x

go build -o echo-request main.go
docker build -t registry.lqingcloud.cn/library/echo-request:latest -f Dockerfile ./
docker push registry.lqingcloud.cn/library/echo-request:latest
docker rmi registry.lqingcloud.cn/library/echo-request:latest
