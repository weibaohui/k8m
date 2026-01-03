# k8m 插件（Feature Module）架构定义 v1.0

临时记录：
都是源码编译安装，没有插件升级概念，因为插件不能单独选择版本，因此是更新大版本后，各个插件，也必须更新。那么这就不是一个可选项。是需要自动处理好的。

> 本文档用于**先行固化 k8m 插件体系的抽象、边界与约束**，在此基础上再开展代码实现。
>
> 目标不是“灵活”，而是：**可控、可裁剪、可维护、可规模化扩展**。

---

## 1. 设计目标

k8m 插件体系用于解决以下问题：

* 功能模块数量持续增长，核心复杂度过高
* 不同部署环境对功能裁剪需求不同
* 新功能希望低侵入、可独立开发、可独立启停
* 前后端、权限、数据模型需要强一致

因此插件体系的设计目标是：

1. **一个插件 = 一个完整功能单元（Feature Module）**
2. 插件可安装 / 启用 / 禁用 / 卸载
3. 插件能力边界清晰、显式声明
4. 插件之间无隐式依赖
5. 插件描述以 Go 代码为主，编译期加载
6. 前端统一使用 AMIS JSON 作为渲染描述

---

## 2. 插件基本定义

### 2.1 插件定位

在 k8m 中：

> **插件不是 Hook，也不是轻量扩展，而是“可插拔子系统”。**

插件通常具备以下能力中的若干项：

* 菜单入口
* 前端页面（AMIS JSON）
* 后端 API
* 权限定义（RBAC）
* SQL 表结构或数据模型
* 初始化 / 清理逻辑

---

## 3. 插件生命周期模型

插件具备完整、显式的生命周期：

```
Discover → Install → Enable → Disable → Uninstall
```

### 3.1 Discover（发现）

* 插件在编译期通过 Go 注册
* 系统启动时完成插件元信息加载

### 3.2 Install（安装）

安装阶段负责**不可逆或重成本操作**：

* 创建数据库表
* 初始化基础数据
* 注册权限模型

> Install 只执行一次，具有幂等性要求。

### 3.3 Upgrade（升级）

升级阶段负责**版本变更时的安全迁移**：

* 执行数据库迁移（表结构变更、数据迁移）
* 权限模型更新
* 其他版本兼容性处理

> Upgrade 在版本号变化时触发，不改变插件状态。

### 3.4 Enable（启用）

启用阶段负责**运行期能力暴露**：

* 菜单可见
* API 可访问
* 前端 AMIS JSON 可加载

### 3.5 Disable（禁用）

禁用阶段负责**能力收敛**：

* 菜单隐藏
* API 不可访问
* 前端资源返回 404

> 禁用不删除数据、不删除权限定义。

### 3.6 Uninstall（卸载）

卸载阶段负责**彻底移除插件痕迹**（可选）：

* 删除数据库表（根据 keepData 参数决定）
* 删除初始化数据
* 清理插件注册信息

> 卸载后插件条目保留，状态变为 Discovered，可再次安装。

### 3.7 Start（启动后台任务）

启动阶段负责**启动插件的后台任务**：

* 启动非阻塞的后台协程
* 监听 EventBus 事件
* 启动定时任务

> Start 在系统启动时调用，不可阻塞。

### 3.8 StartCron（执行定时任务）

定时任务执行阶段负责**执行插件定义的定时任务**：

* 根据 metadata 中定义的 cron 表达式触发
* 执行具体的定时任务逻辑

> StartCron 由系统统一调度，不可阻塞。

### 转换关系

* Discover → Install → Enable → Disable → Uninstall
* Upgrade 可在任何状态触发（版本变更时）

---

## 4. 插件描述方式（核心约束）

### 4.1 描述语言约束

* **除 AMIS JSON 外，所有插件描述必须使用 Go 代码**
* 禁止使用 YAML / JSON 描述插件结构

原因：

* 编译期校验
* IDE 自动补全
* 可重构
* 可审计
* 避免运行期解析错误

---

## 5. 插件目录结构规范

```text
modules/
 └── <plugin-name>/
     ├── metadata.go          # 插件元信息与能力声明
     ├── lifecycle.go         # 生命周期实现
     ├── models/              # 数据模型定义
     │   ├── db.go           # 数据库初始化/升级/删除
     │   └── *.go            # 具体模型定义
     ├── route/               # 路由注册
     │   ├── cluster_api.go  # 集群类操作路由
     │   ├── mgm_api.go      # 管理类操作路由
     │   └── admin_api.go    # 插件管理员类操作路由
     ├── frontend/            # 前端 AMIS JSON
     │   └── *.json
     ├── admin/               # 插件管理员类操作实现
     ├── cluster/             # 集群类操作实现
     ├── mgm/                 # 管理类操作实现
     └── ...                  # 其他业务逻辑
```

---

## 6. 插件元信息定义

插件必须声明基础元信息，用于：

* 插件管理
* 版本控制
* 依赖判断
* 启动顺序控制

插件元信息包含：

* Name：插件唯一标识（系统级唯一）
* Title：插件展示名称
* Version：插件版本号
* Description：插件功能描述

> 插件名称在系统内必须唯一。

插件能力声明包含：

* Menus：菜单声明（0..n）
* Dependencies：插件依赖的其他插件名称列表；启用前需确保均已启用
* RunAfter：不依赖 RunAfter 中的插件，但是必须在它们之后启动
* Crons：插件的定时任务调度表达式（5段 cron）
* Tables：插件使用的数据库表名列表
* ClusterRouter：集群类操作路由注册回调
* ManagementRouter：管理类操作路由注册回调
* PluginAdminRouter：插件管理员类操作路由注册回调
* RootRouter：根路由注册回调（公开API，无需登录）
* Lifecycle：生命周期实现

---

## 7. 菜单模型定义

插件可声明 0 个或多个菜单。

菜单用于：

* 前端导航
* 权限绑定
* 页面入口定位

菜单模型包含以下字段：

* Key：菜单唯一标识
* Title：菜单展示标题
* Icon：图标（Font Awesome 类名）
* URL：跳转地址
* EventType：事件类型（'url' 或 'custom'）
* CustomEvent：自定义事件，如：`() => loadJsonPage("/path")`
* Order：排序号
* Children：子菜单
* Show：显示表达式（字符串形式的 JS 表达式）
  * `isPlatformAdmin()`：判断是否为平台管理员
  * `isUserHasRole('role')`：判断用户是否有指定角色（guest/platform_admin）
  * `isUserInGroup('group')`：判断用户是否在指定组

菜单仅在插件 **Enable** 后可见。

> 注意：Show 表达式是菜单的显示权限。后端 API 业务逻辑需调用 service.AuthService().EnsureUserIsPlatformAdmin(*gin.Context) 等方法进行显式权限校验，后端 API 的权限校验不能依赖此表达式。

---

## 8. 前端模型（AMIS）定义

### 8.1 前端技术约束

* 插件前端 **只允许 AMIS JSON**
* 禁止插件引入 React / Vue 代码
* 禁止插件执行任意 JS 逻辑

### 8.2 前端加载方式

* 前端通过统一 API 获取 AMIS JSON
* 插件启用前，请求返回 404

AMIS JSON 仅用于描述界面结构，不参与权限决策。

---

## 9. 权限模型（RBAC）

插件通过三类 API 路由实现不同权限级别的操作：

### 9.1 集群类操作（ClusterRouter）

* 访问路径：`/k8s/cluster/<cluster-id>/plugins/<plugin-name>/xxxx`
* 权限要求：必须是登录用户
* 适用场景：针对集群操作的插件，如集群监控、集群配置等
* 特点：路径自动注入具体的集群 ID

### 9.2 管理类操作（ManagementRouter）

* 访问路径：`/mgm/plugins/<plugin-name>/xxxx`
* 权限要求：必须是登录用户
* 适用场景：一般的管理类插件，如巡检配置等
* 特点：无法获取到集群 ID

### 9.3 平台管理员类操作（PluginAdminRouter）

* 访问路径：`/admin/plugins/<plugin-name>/xxxx`
* 权限要求：必须是平台管理员用户
* 适用场景：对整个平台进行操作的插件，如分布式功能
* 特点：无法获取到集群 ID

### 9.4 公开 API（RootRouter）

* 访问路径：`/xxxx`
* 权限要求：无需登录即可访问
* 适用场景：公开的 API 接口
* 特点：一般不建议使用，如需使用要特别注意注册路由的正确性

> 后端 API 的权限校验不能依赖菜单的 Show 表达式，必须在业务逻辑中显式调用权限校验方法。

---

## 10. 后端 API 规范

插件后端 API 必须满足：

* 路径以 `/k8s/cluster/<cluster-id>/plugins/<plugin-name>/` 开头（集群类操作）
* 路径以 `/mgm/plugins/<plugin-name>/` 开头（管理类操作）
* 路径以 `/admin/plugins/<plugin-name>/` 开头（平台管理员类操作）
* 路径以 `/` 开头（公开 API，一般不建议使用）
* API 在 Enable 阶段注册
* API 在 Disable 阶段不可访问
* 插件 API 不允许绕过统一鉴权与审计体系

插件通过以下路由注册回调定义 API：

* ClusterRouter：注册集群类操作路由
* ManagementRouter：注册管理类操作路由
* PluginAdminRouter：注册插件管理员类操作路由
* RootRouter：注册根路由（公开 API）

路由注册示例：

```go
ClusterRouter: func(cluster *gin.RouterGroup) {
    g := cluster.Group("/plugins/" + pluginName)
    g.GET("/items", handler.List)
    g.POST("/items", handler.Create)
}
```

---

## 11. SQL / 数据模型规范

插件可声明：

* 独立数据库表
* 数据初始化逻辑

要求：

* 表名前缀必须包含插件名
* SQL 定义具备幂等性
* 不允许修改其他插件或核心表结构

### 11.1 初始化（Install）

初始化阶段负责创建数据库表结构和初始化基础数据：

* 使用 GORM 的 AutoMigrate 自动创建表结构
* 初始化基础数据（如果有）
* 保证幂等性（可重复执行）

示例代码：

```go
func InitDB() error {
    return dao.DB().AutoMigrate(&Item{})
}
```

### 11.2 升级（Upgrade）

升级阶段负责版本变更时的数据迁移：

* 使用 GORM 的 AutoMigrate 自动迁移表结构
* 根据版本号进行安全迁移
* 可在 UpgradeDB 函数中处理复杂的数据迁移逻辑
* 保证幂等性（可重复执行）

示例代码：

```go
func UpgradeDB(fromVersion string, toVersion string) error {
    klog.V(6).Infof("开始升级插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
    if err := dao.DB().AutoMigrate(&Item{}); err != nil {
        klog.V(6).Infof("自动迁移插件数据库失败: %v", err)
        return err
    }
    klog.V(6).Infof("升级插件数据库完成")
    return nil
}
```

### 11.3 卸载删除数据（Uninstall with KeepData=false）

卸载阶段负责彻底移除插件痕迹：

* 使用 GORM Migrator.DropTable 删除所有相关表
* 删除所有相关数据
* 清理插件注册信息
* 插件状态变为 Discovered，可再次安装

示例代码：

```go
func DropDB() error {
    db := dao.DB()
    if db.Migrator().HasTable(&Item{}) {
        if err := db.Migrator().DropTable(&Item{}); err != nil {
            klog.V(6).Infof("删除插件表失败: %v", err)
            return err
        }
        klog.V(6).Infof("已删除插件表及数据")
    }
    return nil
}
```

在 Uninstall 生命周期方法中：

```go
func (l *PluginLifecycle) Uninstall(ctx plugins.UninstallContext) error {
    if !ctx.KeepData() {
        if err := models.DropDB(); err != nil {
            return err
        }
        klog.V(6).Infof("卸载插件完成，已删除相关表及数据")
    }
    return nil
}
```

### 11.4 卸载保留数据（Uninstall with KeepData=true）

卸载阶段保留数据：

* 不删除表和数据
* 只清理插件注册信息
* 插件状态变为 Discovered，可再次安装
* 再次安装时，数据仍然存在

示例代码：

```go
func (l *PluginLifecycle) Uninstall(ctx plugins.UninstallContext) error {
    if ctx.KeepData() {
        klog.V(6).Infof("卸载插件完成，保留相关表及数据")
    }
    return nil
}
```

### 11.5 数据处理最佳实践

* 表名必须包含插件名前缀，避免命名冲突
* 使用 GORM 的 AutoMigrate 进行表结构管理
* UpgradeDB 函数应处理版本间的数据迁移逻辑
* DropDB 函数应删除所有相关表，确保彻底清理
* 在 Uninstall 中根据 KeepData 参数决定是否删除数据
* 所有数据库操作都应具备幂等性

---

## 12. 插件之间的关系约束

* 插件之间 **不得直接相互调用内部实现**
* 允许通过核心提供的公共能力交互
* 插件依赖关系必须显式声明（Dependencies 字段）
* 插件启动顺序可以通过 RunAfter 字段控制
  * Dependencies：强依赖关系，启用前必须确保所有依赖插件均已启用
  * RunAfter：启动顺序约束，不依赖这些插件，但必须在它们之后启动
* 系统启动时会按依赖顺序启动插件（拓扑排序）
* 禁用插件时需要检查是否有其他插件依赖于当前插件

---

## 13. 插件管理原则（重要）

* 插件不是热插拔组件
* 插件启停属于**运维级操作**，且启用、关闭需要重启k8m生效。
* 稳定性优先于动态性

---

## 14. 反设计原则（明确禁止）

以下行为在插件体系中**明确禁止**：

* 插件直接修改核心代码
* 插件私自注册全局路由
* 插件返回任意前端代码
* 插件绕过 RBAC 鉴权
* 插件跨模块访问数据库表

---

## 15. 核心抽象与接口定义

> 本节用于**固化插件体系在代码层面的抽象模型**，只定义数据结构与接口签名，
> 不涉及任何具体实现逻辑。

---

### 15.1 插件模块（Feature Module）核心定义

插件在代码层面被抽象为一个 **Module**，用于描述插件的能力集合。

Module 只负责“声明”，不负责“执行”。

核心要素包括：

* 插件元信息（Meta）
* 生命周期回调（Lifecycle）
* 可选能力声明（菜单 / 权限 / API / SQL / 前端）

---

### 15.2 插件元信息（Meta）

插件元信息用于插件管理与系统识别，必须在编译期确定。

Meta 包含以下字段：

* Name：插件唯一标识（系统级唯一）
* Title：插件展示名称
* Version：插件版本号
* Description：插件功能描述

Meta 不参与业务逻辑，仅用于管理与展示。

---

### 15.3 生命周期接口（Lifecycle）

插件生命周期通过显式接口定义，禁止隐式行为。

生命周期接口包括：

* Install：安装阶段，只执行一次，必须幂等
* Upgrade：升级阶段，当版本变化时触发，用于安全迁移
* Enable：启用阶段，注册运行期能力
* Disable：禁用阶段，撤销运行期能力
* Uninstall：卸载阶段，清理插件资源（可选）
* Start：启动阶段，用于启动后台任务。按依赖顺序启动各插件
* StartCron：启动定时任务，用于执行定时任务逻辑

生命周期方法由系统统一调度，插件不得自行调用。

---

### 15.4 运行上下文（Context）

插件在生命周期方法中只能通过 Context 与系统交互。

Context 是插件访问系统能力的**唯一入口**，用于隔离插件与核心实现。

Context 包含但不限于以下能力入口：

* EventBus 事件总线
  ```go 
  // 发布事件
  ctx.Bus().Publish(eventbus.Event{
      Type: eventbus.EventLeaderElected,
      Data: any, // 可选的事件数据
  })
  
  // 订阅事件
  elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
  lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)
  
  // 监听多个 channel，根据 channel 的信号启动或停止事件转发
  go func() {
      for {
          select {
          case <-elect:
              klog.V(6).Infof("成为Leader")
          case <-lost:
              klog.V(6).Infof("不再是Leader")
          }
      }
  }()
  ```

  EventBus 支持以下事件类型：
  * EventLeaderElected：选举成为 Leader
  * EventLeaderLost：失去 Leader 身份

  EventBus 特性：
  * Subscribe 返回一个只读 channel，用于接收事件
  * Publish 会向所有订阅者发送事件，慢消费者的事件会被丢弃
  * 每个订阅者的 channel 缓冲大小为 1，防止阻塞

插件不得直接操作核心内部对象。

---

### 15.5 菜单能力声明（Menu Capability）

插件可声明 0 个或多个菜单项。

菜单声明仅描述菜单结构，不包含渲染逻辑。

菜单字段包括：

* Key：菜单唯一标识
* Title：菜单展示标题
* Icon：图标（Font Awesome 类名）
* URL：跳转地址
* EventType：事件类型（'url' 或 'custom'）
* CustomEvent：自定义事件，如：`() => loadJsonPage("/path")`
* Order：排序号
* Children：子菜单
* Show：显示表达式（字符串形式的 JS 表达式）
  * `isPlatformAdmin()`：判断是否为平台管理员
  * `isUserHasRole('role')`：判断用户是否有指定角色（guest/platform_admin）
  * `isUserInGroup('group')`：判断用户是否在指定组

菜单仅在插件 Enable 后生效。

> 注意：Show 表达式是菜单的显示权限。后端 API 业务逻辑需调用 service.AuthService().EnsureUserIsPlatformAdmin(*gin.Context) 等方法进行显式权限校验，后端 API 的权限校验不能依赖此表达式。

---

### 15.6 前端能力声明（Frontend Capability）

插件前端能力以 **AMIS JSON 资源集合** 的形式存在。

前端能力声明用于：

* 标识插件是否提供前端界面
* 确定前端资源的访问范围

插件前端资源只允许被系统统一加载，不允许插件自行返回。

---

### 15.7 权限能力声明（RBAC Capability）

插件通过菜单的 Show 表达式和后端 API 的显式权限校验实现权限控制。

权限控制方式：

* 菜单显示权限：通过 Show 表达式控制菜单的可见性
  * `isPlatformAdmin()`：判断是否为平台管理员
  * `isUserHasRole('role')`：判断用户是否有指定角色（guest/platform_admin）
  * `isUserInGroup('group')`：判断用户是否在指定组

* 后端 API 权限：在业务逻辑中显式调用权限校验方法
  * `service.AuthService().EnsureUserIsPlatformAdmin(*gin.Context)`：确保用户是平台管理员
  * 其他自定义权限校验方法

权限在插件启用时生效，禁用时不删除权限定义。

---

### 15.8 API 能力声明（Backend API Capability）

插件可声明后端 API 能力。

API 能力声明用于：

* 标识插件是否提供后端接口
* 确定 API 的注册与撤销时机

插件 API 必须统一挂载在插件命名空间下。

---

### 15.9 SQL / 数据能力声明（Data Capability）

插件可声明独立的数据模型。

数据能力声明包括：

* 表结构定义
* 初始化数据定义（可选）

数据能力只在 Install / Uninstall 阶段生效。

---

### 15.10 插件管理器（Manager）职责定义

插件管理器是插件体系的唯一调度者，负责：

* 插件注册
* 生命周期调度（Install、Upgrade、Enable、Disable、Uninstall、Start、StartCron）
* 插件状态管理
* 插件依赖校验（Dependencies 和 RunAfter）
* 拓扑排序（按依赖顺序启动插件）
* 定时任务调度（基于 cron 表达式）
* EventBus 管理（为每个生命周期提供独立的事件总线实例）

Manager 不包含具体业务逻辑，仅负责流程与约束。

---

### 15.11 插件状态模型

插件在系统中具备以下状态之一：

* Discovered：已发现，未安装
* Installed：已安装，未启用
* Enabled：已启用
* Disabled：已禁用

插件状态由系统维护，插件本身不得修改。

状态转换关系：

* Discover → Install：插件从已发现状态变为已安装状态
* Install → Enable：插件从已安装状态变为已启用状态
* Enable → Disable：插件从已启用状态变为已禁用状态
* Disable → Enable：插件从已禁用状态变为已启用状态
* Enabled/Disabled → Uninstall：插件从已启用或已禁用状态变为已发现状态
* Upgrade：可在任何状态触发（版本变更时，不改变状态）

---

### 15.12 设计约束回顾（再次强调）

* 插件是 Feature Module，不是 Hook
* 插件描述必须编译期确定
* 插件通过 Context 与系统交互
* 插件前端只允许 AMIS JSON
* 插件生命周期必须显式

---

## 16. 后续工作计划

在本文档确认后，将依次开展：

1. 插件核心接口与 Manager 代码实现
2. AMIS Schema 加载与校验机制
3. 插件管理 API 设计
4. 示例插件（inspect）完整实现
5. 《k8m 插件开发指南》

---

> **本定义文档作为 k8m 插件体系的唯一权威规范，在 v1.0 范围内保持向后兼容。**
