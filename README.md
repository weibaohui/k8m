## **k8m**

[English](README_en.md) | [中文](README.md)

[![k8m](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/k8m/blob/master/LICENSE)

**k8m** 是一款AI驱动的 Mini Kubernetes AI Dashboard 轻量级控制台工具，专为简化集群管理设计。它基于 AMIS 构建，并通过  [
`kom`](https://github.com/weibaohui/kom)  作为 Kubernetes API 客户端，**k8m** 内置了
Qwen2.5-Coder-7B，支持deepseek-ai/DeepSeek-R1-Distill-Qwen-7B模型
模型交互能力，同时支持接入您自己的私有化大模型。

### 主要特点

- **迷你化设计**：所有功能整合在一个单一的可执行文件中，部署便捷，使用简单。
- **简便易用**：友好的用户界面和直观的操作流程，让 Kubernetes 管理更加轻松。
- **高效性能**：后端采用 Golang 构建，前端基于百度 AMIS，保证资源利用率高、响应速度快。
- **AI驱动融合**：基于ChatGPT实现划词解释、资源指南、YAML属性自动翻译、Describe信息解读、日志AI问诊、运行命令推荐,并集成了[
  `k8s-gpt`](https://github.com/k8sgpt-ai/k8sgpt)功能，实现中文展现，为管理k8s提供智能化支持。
- **多集群管理**：自动识别集群内部使用InCluster模式，配置kubeconfig路径后自动扫描同级目录下的配置文件，同时注册管理多个集群。
- **Pod 文件管理**：支持 Pod 内文件的浏览、编辑、上传、下载、删除，简化日常操作。
- **Pod 运行管理**：支持实时查看 Pod 日志，下载日志，并在 Pod 内直接执行 Shell 命令。
- **CRD 管理**：可自动发现并管理 CRD 资源，提高工作效率。
- **Helm 市场**：支持Helm自由添加仓库，一键安装、卸载、升级 Helm 应用。
- **跨平台支持**：兼容 Linux、macOS 和 Windows，并支持 x86、ARM 等多种架构，确保多平台无缝运行。
- **完全开源**：开放所有源码，无任何限制，可自由定制和扩展，可商业使用。

**k8m** 的设计理念是“AI驱动，轻便高效，化繁为简”，它帮助开发者和运维人员快速上手，轻松管理 Kubernetes 集群。

## **运行**

1. **下载**：从 [GitHub](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。
3. **参数**：

```shell
Usage of ./k8m:
      --admin-password string            管理员密码 (default "123456")
      --admin-username string            管理员用户名 (default "admin")
  -k, --chatgpt-key string               大模型的自定义API Key (default "sk-xxxxxxx")
  -m, --chatgpt-model string             大模型的自定义模型名称 (default "Qwen/Qwen2.5-Coder-7B-Instruct")
  -u, --chatgpt-url string               大模型的自定义API URL (default "https://api.siliconflow.cn/v1")
  -d, --debug                            调试模式
      --in-cluster                       是否自动注册纳管宿主集群，默认启用
      --jwt-token-secret string          登录后生成JWT token 使用的Secret (default "your-secret-key")
  -c, --kubeconfig string                kubeconfig文件路径 (default "/root/.kube/config")
      --kubectl-shell-image string       Kubectl Shell 镜像。默认为 bitnami/kubectl:latest，必须包含kubectl命令 (default "bitnami/kubectl:latest")
      --log-v int                        klog的日志级别klog.V(2) (default 2)
      --login-type string                登录方式，password, oauth, token等,default is password (default "password")
      --node-shell-image string          NodeShell 镜像。 默认为 alpine:latest，必须包含`nsenter`命令 (default "alpine:latest")
  -p, --port int                         监听端口 (default 3618)
      --sqlite-path string               sqlite数据库文件路径， (default "./data/k8m.db")
  -v, --v Level                          klog的日志级别 (default 2)
```

## **ChatGPT 配置指南**

### 内置GPT

从v0.0.8版本开始，将内置GPT，无需配置。
如果您需要使用自己的GPT，请参考以下步骤。

### **环境变量配置**

需要设置环境变量，以启用ChatGPT。

```bash
export OPENAI_API_KEY="sk-XXXXX"
export OPENAI_API_URL="https://api.siliconflow.cn/v1"
export OPENAI_MODEL="Qwen/Qwen2.5-Coder-7B-Instruct"
```

### **ChatGPT 状态调试**

如果设置参数后，依然没有效果，请尝试使用`./k8m -v 6`获取更多的调试信息。
会输出以下信息，通过查看日志，确认是否启用ChatGPT。

```go
ChatGPT 开启状态:true
ChatGPT 启用 key:sk-hl**********************************************, url:https: // api.siliconflow.cn/v1
ChatGPT 使用环境变量中设置的模型:Qwen/Qwen2.5-Coder-7B-Instruc
```

### **ChatGPT 账户**

本项目集成了[github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)SDK。
国内访问推荐使用[硅基流动](https://cloud.siliconflow.cn/)的服务。
登录后，在[https://cloud.siliconflow.cn/account/ak](https://cloud.siliconflow.cn/account/ak)创建API_KEY

## **k8m 支持环境变量设置**

以下是k8m支持的环境变量设置参数及其作用的表格：

| 环境变量                  | 默认值                              | 说明                                                                |
|-----------------------|----------------------------------|-------------------------------------------------------------------|
| `PORT`                | `3618`                           | 监听的端口号                                                            |
| `KUBECONFIG`          | `~/.kube/config`                 | `kubeconfig` 文件路径                                                 |
| `OPENAI_API_KEY`      | `""`                             | 大模型的 API Key                                                      |
| `OPENAI_API_URL`      | `""`                             | 大模型的 API URL                                                      |
| `OPENAI_MODEL`        | `Qwen/Qwen2.5-Coder-7B-Instruct` | 大模型的默认模型名称，如需DeepSeek，请设置为deepseek-ai/DeepSeek-R1-Distill-Qwen-7B |
| `LOGIN_TYPE`          | `"password"`                     | 登录方式（如 `password`, `oauth`, `token`）                              |
| `ADMIN_USERNAME`      | `"admin"`                        | 管理员用户名                                                            |
| `ADMIN_PASSWORD`      | `"123456"`                       | 管理员密码                                                             |
| `DEBUG`               | `"false"`                        | 是否开启 `debug` 模式                                                   |
| `LOG_V`               | `"2"`                            | log输出日志，同klog用法                                                   |
| `JWT_TOKEN_SECRET`    | `"your-secret-key"`              | 用于 JWT Token 生成的密钥                                                |
| `KUBECTL_SHELL_IMAGE` | `bitnami/kubectl:latest`         | kubectl shell 镜像地址                                                |
| `NODE_SHELL_IMAGE`    | `alpine:latest`                  | Node shell 镜像地址                                                   |
| `SQLITE_PATH`         | `/data/k8m.db`                   | 持久化数据库地址，默认sqlite数据库，文件地址/data/k8m.db                             |
| `IN_CLUSTER`          | `"true"`                         | 是否自动注册纳管宿主集群，默认启用                                                 |

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

# 将k8m部署到集群中体验

## 安装脚本

```docker
kubectl apply -f https://raw.githubusercontent.com/weibaohui/k8m/refs/heads/main/deploy/k8m.yaml
```

* 访问：
  默认使用了nodePort开放，请访问31999端口。或自行配置Ingress
  http://NodePortIP:31999

## 修改配置

首选建议通过修改环境变量方式进行修改。 例如增加deploy.yaml中的env参数

### **跨平台编译支持**

**build-all** 目标支持以下操作系统和架构组合的交叉编译：

- **Linux**:
    - `amd64`
    - `arm64`
    - `ppc64le`
    - `s390x`
    - `mips64le`
    - `riscv64`
- **Darwin（macOS）**:
    - `amd64`
    - `arm64`
- **Windows**:
    - `amd64`
    - `arm64`

### **使用示例**

#### **1. 为当前平台构建**

构建适用于当前操作系统和架构的 `k8m` 可执行文件：

```bash
make build
```

#### **2. 为所有支持的平台构建**

交叉编译 `k8m` 为所有指定的平台和架构：

```bash
make build-all
```

#### **3. 运行可执行文件**

在 Unix 系统上构建并运行 `k8m`：

```bash
make run
```

#### **4. 清理构建产物**

删除所有编译生成的可执行文件和 `bin/` 目录：

```bash
make clean
```

#### **5. 查看帮助信息**

显示所有可用的 Makefile 目标及其描述：

```bash
make help
```

### **附加说明**

- **版本控制**：你可以在构建时通过传递 `VERSION` 变量来指定自定义版本：
  ```bash
  make build VERSION=v2.0.0
  ```
- **可执行文件扩展名**：对于 Windows 构建，Makefile 会自动为可执行文件添加 `.exe` 扩展名。
- **依赖性**：确保 Git 已安装并且项目已初始化为 Git 仓库，以便正确获取 `GIT_COMMIT` 哈希值。

### **故障排除**

- **缺少依赖**：如果遇到与缺少命令相关的错误（如 `make`、`go` 等），请确保所有先决条件已安装并正确配置在系统的 `PATH` 中。
- **权限问题**：如果在运行 `make run` 时收到权限被拒绝的错误，请确保 `bin/` 目录和编译后的二进制文件具有必要的执行权限：
  ```bash
  chmod +x bin/k8m
  ```
- **文件浏览权限问题**：依赖容器内的ls命令，请在容器内安装shell、tar、cat等命令 。
- **无法启动**：启动时卡住，请使用 k8m -v 6
  命令启动，会输出更多日志，一般是由于部分版本的k8s集群的openAPI文档格式问题导致，请将日志贴到issue，或微信发我，我将优先处理 。

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

## 联系我

微信（大罗马的太阳） 搜索ID：daluomadetaiyang,备注k8m。

## 微信群

![img](images/wechat.jpg)
