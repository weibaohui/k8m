

# k8m 插件体系 —— Context 能力模型与生命周期绑定规范

## 1. 设计目标

为支撑 k8m 插件体系中 **复杂功能模块（前后端、权限、SQL、任务、事件）** 的可插拔、可升级与可治理能力，系统必须在**生命周期层面**对插件能力进行严格约束。

本规范通过 **Context 拆分** 的方式，将插件在不同生命周期阶段可使用的能力进行强隔离，防止：

* 插件越权操作
* 生命周期语义混乱
* 后期安全、审计、升级不可控

**Context 即能力合同（Capability Contract）**。

---

## 2. Context 总体分层模型

```text
                +------------------+
                |   BaseContext    |
                | (通用只读能力)   |
                +------------------+
                         ▲
        ┌───────────────┼────────────────┐
        │               │                │
InstallContext   UpgradeContext   EnableContext
        │                                │
        └────────────── RunContext ──────┘
```

---

## 3. BaseContext（所有生命周期共享）

### 3.1 职责定位

* 提供插件运行所需的**基础只读信息**
* 不允许产生任何副作用
* 所有其他 Context **必须继承 BaseContext**

### 3.2 能力范围

* 插件元信息（名称、版本）
* 日志能力（自动带插件名、生命周期）
* 插件私有配置读取

### 3.3 能力约束

* ❌ 不允许访问数据库
* ❌ 不允许注册菜单、权限、API
* ❌ 不允许启动任务

### 3.4 抽象定义（规范）

```go
type BaseContext interface {
    Meta() PluginMeta
    Logger() Logger
    Config() PluginConfig
}
```

---

## 4. InstallContext（安装期）

### 4.1 生命周期语义

* 仅在插件**首次安装**时执行
* 用于完成插件运行前的**环境准备**
* **只执行一次，不可重复**

### 4.2 允许能力

* 创建 SQL 表
* 初始化基础数据
* 注册插件级配置项

### 4.3 明确禁止

* ❌ 注册菜单
* ❌ 注册 AMIS 页面
* ❌ 注册 API
* ❌ 启动后台任务

### 4.4 抽象定义（规范）

```go
type InstallContext interface {
    BaseContext

    DB() SchemaOperator
    ConfigRegistry() ConfigRegistry
}
```

---

## 5. UpgradeContext（升级期）

### 5.1 生命周期语义

* 当插件版本发生变化时触发
* 负责 **从旧版本状态安全迁移到新版本**
* Upgrade 成功之前，插件不得 Enable

### 5.2 允许能力

* SQL Schema 变更
* 数据迁移
* 权限模型结构调整

### 5.3 明确禁止（强约束）

* ❌ 注册菜单
* ❌ 注册 AMIS 页面
* ❌ 注册 API
* ❌ 启动后台任务
* ❌ 执行业务逻辑

### 5.4 抽象定义（规范）

```go
type UpgradeContext interface {
    BaseContext

    FromVersion() string
    ToVersion() string

    DB() MigrationOperator
}
```

---

## 6. EnableContext（启用期）

### 6.1 生命周期语义

* 插件“对外暴露能力”的唯一阶段
* 插件被视为**可见、可访问、可授权**

### 6.2 允许能力

* 注册菜单
* 注册 AMIS 页面（JSON）
* 注册权限模型

### 6.3 明确禁止

* ❌ 修改数据库结构
* ❌ 数据迁移
* ❌ 启动后台任务

### 6.4 抽象定义（规范）

```go
type EnableContext interface {
    BaseContext

    MenuRegistry() MenuRegistry
    PermissionRegistry() PermissionRegistry
    PageRegistry() AmisPageRegistry
}
```

---

## 7. RunContext（运行期）

### 7.1 生命周期语义

* 插件正常运行阶段
* 支撑巡检、Lua、Event、Webhook 等动态能力

### 7.2 允许能力

* 数据查询（只读或受限写）
* 启动后台任务 / 定时任务
* 监听事件、转发 Event / Webhook

### 7.3 明确禁止

* ❌ 修改数据库 Schema
* ❌ 注册菜单
* ❌ 注册权限
* ❌ 注册页面

### 7.4 抽象定义（规范）

```go
type RunContext interface {
    BaseContext

    DB() QueryExecutor
    Scheduler() Scheduler
    EventBus() EventBus
}
```

---

## 8. Context 使用的强制规则（规范级）

1. 插件 **不得对 Context 进行类型断言或强转**
2. 插件仅能访问当前生命周期提供的 Context
3. Context 不得暴露底层实现（如 *sql.DB）
4. 不同 Context 之间 **不得共享可变状态**
5. 所有能力必须通过 Context 间接获得

---

## 9. v1 明确不支持的能力

为保证体系稳定，k8m 插件体系 v1 明确不支持：

* Context 动态扩展
* Context 能力提升（Upgrade → Run）
* 插件之间共享 Context
* 插件访问其他插件的能力

---
 
