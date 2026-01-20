<div align="center">
<h1>K8M</h1>
</div>

<div align=center>
 
[![weibaohui%2Fk8m | Trendshift](https://trendshift.io/api/badge/repositories/14095)](https://trendshift.io/repositories/14095)

</div>

<div align=center>
 
![GitHub Repo Stars](https://img.shields.io/github/stars/weibaohui/k8m)
![GitHub Repo Forks](https://img.shields.io/github/forks/weibaohui/k8m)

</div>

<div align=center>

 [![License MIT](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/k8m/blob/master/LICENSE)
 [![Go Report Card](https://goreportcard.com/badge/github.com/weibaohui/k8m)](https://goreportcard.com/report/github.com/weibaohui/k8m)
![GitHub Release](https://img.shields.io/github/v/release/weibaohui/k8m)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/weibaohui/k8m/total)
![GitHub Repo Issues](https://img.shields.io/github/issues/weibaohui/k8m)
[![Trust Score](https://archestra.ai/mcp-catalog/api/badge/quality/weibaohui/k8m)](https://archestra.ai/mcp-catalog/weibaohui__k8m)
![Repobeats analytics image](https://repobeats.axiom.co/api/embed/9fde094e5c9a1d4c530e875864ee7919b17d0690.svg)

</div>


[English](README_en.md) | [中文](README.md)



**k8m** 是一款AI驱动的 Mini Kubernetes AI Dashboard 轻量级控制台工具，专为简化集群管理设计。它基于 AMIS 构建，并通过  [
`kom`](https://github.com/weibaohui/kom)  作为 Kubernetes API 客户端，**k8m** 内置了
Qwen2.5-Coder-7B，支持deepseek-ai/DeepSeek-R1-Distill-Qwen-7B模型
模型交互能力，同时支持接入您自己的私有化大模型（包括ollama）。

### 演示DEMO

[DEMO](http://107.150.119.151:3618)
[DEMO-InCluster模式](http://107.150.119.151:31999)
用户名密码 demo/demo

### 文档

- 详细的配置和使用说明请参考[文档](docs/README.md)。
- 更新日志请参考[更新日志](CHANGELOG.md)。
- DeepWiki 文档：[开发设计文档](https://deepwiki.com/weibaohui/k8m)

### 主要特点

- **迷你化设计**：所有功能整合在一个单一的可执行文件中，部署便捷，使用简单。
- **简便易用**：友好的用户界面和直观的操作流程，让 Kubernetes 管理更加轻松。支持标准k8s、aws eks、k3s、kind、k0s等集群类型。
- **高效性能**：后端采用 Golang 构建，前端基于百度 AMIS，保证资源利用率高、响应速度快。
- **AI驱动融合**
  ：基于ChatGPT实现划词解释、资源指南、YAML属性自动翻译、Describe信息解读、日志AI问诊、运行命令推荐,并集成了[k8s-gpt](https://github.com/k8sgpt-ai/k8sgpt)
  功能，实现中文展现，为管理k8s提供智能化支持。
- **功能插件化**：特性功能插件化，按需开启，不开启不占资源。  
- **MCP集成**:可视化管理MCP，实现大模型调用Tools，内置k8s多集群MCP工具49种，可组合实现超百种集群操作，可作为MCP Server
  供其他大模型软件使用。轻松实现大模型管理k8s。可详细记录每一次MCP调用。支持mcp.so主流服务。
- **MCP权限打通**:多集群管理权限与MCP大模型调用权限打通，一句话概述：谁使用大模型，就用谁的权限执行MCP。安全使用，无后顾之忧，避免操作越权。
- **多集群管理**：自动识别集群内部使用InCluster模式，配置kubeconfig路径后自动扫描同级目录下的配置文件，同时注册管理多个集群，支持心跳检测与自动重连。
- **多集群权限管理**：支持对用户、用户组进行授权，可按集群授权，包括集群只读、Exec命令、集群管理员三种权限。对用户组授权后，组内用户均获得相应授权。支持设置命名空间黑白名单。
- **支持k8s最新特性**:支持APIGateway、OpenKruise等功能特性。
- **Pod文件管理**：在Console 界面左侧的文件树，右键菜单，支持 Pod 内文件的浏览、编辑、上传、下载、删除，简化日常操作。
- **Pod运行管理**：支持实时查看 Pod 日志，下载日志，并在 Pod 内直接执行 Shell 命令。支持Ctrl+F搜索，类似grep -A -B高亮搜索
- **API开放**:支持创建API KEY，从第三方外部访问，提供swagger接口管理页面。
- **集群巡检支持**：支持多集群定时巡检、自定义巡检规则，支持lua脚本规则。支持发送到钉钉群、微信群、飞书群以及自定义webhook。支持AI总结。
- **k8s Event转发**：支持多集群k8s Event转发到webhook中，可按集群、关键字、命名空间、名称等进行过滤，建立多个专门的监控转发通道。支持AI总结。
- **CRD管理**：可自动发现并管理 CRD 资源，树状列出所有CRD，提高工作效率。
- **Helm市场**：支持Helm自由添加仓库，一键安装、卸载、升级 Helm 应用，支持自动更新。
- **跨平台支持**：兼容 Linux、macOS 和 Windows，并支持 x86、ARM 等多种架构，确保多平台无缝运行。
- **多数据库支持**：支持SQLite、MySql、PostgreSql等多种数据库。
- **完全开源**：开放所有源码，无任何限制，可自由定制和扩展，可商业使用。

**k8m** 的设计理念是"AI驱动，轻便高效，化繁为简"，它帮助开发者和运维人员快速上手，轻松管理 Kubernetes 集群。

![](https://github.com/user-attachments/assets/0951d6c1-389c-49cb-b247-84de15b6ec0e)


## **运行**

1. **下载**：从 [GitHub release](https://github.com/weibaohui/k8m/releases) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。
3. **登录用户名密码**：
    - 用户名：`k8m`
    - 密码：`k8m`
    - 请注意上线后修改用户名密码、启用两步验证。
4. **参数**：

```shell
Usage of ./k8m:
      --enable-temp-admin                是否启用临时管理员账户配置，默认关闭
      --admin-password string            管理员密码，启用临时管理员账户配置后生效 
      --admin-username string            管理员用户名，启用临时管理员账户配置后生效
      --print-config                     是否打印配置信息 (default false)
      --connect-cluster                  启动集群是是否自动连接现有集群，默认关闭
  -d, --debug                            调试模式
      --in-cluster                       是否自动注册纳管宿主集群，默认启用
      --jwt-token-secret string          登录后生成JWT token 使用的Secret (default "your-secret-key")
  -c, --kubeconfig string                kubeconfig文件路径 (default "/root/.kube/config")
      --kubectl-shell-image string       Kubectl Shell 镜像。默认为 bitnami/kubectl:latest，必须包含kubectl命令 (default "bitnami/kubectl:latest")
      --log-v int                        klog的日志级别klog.V(2) (default 2)
      --login-type string                登录方式，password, oauth, token等,default is password (default "password")
      --image-pull-timeout               Node Shell、Kubectl Shell 镜像拉取超时时间。默认为 30 秒
      --node-shell-image string          NodeShell 镜像。 默认为 alpine:latest，必须包含`nsenter`命令 (default "alpine:latest")
  -p, --port int                         监听端口 (default 3618)
  -v, --v Level                          klog的日志级别 (default 2)
```

也可以直接通过docker-compose(推荐)启动：

```yaml
services:
  k8m:
    container_name: k8m
    image: registry.cn-hangzhou.aliyuncs.com/minik8m/k8m
    restart: always
    ports:
      - "3618:3618"
    environment:
      TZ: Asia/Shanghai
    volumes:
      - ./data:/app/data
```

启动之后，访问`3618`端口，默认用户：`k8m`，默认密码`k8m`。
如果你想通过在线环境快速拉起体验，可以访问：[k8m](https://cnb.cool/znb/qifei/-/tree/main/letsfly/justforfun/k8m)


## 容器化k8s集群方式运行

使用[KinD](https://kind.sigs.k8s.io/docs/user/quick-start/)、[MiniKube](https://minikube.sigs.k8s.io/docs/start/)
安装一个小型k8s集群

## KinD方式

* 创建 KinD Kubernetes 集群

```
brew install kind
```

* 创建新的 Kubernetes 集群：

```
kind create cluster --name k8sgpt-demo
```

## 将k8m部署到集群中体验

### 安装脚本

```docker
kubectl apply -f https://raw.githubusercontent.com/weibaohui/k8m/refs/heads/main/deploy/k8m.yaml
```

* 访问：
  默认使用了nodePort开放，请访问31999端口。或自行配置Ingress
  http://NodePortIP:31999

## 生产部署启用主备选举插件，注意事项

- 单实例运行service的定义`不要加` `k8m.io/role: leader` 标签，加了不能正常访问。
- 多实例运行service的定义`必须加` `k8m.io/role: leader` 标签，否则不会切换。
- 多实例运行的yaml如下：
```docker
kubectl apply -f https://raw.githubusercontent.com/weibaohui/k8m/refs/heads/main/deploy/k8m-ms.yaml
```


## **ChatGPT 配置指南**

### 内置GPT

从v0.0.8版本开始，将内置GPT，无需配置。
如果您需要使用自己的GPT，请参考以下文档。

- [自托管/自定义大模型支持](docs/use-self-hosted-ai.md) - 如何使用自托管的
- [Ollama配置](docs/ollama.md) - 如何配置使用Ollama大模型。

### **ChatGPT 状态调试**

如果设置参数后，依然没有效果，请尝试使用`./k8m -v 6`获取更多的调试信息。
会输出以下信息，通过查看日志，确认是否启用ChatGPT。
  



## 开发调试

如果你想在本地开发调试，请先执行一次本地前端构建，自动生成dist目录。因为本项目采用了二进制嵌入，没有dist前端会报错。

#### 第一步编译前端

```bash 
cd ui
pnpm run build
```

#### 编译调试后端

```bash
#下载依赖
go mod tidy
#运行
air
#或者
go run *.go 
# 监听localhost:3618端口
```

#### 前端热加载

```bash
cd ui
pnpm run dev
#Vite服务会监听在localhost:3000端口
#Vite转发后端访问到3618端口
```

访问http://localhost:3000

### HELP & SUPPORT

如果你有任何进一步的问题或需要额外的帮助，请随时与我联系！

### 特别鸣谢

[zhaomingcheng01](https://github.com/zhaomingcheng01)：提出了诸多非常高质量的建议，为k8m的易用好用做出了卓越贡献~

[La0jin](https://github.com/La0jin):提供在线资源及维护，极大提升了k8m的展示效果

[eryajf](https://github.com/eryajf):为我们提供了非常好用的github actions，为k8m增加了自动化的发版、构建、发布等功能

## 联系我

微信（大罗马的太阳） 搜索ID：daluomadetaiyang,备注k8m。
<br><img width="214" alt="Image" src="https://github.com/user-attachments/assets/166db141-42c5-42c4-9964-8e25cf12d04c" />

## 微信群
![输入图片说明](https://foruda.gitee.com/images/1768919272453116492/7d2cfe39_77493.png "屏幕截图")
