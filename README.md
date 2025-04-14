<div align="center">
<h1>K8M</h1>
</div>


[English](README_en.md) | [中文](README.md)

[![k8m](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/k8m/blob/master/LICENSE)

**k8m** 是一款AI驱动的 Mini Kubernetes AI Dashboard 轻量级控制台工具，专为简化集群管理设计。它基于 AMIS 构建，并通过  [
`kom`](https://github.com/weibaohui/kom)  作为 Kubernetes API 客户端，**k8m** 内置了
Qwen2.5-Coder-7B，支持deepseek-ai/DeepSeek-R1-Distill-Qwen-7B模型
模型交互能力，同时支持接入您自己的私有化大模型。

### 演示DEMO

[DEMO](http://107.150.119.151:3618)
用户名密码 demo/demo

### 文档

详细的配置和使用说明请参考[文档](docs/README.md)。
更新日志请参考[更新日志](docs/changelog.md)。


### 主要特点

- **迷你化设计**：所有功能整合在一个单一的可执行文件中，部署便捷，使用简单。
- **简便易用**：友好的用户界面和直观的操作流程，让 Kubernetes 管理更加轻松。
- **高效性能**：后端采用 Golang 构建，前端基于百度 AMIS，保证资源利用率高、响应速度快。
- **AI驱动融合**：基于ChatGPT实现划词解释、资源指南、YAML属性自动翻译、Describe信息解读、日志AI问诊、运行命令推荐,并集成了[
  `k8s-gpt`](https://github.com/k8sgpt-ai/k8sgpt)功能，实现中文展现，为管理k8s提供智能化支持。
- **MCP集成**:可视化管理MCP，实现大模型调用Tools，内置k8s多集群MCP工具49种，可组合实现超百种集群操作，可作为MCP Server
  供其他大模型软件使用。轻松实现大模型管理k8s。支持mcp.so主流服务。
- **MCP权限打通**:多集群管理权限与MCP大模型调用权限打通，一句话概述：谁使用大模型，就用谁的权限执行MCP。安全使用，无后顾之忧，避免操作越权。
- **多集群管理**：自动识别集群内部使用InCluster模式，配置kubeconfig路径后自动扫描同级目录下的配置文件，同时注册管理多个集群。
- **多集群权限管理**：支持对用户、用户组进行授权，可按集群授权，包括集群只读、Exec命令、集群管理员三种权限。对用户组授权后，组内用户均获得相应授权。
- **Pod 文件管理**：支持 Pod 内文件的浏览、编辑、上传、下载、删除，简化日常操作。
- **Pod 运行管理**：支持实时查看 Pod 日志，下载日志，并在 Pod 内直接执行 Shell 命令。
- **CRD 管理**：可自动发现并管理 CRD 资源，提高工作效率。
- **Helm 市场**：支持Helm自由添加仓库，一键安装、卸载、升级 Helm 应用。
- **跨平台支持**：兼容 Linux、macOS 和 Windows，并支持 x86、ARM 等多种架构，确保多平台无缝运行。
- **完全开源**：开放所有源码，无任何限制，可自由定制和扩展，可商业使用。

**k8m** 的设计理念是“AI驱动，轻便高效，化繁为简”，它帮助开发者和运维人员快速上手，轻松管理 Kubernetes 集群。

![](https://github.com/user-attachments/assets/0951d6c1-389c-49cb-b247-84de15b6ec0e)

## **运行**

1. **下载**：从 [GitHub](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。
3. **参数**：

```shell
Usage of ./k8m:
      --admin-password string            管理员密码，启用临时管理员账户配置后生效 (default "123456")
      --admin-username string            管理员用户名，启用临时管理员账户配置后生效 (default "admin")      --any-select                       是否开启任意选择划词解释，默认开启 (default true)
      --print-config                     是否打印配置信息 (default false)
  -k, --chatgpt-key string               大模型的自定义API Key (default "sk-xxxxxxx")
  -m, --chatgpt-model string             大模型的自定义模型名称 (default "Qwen/Qwen2.5-7B-Instruct")
  -u, --chatgpt-url string               大模型的自定义API URL (default "https://api.siliconflow.cn/v1")
      --connect-cluster                  启动集群是是否自动连接现有集群，默认关闭
  -d, --debug                            调试模式
      --enable-ai                        是否启用AI功能，默认开启 (default true)
      --enable-temp-admin                是否启用临时管理员账户配置，默认关闭
      --in-cluster                       是否自动注册纳管宿主集群，默认启用
      --jwt-token-secret string          登录后生成JWT token 使用的Secret (default "your-secret-key")
  -c, --kubeconfig string                kubeconfig文件路径 (default "/root/.kube/config")
      --kubectl-shell-image string       Kubectl Shell 镜像。默认为 bitnami/kubectl:latest，必须包含kubectl命令 (default "bitnami/kubectl:latest")
      --log-v int                        klog的日志级别klog.V(2) (default 2)
      --login-type string                登录方式，password, oauth, token等,default is password (default "password")
      --node-shell-image string          NodeShell 镜像。 默认为 alpine:latest，必须包含`nsenter`命令 (default "alpine:latest")
  -p, --port int                         监听端口 (default 3618)
      --sqlite-path string               sqlite数据库文件路径， (default "./data/k8m.db")
  -s, --mcp-server-port int              MCP Server 监听端口，默认3619 (default 3619)
      --use-builtin-model                是否使用内置大模型参数，默认开启 (default true)
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
      - "3619:3619"
    environment:
      TZ: Asia/Shanghai
      ENABLE_TEMP_ADMIN: true
      ADMIN_USERNAME: admin
      ADMIN_PASSWORD: 123456
      #启用AI
      ENABLE_AI: true
      #关闭内置大模型
      USE_BUILTIN_MODEL: false
      # 设置私有化大模型
      OPENAI_API_KEY: sk-xxxxxxxx
      OPENAI_API_URL: https://api.siliconflow.cn/v1
      OPENAI_MODEL: Qwen/Qwen2.5-7B-Instruct
    volumes:
      - ./data:/app/data
```

启动之后，访问`3618`端口，默认用户：`admin`，默认密码`123456`。

如果你想通过在线环境快速拉起体验，可以访问：[k8m](https://cnb.cool/znb/qifei/-/tree/main/letsfly/justforfun/k8m)
，FORK仓库之后，拉起体验。

## **ChatGPT 配置指南**

### 内置GPT

从v0.0.8版本开始，将内置GPT，无需配置。
如果您需要使用自己的GPT，请参考以下步骤。

### **环境变量配置**

需要设置环境变量，以启用ChatGPT。

```bash
export OPENAI_API_KEY="sk-XXXXX"
export OPENAI_API_URL="https://api.siliconflow.cn/v1"
export OPENAI_MODEL="Qwen/Qwen2.5-7B-Instruct"
```

### **ChatGPT 状态调试**

如果设置参数后，依然没有效果，请尝试使用`./k8m -v 6`获取更多的调试信息。
会输出以下信息，通过查看日志，确认是否启用ChatGPT。

```go
ChatGPT 开启状态:true
ChatGPT 启用 key:sk-hl**********************************************, url:https: // api.siliconflow.cn/v1
ChatGPT 使用环境变量中设置的模型:Qwen/Qwen2.5-7B-Instruc
```

### **ChatGPT 账户**

本项目集成了[github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)SDK。
国内访问推荐使用[硅基流动](https://cloud.siliconflow.cn/)的服务。
登录后，在[https://cloud.siliconflow.cn/account/ak](https://cloud.siliconflow.cn/account/ak)创建API_KEY

## **k8m 支持环境变量设置**

以下是k8m支持的环境变量设置参数及其作用的表格：

| 环境变量                  | 默认值                        | 说明                                                                    |
|-----------------------|----------------------------|-----------------------------------------------------------------------|
| `PORT`                | `3618`                     | 监听的端口号                                                                |
| `MCP_SERVER_PORT`     | `3619`                     | 内置多集群k8s MCP Server监听的端口号                                             |
| `KUBECONFIG`          | `~/.kube/config`           | `kubeconfig` 文件路径  ，会自动扫描识别同级目录下所有的配置文件                               |
| `ENABLE_AI`           | `"true"`                   | 开启AI功能，默认开启                                                           |
| `USE_BUILTIN_MODEL`   | `"true"`                   | 使用内置大模型参数，默认开启                                                        |
| `OPENAI_API_KEY`      | `""`                       | 自定义 大模型的 API Key                                                      |
| `OPENAI_API_URL`      | `""`                       | 自定义 大模型的 API URL                                                      |
| `OPENAI_MODEL`        | `Qwen/Qwen2.5-7B-Instruct` | 自定义 大模型的默认模型名称，如需DeepSeek，请设置为deepseek-ai/DeepSeek-R1-Distill-Qwen-7B |
| `ANY_SELECT`          | `"true"`                   | 是否开启任意选择划词解释，默认开启 (default true)                                      |
| `LOGIN_TYPE`          | `"password"`               | 登录方式（如 `password`, `oauth`, `token`）                                  |
| `ENABLE_TEMP_ADMIN`   | `"false"`                  | 是否启用临时管理员账户配置，默认关闭。初次登录、忘记密码时使用                                       |
| `ADMIN_USERNAME`      | `"admin"`                  | 管理员用户名 ，启用临时管理员账户配置后生效                                                |
| `ADMIN_PASSWORD`      | `"123456"`                 | 管理员密码，启用临时管理员账户配置后生效                                                  |
| `DEBUG`               | `"false"`                  | 是否开启 `debug` 模式                                                       |
| `LOG_V`               | `"2"`                      | log输出日志，同klog用法                                                       |
| `JWT_TOKEN_SECRET`    | `"your-secret-key"`        | 用于 JWT Token 生成的密钥                                                    |
| `KUBECTL_SHELL_IMAGE` | `bitnami/kubectl:latest`   | kubectl shell 镜像地址                                                    |
| `NODE_SHELL_IMAGE`    | `alpine:latest`            | Node shell 镜像地址                                                       |
| `SQLITE_PATH`         | `./data/k8m.db`            | 持久化数据库地址，默认sqlite数据库，文件地址./data/k8m.db                                |
| `CONNECT_CLUSTER`     | `"false"`                  | 启动程序后，是否自动连接发现的集群，默认关闭                                                |
| `IN_CLUSTER`          | `"true"`                   | 是否自动注册纳管宿主集群，默认启用                                                     |
| `PRINT_CONFIG`        | `"false"`                  | 是否打印配置信息                                                              |

这些环境变量可以通过在运行应用程序时设置，例如：

```sh
export PORT=8080
export OPENAI_API_KEY="your-api-key"
export GIN_MODE="release"
./k8m
```

**注意：环境变量会被启动参数覆盖。**

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

### 修改配置

首选建议通过修改环境变量方式进行修改。 例如增加deploy.yaml中的env参数

## 内置MCP Server 使用说明

### 服务端点，可开发供其他AI工具使用

MCP程序使用3619端口。NodePort使用31919端口。
如果二进制方式直接启动，那么访问地址为http://ip:3619/sse
如果集群方式启动，则访问地址为则访问地址为http://nodeIP:31919/sse

### 集群管理范围

内置MCP Server 管理范围与k8m 纳管的集群范围一致。
界面内已连接的集群均可使用。

### 内置MCP Server 配置说明

#### MCP工具列表（49种）

| 类别                 | 方法                             | 描述                                      |
|--------------------|--------------------------------|-----------------------------------------|
| **集群管理（1）**        | `list_clusters`                | 列出所有已注册的Kubernetes集群                    |
| **部署管理（12）**       | `scale_deployment`             | 扩缩容Deployment                           |
|                    | `restart_deployment`           | 重启Deployment                            |
|                    | `stop_deployment`              | 停止Deployment                            |
|                    | `restore_deployment`           | 恢复Deployment                            |
|                    | `update_tag_deployment`        | 更新Deployment镜像标签                        |
|                    | `rollout_history_deployment`   | 查询Deployment升级历史                        |
|                    | `rollout_undo_deployment`      | 回滚Deployment                            |
|                    | `rollout_pause_deployment`     | 暂停Deployment升级                          |
|                    | `rollout_resume_deployment`    | 恢复Deployment升级                          |
|                    | `rollout_status_deployment`    | 查询Deployment升级状态                        |
|                    | `hpa_list_deployment`          | 查询Deployment的HPA列表                      |
|                    | `list_deployment_pods`         | 获取Deployment管理的Pod列表                    |
| **动态资源管理(含CRD，8)** | `get_k8s_resource`             | 获取k8s资源                                 |
|                    | `describe_k8s_resource`        | 描述k8s资源                                 |
|                    | `delete_k8s_resource`          | 删除k8s资源                                 |
|                    | `list_k8s_resource`            | 列表形式获取k8s资源                             |
|                    | `list_k8s_event`               | 列表形式获取k8s事件                             |
|                    | `patch_k8s_resource`           | 更新k8s资源，以JSON Patch方式更新                 |                               |
|                    | `label_k8s_resource`           | 为k8s资源添加或删除标签                           |
|                    | `annotate_k8s_resource`        | 为k8s资源添加或删除注解                           |
| **节点管理（8）**        | `taint_node`                   | 为节点添加污点                                 |
|                    | `untaint_node`                 | 为节点移除污点                                 |
|                    | `cordon_node`                  | 为节点设置Cordon                             |
|                    | `uncordon_node`                | 为节点取消Cordon                             |
|                    | `drain_node`                   | 为节点执行Drain                              |
|                    | `get_node_resource_usage`      | 查询节点的资源使用情况                             |
|                    | `get_node_ip_usage`            | 查询节点上Pod IP资源使用情况                       |
|                    | `get_node_pod_count`           | 查询节点上的Pod数量                             |
| **Pod 管理（14）**     | `list_pod_files`               | 列出Pod文件                                 |
|                    | `list_all_pod_files`           | 列出Pod所有文件                               |
|                    | `delete_pod_file`              | 删除Pod文件                                 |
|                    | `upload_file_to_pod`           | 上传文件到Pod内，支持传递文本内容，存储为Pod内文件            |
|                    | `get_pod_logs`                 | 获取Pod日志                                 |
|                    | `run_command_in_pod`           | 在Pod中执行命令                               |
|                    | `get_pod_linked_service`       | 获取Pod关联的Service                         |
|                    | `get_pod_linked_ingress`       | 获取Pod关联的Ingress                         |
|                    | `get_pod_linked_endpoints`     | 获取Pod关联的Endpoints                       |
|                    | `get_pod_linked_pvc`           | 获取Pod关联的PVC                             |
|                    | `get_pod_linked_pv`            | 获取Pod关联的PV                              |
|                    | `get_pod_linked_env`           | 通过在pod内运行env命令获取Pod运行时环境变量              |
|                    | `get_pod_linked_env_from_yaml` | 通过Pod yaml定义获取Pod运行时环境变量                |
|                    | `get_pod_resource_usage`       | 获取Pod的资源使用情况，包括CPU和内存的请求值、限制值、可分配值和使用比例 |
| **YAML管理（2）**      | `apply_yaml`                   | 应用YAML资源                                |
|                    | `delete_yaml`                  | 删除YAML资源                                |
| **存储管理（3）**        | `set_default_storageclass`     | 设置默认StorageClass                        |
|                    | `get_storageclass_pvc_count`   | 获取StorageClass下的PVC数量                   |
|                    | `get_storageclass_pv_count`    | 获取StorageClass下的PV数量                    |
| **Ingress管理（1）**   | `set_default_ingressclass`     | 设置默认IngressClass                        |

### AI工具集成

#### 通用配置文件

适合MCP工具集成，如Cursor、Claude Desktop、Windsurf等，此外也可以使用这些软件的UI操作界面进行添加。

```json
{
  "mcpServers": {
    "kom": {
      "type": "sse",
      "url": "http://IP:9096/sse"
    }
  }
}
```

#### Cursor

1. 进入Cursor设置界面
2. 找到扩展服务配置选项
3. 添加MCP Server的URL（例如：http://localhost:3619/sse）

#### Windsurf

1. 访问配置中心
2. 设置API服务器地址

### MCP常见问题

1. 确保MCP Server正常运行且端口可访问
2. 检查网络连接是否正常
3. 验证SSE连接是否成功建立
4. 查看工具日志以排查连接问题，MCP执行失败会有报错记录。

### HELP & SUPPORT

如果你有任何进一步的问题或需要额外的帮助，请随时与我联系！

### 特别鸣谢

[zhaomingcheng01](https://github.com/zhaomingcheng01)：提出了诸多非常高质量的建议，为k8m的易用好用做出了卓越贡献~

[La0jin](https://github.com/La0jin):提供在线资源及维护，极大提升了k8m的展示效果

## 联系我

微信（大罗马的太阳） 搜索ID：daluomadetaiyang,备注k8m。

## 微信群

![输入图片说明](https://foruda.gitee.com/images/1744617633379170574/e3a9495e_77493.png "屏幕截图")
