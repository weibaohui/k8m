#!/bin/bash
#
set -e 
podman rmi golang:alpine
podman rmi registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64
podman rmi registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64

# 拉取 amd64 架构的 Golang 镜像
podman pull --platform=linux/amd64 golang:alpine
podman tag golang:alpine registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64
podman rmi golang:alpine
# 拉取 arm64 架构的 Golang 镜像
podman pull --platform=linux/arm64 golang:alpine
podman tag golang:alpine registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64
podman rmi golang:alpine

# 创建多架构 Manifest List
podman manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine \
  registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-amd64 \
  registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine-arm64

# 推送 Manifest List 到镜像仓库
podman push registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine


# 拉起alpine镜像
podman pull --platform=linux/amd64 alpine:latest
podman tag alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-amd64
podman pull --platform=linux/arm64 alpine:latest
podman tag alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-arm64
podman manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-arm64 registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest-amd64 
podman push registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest

# 拉起bitnami/kubectl:latest镜像
podman pull --platform=linux/amd64 bitnami/kubectl:latest
podman tag bitnami/kubectl:latest registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest-amd64
podman pull --platform=linux/arm64 bitnami/kubectl:latest
podman tag bitnami/kubectl:latest registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest-arm64
podman manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest-arm64 registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest-amd64 
podman push registry.cn-hangzhou.aliyuncs.com/minik8m/kubectl:latest
