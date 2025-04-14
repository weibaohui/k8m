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

## 更新日志

**v0.0.87更新**
1.  集群授权支持对用户组进行授权
集群授权：
![输入图片说明](https://foruda.gitee.com/images/1744554238488470925/351bbc00_77493.png "屏幕截图")
用户管理视角，看用户有哪些集群权限：
![](https://foruda.gitee.com/images/1744554316927031816/24a3c6ce_77493.png "屏幕截图")
集群管理视角，看某集群下已授权用户：
![输入图片说明](https://foruda.gitee.com/images/1744554384827407363/e3d0136b_77493.png "屏幕截图")
用户视角，看自己有哪些已获得授权的集群列表：
![输入图片说明](https://foruda.gitee.com/images/1744554435367667674/1af1bd5e_77493.png "屏幕截图")


**v0.0.86更新**
1. 资源状态翻转
新增状态指标翻转，将压力、问题等表述的状态，翻转显示为正常
![输入图片说明](https://foruda.gitee.com/images/1744466360319112414/5554605f_77493.png "屏幕截图")
![输入图片说明](https://foruda.gitee.com/images/1744466344472247290/484335b8_77493.png "屏幕截图")
2. 新增MCP工具的独立开关
约束每一个工具，控制大模型可使用tools的范围，屏蔽高危操作，减低大模型交互负担。
![输入图片说明](https://foruda.gitee.com/images/1744466440407504939/108fd6d9_77493.png "屏幕截图")
3. 新增临时管理员账户配置开关
开启后，可通过启动参数、环境变量设置平台管理员用户名密码。增加正常管理员后，可关闭临时管理员。
该功能默认不生效，也就是不设置开启，只能使用数据库用户名密码登录。确保安全。
建议生产环境非必要不要启用。
4. 新增集群自动连接开关
开启后，会自动连接已注册的集群。

**v0.0.75更新**

1. 分离用户操作界面、平台管理界面。平台管理界面新增一个平台管理菜单。
   1.1 用户多集群切换，保留切换、连接功能:
   ![输入图片说明](https://foruda.gitee.com/images/1743904007097350906/c1dd8712_77493.png "屏幕截图")
   1.2 管理员操作多集群，新增断开功能：
   ![输入图片说明](https://foruda.gitee.com/images/1743904225916002465/6ee9a422_77493.png "屏幕截图")
   1.3 集群管理新增已授权页面，展示集群下所有的授权用户
   ![输入图片说明](https://foruda.gitee.com/images/1743904287185877723/dbc711cb_77493.png "屏幕截图")
   1.4 用户管理新增授权页面，查看某用户所有的授权集群
   ![输入图片说明](https://foruda.gitee.com/images/1743904361656769506/de632dca_77493.png "屏幕截图")
2. 新增权限可设置ns，集群授权后，可补充ns，默认为不限制，填写后，将限制用户活动范围。
   ![输入图片说明](https://foruda.gitee.com/images/1743904016110156134/7aa4c81c_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1743904052295697800/8f38845c_77493.png "屏幕截图")
3. 新增参数配置页面。
   启动后会先加载环境变量、env文件、页面配置，依次覆盖。最终页面配置为准。
   ![输入图片说明](https://foruda.gitee.com/images/1743904079152543105/cf923008_77493.png "屏幕截图")
4. 新增资源、副本数调整页面
   ![输入图片说明](https://foruda.gitee.com/images/1743904476260674721/310b0f04_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1743904500139794748/df5c4bed_77493.png "屏幕截图")

**V0.0.73 更新**

1. 新增Deploy探针管理页面
   ![输入图片说明](https://foruda.gitee.com/images/1743686996148531876/a3dc1131_77493.png "屏幕截图")
1. 新增MCP多集群不传值时提示，只有一个集群时可以省去集群名称
   ![输入图片说明](https://foruda.gitee.com/images/1743687014148809259/e5526f1f_77493.png "屏幕截图")
1. 修复未授权用户看到一个默认集群的问题

**V0.0.72 更新**

1. MCP 大模型调用权限上线，一句话概述：谁使用大模型，就用谁的权限执行MCP
   ![输入图片说明](https://foruda.gitee.com/images/1743650492231083539/72855c43_77493.png "屏幕截图")

**V0.0.70 更新**

1. 权限管理调整：按集群进行权限隔离
   ![输入图片说明](https://foruda.gitee.com/images/1743436163730546653/203d33f7_77493.png "屏幕截图")

**v0.0.67 更新**

1. 新增：MCP查询事件工具
   ![输入图片说明](https://foruda.gitee.com/images/1742916865442166281/43b26650_77493.png "屏幕截图")
2. 新增：MCP查询注册集群工具
   ![输入图片说明](https://foruda.gitee.com/images/1742917222171687147/216d03f1_77493.png "屏幕截图")
3. 新增：MCP查询事件工具
   ![输入图片说明](https://foruda.gitee.com/images/1742917268538391635/9e25fbb3_77493.png "屏幕截图")
4. 增强：列表查询资源支持label ，如app=k8m
   ![输入图片说明](https://foruda.gitee.com/images/1742916917319897798/a2171fd2_77493.png "屏幕截图")
5. 增强：MCP服务器增加快捷开启关闭按钮
   ![输入图片说明](https://foruda.gitee.com/images/1742916947056442916/6c33d7c2_77493.png "屏幕截图")

**V0.0.66更新**

1. 新增MCP支持。
2. 内置支持k8s多集群操作：
    1. list_k8s_resource
    2. get_k8s_resource
    3. delete_k8s_resource
    4. describe_k8s_resource
    5. get_pod_logs

**v0.0.64 更新**

1. 增加MCP支持
   ![输入图片说明](https://foruda.gitee.com/images/1742621225108846936/0a614dcb_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742621196785322998/4174b937_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742621204002335466/8a02cd2c_77493.png "屏幕截图")

**v0.0.62 更新**

1. 划词解释增加全屏按钮
   解决部分情况下解释内容非常多，查看不方便，以及滚动条不能完整滚动的问题。
   ![输入图片说明](https://foruda.gitee.com/images/1742085361623662812/c569323a_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742085379102268742/769429f2_77493.png "屏幕截图")

**v0.0.61 更新**

1. 新增2FA两步验证
   启用后，登录时需填写验证码，增强安全性
   ![输入图片说明](https://foruda.gitee.com/images/1742012358386285979/eada8b94_77493.png "屏幕截图")
2. InCluster运行模式增加开关
   默认开启，可设置环境变量显式关闭。按需开启。
3. 优化资源用量显示逻辑
   未设置资源用量，在k8s中属于最低保障等级。界面显示进度条调整为红色100%，提醒管理员关注。
   ![资源用量](https://foruda.gitee.com/images/1742012525046823733/35acfc96_77493.png "屏幕截图")

**v0.0.60更新**

1. 增加helm 常用仓库
   ![输入图片说明](https://foruda.gitee.com/images/1741792802066909841/f20b8736_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741792815487933294/a4b9c193_77493.png "屏幕截图")
2. Namespace增加LimitRange、ResourceQuota快捷菜单
   ![输入图片说明](https://foruda.gitee.com/images/1741792871141287157/f0a51266_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741792891386812848/ad928eb1_77493.png "屏幕截图")
3. 增加InCluster模式开关
   默认开启InCluster模式，如需关闭，可以注入环境变量，或修改配置文件，或修改命令行参数

**v0.0.53更新**

1. 日志查看支持颜色，如果输出console的时候带有颜色，那么在pod 日志查看时就可以显示。
   ![输入图片说明](https://foruda.gitee.com/images/1741180128542917712/d4034cfb_77493.png "屏幕截图")
2. Helm功能上线
   2.1 新增helm仓库
   ![输入图片说明](https://foruda.gitee.com/images/1741180306318265893/f7c561cf_77493.png "屏幕截图")
   2.2 安装helm chart 应用
   应用列表
   ![输入图片说明](https://foruda.gitee.com/images/1741180337250117323/373632c3_77493.png "屏幕截图")
   查看应用
   ![输入图片说明](https://foruda.gitee.com/images/1741180373708023891/01b2eef5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741180423218217871/b1b2b06f_77493.png "屏幕截图")
   支持对参数内容选中划词AI解释
   ![输入图片说明](https://foruda.gitee.com/images/1741180604109610379/b26ae294_77493.png "屏幕截图")
   2.3 查看已部署release
   ![输入图片说明](https://foruda.gitee.com/images/1741180730249955448/bd51776e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741180757526613636/3cff8334_77493.png "屏幕截图")
   2.4 查看安装参数
   ![输入图片说明](https://foruda.gitee.com/images/1741180785289693466/dd1e08ab_77493.png "屏幕截图")
   2.5 更新、升级、降级部署版本
   ![输入图片说明](https://foruda.gitee.com/images/1741180817303995346/b2bb7472_77493.png "屏幕截图")
   2.6 查看已部署release变更历史
   ![输入图片说明](https://foruda.gitee.com/images/1741180840762812700/ccd3aa07_77493.png "屏幕截图")

**v0.0.50更新**

1. 新增HPA
   ![输入图片说明](https://foruda.gitee.com/images/1740664600490309267/48ff3895_77493.png "屏幕截图")
2. 关联资源增加HPA
   ![输入图片说明](https://foruda.gitee.com/images/1740664626159889748/96a40af4_77493.png "屏幕截图")

**v0.0.49更新**

1. 新增标签搜索：支持精确搜索、模糊搜索。
   精确搜索。可以搜索k，k=v两种方式精确搜索。默认列出所有标签。支持自定义新增搜索标签。
   ![输入图片说明](https://foruda.gitee.com/images/1740664804869894211/257140ad_77493.png "屏幕截图")
   模糊搜索。可以搜索k，v中的任意满足。类似like %xx%的搜索方式。
   ![输入图片说明](https://foruda.gitee.com/images/1740664820221541385/cf840a61_77493.png "屏幕截图")
2. 多集群纳管支持自定义名称。
   ![输入图片说明](https://foruda.gitee.com/images/1740664838997975455/95aeec37_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740664855863544600/3496c16f_77493.png "屏幕截图")
3. 优化Pod状态显示
   在列表页展示pod状态，不同颜色区分正常运行与未就绪运行。
   ![输入图片说明](https://foruda.gitee.com/images/1740664869098640512/0d4002eb_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740664883842793338/17f94df3_77493.png "屏幕截图")

**v0.0.44更新**

1. 新增kubectl shell 功能
   可以web 页面执行 kubectl 命令了
   ![输入图片说明](https://foruda.gitee.com/images/1740031049224924895/c8d5357b_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031092919251676/61e6246c_77493.png "屏幕截图")

2. 新增节点终端NodeShell
   在节点上执行命令
   ![输入图片说明](https://foruda.gitee.com/images/1740031147702527911/4cef40dc_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031249763550505/69fddee6_77493.png "屏幕截图")
3. 新增创建功能页面
   执行过的yaml会保存下来，下次打开页面可以直接点击，收藏的yaml可以导入导出。导出的文件为yaml，可以复用
   ![输入图片说明](https://foruda.gitee.com/images/1740031367996726581/e1a357b7_77493.png "屏幕截图")
   ![](https://foruda.gitee.com/images/1740031382494497806/d16b1a79_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031533749791121/4e64e286_77493.png "屏幕截图")
4. deploy、ds、sts等类型新增关联资源
   4.1 容器组
   直接显示其下受控的pod容器组，并提供快捷操作
   ![输入图片说明](https://foruda.gitee.com/images/1740031610441749272/cd485e87_77493.png "屏幕截图")
   4.2 关联事件
   显示deploy、rs、pod等所有相关的事件，一个页面看全相关事件
   ![deploy](https://foruda.gitee.com/images/1740031712446573977/320c920b_77493.png "屏幕截图")
   4.3 日志
   显示Pod列表，可选择某个pod、Container展示日志
   ![](https://foruda.gitee.com/images/1740031809856930240/fbbef393_77493.png "屏幕截图")
   4.4 历史版本
   支持历史版本查看，并可diff
   ![输入图片说明](https://foruda.gitee.com/images/1740031862075460381/ebf50a7e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031912370086873/dfa95a2f_77493.png "屏幕截图")

5. 全新AI对话窗口
   ![输入图片说明](https://foruda.gitee.com/images/1740062818194113045/6ae3af0b_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740062840392675452/a429aab8_77493.png "屏幕截图")

6. 全新AI搜索方式，哪里不懂选哪里
   页面所有地方都可以`划词翻译`,哪里有疑问就选中哪里。
   ![输入图片说明](https://foruda.gitee.com/images/1740062958174067230/7c377b16_77493.png "屏幕截图")

**v0.0.21更新**

1. 新增问AI功能：
   有什么问题，都可以直接询问AI，让AI解答你的疑惑
   ![输入图片说明](https://foruda.gitee.com/images/1736655942078335649/be66c2b5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736655968296155521/d47d247e_77493.png "屏幕截图")
2. 文档界面优化：
   优化AI翻译效果，降低等待时间
   ![AI文档](https://foruda.gitee.com/images/1736656055530922469/df155262_77493.png "屏幕截图")
3. 文档字段级AI示例：
   针对具体的字段，给出解释，给出使用Demo样例。
   ![输入图片说明](https://foruda.gitee.com/images/1736656231132357556/b41109e6_77493.png "屏幕截图")
4. 增加容忍度详情：
   ![输入图片说明](https://foruda.gitee.com/images/1736656289098443083/ce1f5615_77493.png "屏幕截图")
5. 增加Pod关联资源
   一个页面，展示相关的svc、endpoint、pvc、env、cm、secret，甚至集成了pod内的env列表，方便查看
   ![输入图片说明](https://foruda.gitee.com/images/1736656365325777082/410d24c5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656376791203135/64cc4737_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656390371435096/5d93c74a_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656418411787086/2c8510af_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656445050779433/843f56aa_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656457940557219/c1372abd_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656468351816442/aba6f649_77493.png "屏幕截图")
6. yaml创建增加导入功能：
   增加导入功能，可以直接执行，也可导入到编辑器。导入编辑器后可以二次编辑后，再执行。
   ![输入图片说明](https://foruda.gitee.com/images/1736656627742328659/6c4e745e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656647758880134/ca92dcc2_77493.png "屏幕截图")

**v0.0.19更新**

1. 多集群管理功能
   按需选择多集群，可随时切换集群
   ![输入图片说明](https://foruda.gitee.com/images/1736037285365941737/543965e6_77493.png "屏幕截图")
2. 节点资源用量功能
   直观显示已分配资源情况，包括cpu、内存、pod数量、IP数量。
   ![输入图片说明](https://foruda.gitee.com/images/1736037259029155963/72ea1ab4_77493.png "屏幕截图")
3. Pod 资源用量
   ![输入图片说明](https://foruda.gitee.com/images/1736037328973160586/9d322e6d_77493.png "屏幕截图")
4. Pod CPU内存设置
   按范围方式显示CPU设置，内存设置，简洁明了
   ![内存](https://foruda.gitee.com/images/1736037370125604986/7938a1f6_77493.png "屏幕截图")
5. AI页面功能升级为打字机效果
   响应速度大大提升，实时输出AI返回内容，体验升级
   ![输入图片说明](https://foruda.gitee.com/images/1736037522633946187/71955026_77493.png "屏幕截图")

**v0.0.15更新**

1. 所有页面增加资源使用指南。启用AI信息聚合。包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写带有中文注释的yaml示例）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等信息，帮助用户理解k8s
   ![输入图片说明](https://foruda.gitee.com/images/1735400167081694530/e45b55ef_77493.png "屏幕截图")
2. 所有资源页面增加搜索功能。部分页面增高频过滤字段搜索。
   ![输入图片说明](https://foruda.gitee.com/images/1735399974060039020/11bce030_77493.png "屏幕截图")
3. 改进LimitRange信息展示模式
   ![LimitRange](https://foruda.gitee.com/images/1735399148267940416/b4faafbd_77493.png "屏幕截图")
4. 改进状态显示样式
   ![Deployment](https://foruda.gitee.com/images/1735399222088964660/131eda03_77493.png "屏幕截图")
5. 统一操作菜单
   ![操作菜单](https://foruda.gitee.com/images/1735399278081665887/b01c506c_77493.png "屏幕截图")
6. Ingress页面增加域名转发规则信息
   ![输入图片说明](https://foruda.gitee.com/images/1735399689648549556/3d4f8d78_77493.png "屏幕截图")
7. 改进标签显示样式，鼠标悬停展示
   ![输入图片说明](https://foruda.gitee.com/images/1735399387990917764/d06822cb_77493.png "屏幕截图")
8. 优化资源状态样式更小更紧致
   ![输入图片说明](https://foruda.gitee.com/images/1735399419170194492/268b25c8_77493.png "屏幕截图")
9. 丰富Service展示信息
   ![输入图片说明](https://foruda.gitee.com/images/1735399493417833664/fa968343_77493.png "屏幕截图")
10. 突出显示未就绪endpoints
    ![输入图片说明](https://foruda.gitee.com/images/1735399531801079962/9a13cd50_77493.png "屏幕截图")
11. endpoints鼠标悬停展开未就绪IP列表
    ![输入图片说明](https://foruda.gitee.com/images/1735399560648695064/8079b5cf_77493.png "屏幕截图")
12. endpointslice 突出显示未ready的IP及其对应的POD，
    ![输入图片说明](https://foruda.gitee.com/images/1735399614582278222/c1f40aa0_77493.png "屏幕截图")
13. 角色增加延展信息
    ![输入图片说明](https://foruda.gitee.com/images/1735399896080683883/3e9a7359_77493.png "屏幕截图")
14. 角色与主体对应关系
    ![输入图片说明](https://foruda.gitee.com/images/1735399923738735980/c5730152_77493.png "屏幕截图")
15. 界面全量中文化，k8s资源翻译为中文，方便广大用户使用。
    ![输入图片说明](https://foruda.gitee.com/images/1735400283406692980/c778158c_77493.png "屏幕截图")
    ![输入图片说明](https://foruda.gitee.com/images/1735400313832429462/279018dc_77493.png "屏幕截图")

### HELP & SUPPORT

如果你有任何进一步的问题或需要额外的帮助，请随时与我联系！

### 特别鸣谢

[zhaomingcheng01](https://github.com/zhaomingcheng01)：提出了诸多非常高质量的建议，为k8m的易用好用做出了卓越贡献~

[La0jin](https://github.com/La0jin):提供在线资源及维护，极大提升了k8m的展示效果

## 联系我

微信（大罗马的太阳） 搜索ID：daluomadetaiyang,备注k8m。

## 微信群

![输入图片说明1](https://foruda.gitee.com/images/1743782774634886044/96829f36_77493.png "屏幕截图")