# CODEBUDDY.md

本文件为 CodeBuddy Code 在此代码仓库中工作时提供指导。

## 项目概述

K8M 是一款 AI 驱动的 Mini Kubernetes AI Dashboard 轻量级控制台工具，采用 Golang 后端和 AMIS 前端构建。它使用 `kom` 库作为 Kubernetes API 客户端，并拥有完整的插件系统以实现功能扩展。

## 开发命令

### 构建

```bash
# 为当前平台构建
make build

# 为所有平台构建（Linux、macOS、Windows、多种架构）
make build-all

# 仅构建 Linux 平台
make build-linux

# 构建 Docker 镜像
make docker
```

### 运行

```bash
# 使用 air 运行（热重载，推荐用于开发）
air

# 运行构建的二进制文件
./bin/k8m

# 以调试模式运行
./k8m -d -v 6 --print-config --enable-temp-admin
```

### 前端开发

```bash
# 构建前端（首次运行前必需）
cd ui
pnpm run build

# 开发模式（热重载）
cd ui
pnpm run dev
# 前端运行在 localhost:3000，代理后端到 3618
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/comm/utils

# 以详细输出运行测试
go test -v ./...

# 运行特定测试
go test -v ./pkg/comm/utils -run TestRandIntOverflow
```

### 代码检查

```bash
# 检查前端代码
cd ui
pnpm run lint
```

## 架构

### 项目结构

```
k8m/
├── main.go                 # 应用程序入口点、路由设置、初始化
├── pkg/
│   ├── controller/         # HTTP 处理器，按资源类型组织
│   │   ├── admin/         # 仅管理员端点（集群、配置、用户、菜单）
│   │   ├── pod/           # Pod 相关端点（日志、执行、文件、指标）
│   │   ├── node/          # Node 相关端点
│   │   ├── deploy/        # Deployment 端点
│   │   └── ...            # 其他资源控制器
│   ├── plugins/           # 插件系统核心
│   │   ├── modules/       # 插件实现
│   │   ├── api/           # 插件能力抽象接口
│   │   └── readme.md      # 完整的插件架构文档
│   ├── service/           # 业务逻辑层
│   │   ├── clusters.go     # 集群管理（多集群支持）
│   │   └── ...            # 其他服务模块
│   ├── models/            # GORM 数据库模型
│   ├── middleware/        # HTTP 中间件（认证、集群选择）
│   └── response/          # 响应处理工具
├── internal/dao/          # 数据库访问层
├── ui/                    # 前端（React + AMIS）
│   └── package.json
└── docs/                  # 文档
```

### 关键架构模式

#### 1. 插件系统

整个应用围绕插件架构构建。所有功能都以插件形式实现，具有完整的生命周期管理：

- **插件类型**：模块在编译时通过 Go 注册发现
- **生命周期**：安装 → 启用 → 启动 → 运行中 → 停止 → 禁用 → 卸载
- **路由类型**：
  - `ClusterRouter`: `/k8s/cluster/{cluster}/plugins/{name}/*` - 集群特定操作
  - `ManagementRouter`: `/mgm/plugins/{name}/*` - 通用管理操作
  - `PluginAdminRouter`: `/admin/plugins/{name}/*` - 平台管理员操作
  - `RootRouter`: 根级路由（很少使用）
- **依赖关系**：插件声明强依赖（`Dependencies`）和启动顺序（`RunAfter`）
- **前端**：所有插件使用 AMIS JSON 构建 UI（插件中不使用 React/Vue 代码）
- **数据库**：每个插件使用 GORM AutoMigrate 管理自己的表

**插件目录结构**：
```
pkg/plugins/modules/{plugin-name}/
├── metadata.go          # 插件元信息和能力声明
├── lifecycle.go         # 生命周期接口实现
├── models/              # 数据库模型
│   ├── db.go           # Init/Upgrade/DropDB 函数
│   └── *.go            # 模型定义
├── route/               # 路由注册
│   ├── cluster_api.go  # 集群操作
│   ├── mgm_api.go      # 管理操作
│   └── admin_api.go    # 管理员操作
├── frontend/            # AMIS JSON 文件
├── controller/          # HTTP 处理器（可选）
└── service/             # 业务逻辑（可选）
```

**关键**：所有插件描述必须使用 Go 代码（不使用 YAML/JSON）。所有生命周期方法必须是幂等的。

#### 2. 多集群管理

- **集群服务** (`pkg/service/clusters.go`)：管理多个 Kubernetes 集群
- **注册方式**：
  - InCluster 模式：自动注册宿主集群
  - Kubeconfig 扫描：扫描目录中的配置文件
  - 数据库：存储已注册的集群
- **上下文**：集群 ID 通过 `EnsureSelectedClusterMiddleware` 注入到请求中

#### 3. 分层架构

```
HTTP 请求
    ↓
中间件（认证、集群选择）
    ↓
控制器（pkg/controller/*）  # 路由处理器
    ↓
服务层（pkg/service/*）       # 业务逻辑
    ↓
Kom（Kubernetes 客户端）       # K8s API 调用
    ↓
数据库（GORM）                # 数据持久化
```

#### 4. 响应处理

所有 HTTP 处理器使用 `pkg/response.Context`，提供：
- `JSON(status, obj)` - JSON 响应
- `ShouldBindJSON(obj)` - JSON 请求绑定
- `Param(key)` - URL 参数
- `Query(key)` - 查询参数
- 响应上下文池以提高性能

#### 5. 数据库抽象

- **ORM**：支持 SQLite、MySQL、PostgreSQL 的 GORM
- **迁移**：每个插件通过 `AutoMigrate` 处理自己的模式
- **访问**：`internal/dao` 包提供 `dao.DB()` 单例
- **配置**：参见 `docs/database.md` 了解数据库设置选项

### 前端架构

- **框架**：React 18 + AMIS（百度的低代码 UI 框架）
- **构建**：Vite + TypeScript
- **UI 库**：AMIS JSON 模式（不是 React 组件）
- **嵌入**：前端通过 `//go:embed` 嵌入到 Go 二进制文件中
- **开发**：Vite 开发服务器运行在 3000 端口，代理 API 到 3618

#### Fetcher 封装

前端所有 HTTP 请求都通过 `ui/src/components/Amis/fetcher.ts` 封装。该 fetcher 是 AMIS 框架的标准接口。

**核心功能**：

1. **自动认证**：从 localStorage 获取 token，自动添加 `Authorization: Bearer ${token}` 请求头

2. **集群上下文注入**：
   - 自动处理集群 URL 重写
   - 支持通过 `x-k8m-target-cluster` 请求头或 `__cluster` 参数指定目标集群
   - 调用 `ProcessK8sUrlWithCluster()` 函数处理 K8s 集群相关 URL

3. **错误处理**：
   - 401：自动跳转到登录页面（`/#/login`）
   - 512：提示集群权限不足，跳转到集群用户管理页面
   - 403：提示权限不足

4. **请求方法**：
   - GET：data 对象自动转换为查询参数
   - POST/PUT/DELETE：data 作为请求体

5. **响应验证**：
   - 检查空响应或无效响应
   - 在开发模式下打印响应日志
   - 对 204、205、HEAD 请求跳过空数据检查

**使用方式**：

fetcher 在 AMIS 渲染时自动配置（见 `ui/src/components/Amis/index.tsx:148`）：

```typescript
renderAmis(schema, initialData, {
    theme: 'cxd',
    fetcher,  // 注入自定义 fetcher
    isCancel: value => axios.isCancel(value),
})
```

**在 AMIS Schema 中使用**：

AMIS 组件的 `api` 字段会自动使用这个 fetcher，例如：

```json
{
  "type": "crud",
  "api": {
    "method": "get",
    "url": "/k8s/cluster/${cluster}/pods"
  },
  ...
}
```

fetcher 会自动：
- 添加认证 token
- 处理集群 URL 重写
- 处理错误响应
- 验证响应数据

## 重要规则（来自 .trae/rules/project_rules.md）

1. **热重载**：使用 `air` 进行代码热重载。您可以编译验证代码，但不要执行 `k8m` 命令。
2. **登录凭据**：测试时默认用户名/密码为 `k8m`/`k8m`。
3. **日志**：使用中文打印日志消息。使用 `klog.V(6).Infof` 打印日志。
4. **注释**：可以添加或修改现有注释，但不要删除注释。如果需要删除，请标记为 `[注释待删除]`（comment pending deletion）。

## 关键配置

### 环境变量

所有配置选项请参见 `.env.example`：
- `PORT`：服务器端口（默认：3618）
- `KUBECONFIG`：kubeconfig 文件路径
- `OPENAI_API_KEY`、`OPENAI_API_URL`、`OPENAI_MODEL`：AI 模型配置
- `IN_CLUSTER`：启用 InCluster 模式
- `DB_DRIVER`：数据库类型（sqlite/mysql/postgresql）
- `DEBUG`：调试模式
- `LOG_V`：klog 详细程度级别（默认：2）

### 数据库配置

支持 SQLite（开发默认）、MySQL、PostgreSQL。详细配置请参见 `docs/database.md`。

### 启动流程

1. 从标志/环境变量/数据库加载配置
2. 使用内置模型初始化 AI 服务
3. 注册 kom 回调
4. 注册 InCluster 集群（如果启用）
5. 从数据库和 kubeconfig 目录扫描并注册集群
6. 连接集群（如果启用了 `--connect-cluster`）
7. 启动插件管理器（根据配置加载和启用插件）
8. 构建并服务 HTTP 路由器

## 插件开发

### 创建新插件

1. 创建目录：`pkg/plugins/modules/{plugin-name}/`
2. 在 `metadata.go` 中定义元数据，遵循 demo 插件结构
3. 在 `lifecycle.go` 中实现生命周期接口
4. 在 `models/db.go` 中添加模型，包含 `InitDB()`、`UpgradeDB()`、`DropDB()`
5. 在 `route/*.go` 中注册路由
6. 在 `frontend/*.json` 中添加前端 AMIS JSON
7. 在 `pkg/plugins/modules/registrar/` 中注册插件

**关键参考**：完整的插件架构文档请参见 `pkg/plugins/readme.md`（857 行）。

### 插件生命周期方法

所有方法必须是幂等的：
- `Install(ctx)`：创建表、初始化数据
- `Upgrade(ctx)`：版本变更时迁移模式/数据
- `Enable(ctx)`：注册路由、暴露菜单
- `Disable(ctx)`：隐藏菜单、取消注册路由
- `Start(ctx)`：启动后台协程、订阅 EventBus
- `Stop(ctx)`：停止协程、清理资源
- `StartCron(ctx, spec)`：执行计划任务
- `Uninstall(ctx)`：删除表（或根据请求保留数据）

### EventBus

插件可以订阅系统事件：
- `EventLeaderElected`：当此实例成为 leader 时
- `EventLeaderLost`：当此实例失去 leader 状态时

```go
elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)
```

## API 路由结构

完整的路由图请参见 `docs/route_structure.md`：

- `/auth/*`：认证（登录、SSO）
- `/k8s/cluster/{cluster}/*`：集群特定操作
  - `/plugins/{name}/*`：插件集群操作
- `/mgm/*`：管理操作
  - `/plugins/{name}/*`：插件管理操作
- `/admin/*`：仅平台管理员
  - `/plugins/{name}/*`：插件管理员操作
- `/swagger/*`：API 文档（如果启用了 swagger 插件）

## 测试

- 测试文件位于：`pkg/comm/utils/*_test.go`、`pkg/plugins/modules/webhook/core/*_test.go`
- 使用 `go test` 运行测试
- 测试应遵循 Go 惯例
- 使用表驱动测试进行多用例测试

## 常见开发任务

### 添加新的 API 端点

1. 在适当的 `pkg/controller/*` 包中添加控制器函数
2. 在 `main.go` buildRouter 函数中注册路由
3. 使用 `response.Adapter` 包装处理器
4. 使用 `c.JSON()` 返回响应
5. 如果需要，添加文档（Swagger）

**前端调用**：
- 在 AMIS JSON 中使用 `api` 字段定义请求
- fetcher 会自动处理认证、集群上下文、错误处理
- GET 请求的查询参数通过 `sendOn` 配置发送

示例：
```json
{
  "type": "crud",
  "api": {
    "method": "get",
    "url": "/k8s/cluster/${cluster}/pods",
    "data": {
      "namespace": "${selectedNs}"
    },
    "sendOn": "init"
  }
}
```

### 添加新的 Kubernetes 资源类型

1. 在 `pkg/controller/{resource}/` 中创建控制器包
2. 使用 `kom` 客户端实现 CRUD 操作
3. 在 `main.go` 中的 `/k8s/cluster/{cluster}/` 下注册路由
4. 为 UI 添加前端 AMIS JSON

### 调试

- 使用 `-d` 标志启用调试模式
- 使用 `-v 6` 获取详细日志
- 使用 `--print-config` 查看启动时的配置
- 使用 `klog.V(6).Infof()` 打印调试日志（仅中文）

## 插件能力 API

`pkg/plugins/api` 包提供抽象的插件能力：
- `AIChatService()`：AI 聊天功能
- `WebhookService()`：Webhook 发送
- 其他能力可以按模式添加

这使用 No-Op 模式：调用未注册的能力返回安全的无操作实现。

## 重要提示

- 运行前必须构建前端：`cd ui && pnpm run build`
- 嵌入式前端从 `ui/dist/` 加载
- Air 配置在 `.air.toml` 中
- Air 监视 `.go` 文件，排除 `bin/`、`vendor/`、`ui/`、`images/`
- 生产构建使用 `make build`
- 所有插件操作必须使用带 `dao.DB()` 的 GORM
- 集群上下文通过中间件在集群路由中可用
- 所有面向用户的消息和日志使用中文
- 永远不要删除现有注释而不标记它们
