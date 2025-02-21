#!/bin/bash
#
set -e 
docker rmi golang:alpine
docker rmi registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64
docker rmi registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64

# 拉取 amd64 架构的 Golang 镜像
docker pull --platform=linux/amd64 golang:alpine
docker tag golang:alpine registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64
docker rmi golang:alpine
# 拉取 arm64 架构的 Golang 镜像
docker pull --platform=linux/arm64 golang:alpine
docker tag golang:alpine registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64
docker rmi golang:alpine

# 创建多架构 Manifest List
docker manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine \
  registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64 \
  registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64

# 推送 Manifest List 到镜像仓库
docker push registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine


# 拉起alpine镜像
docker pull --platform=linux/amd64 alpine:latest
docker tag alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-amd64
docker pull --platform=linux/arm64 alpine:latest
docker tag alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-arm64
docker manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-arm64 registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-amd64 
docker push registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest