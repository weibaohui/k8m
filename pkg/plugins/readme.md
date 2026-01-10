# k8m 插件（Feature Module）架构定义 v1.1

> 本文档用于**先行固化 k8m 插件体系的抽象、边界与约束**，在此基础上再开展代码实现。
>
> 目标不是"灵活"，而是：**可控、可裁剪、可维护、可规模化扩展**。

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
3. 插件可启动 / 停止后台任务
4. 插件能力边界清晰、显式声明
5. 插件之间无隐式依赖
6. 插件描述以 Go 代码为主，编译期加载
7. 前端统一使用 AMIS JSON 作为渲染描述

---

## 2. 插件基本定义

### 2.1 插件定位

在 k8m 中：

> **插件不是 Hook，也不是轻量扩展，而是"可插拔子系统"。**

插件通常具备以下能力中的若干项：

* 菜单入口
* 前端页面（AMIS JSON）
* 后端 API
* 权限定义（RBAC）
* SQL 表结构或数据模型
* 初始化 / 清理逻辑
* 后台任务（协程、定时任务）

---

## 3. 插件生命周期与状态模型

### 3.1 生命周期流转图

插件具备完整、显式的生命周期：

```
Discover → Install → Enable → Start → Running
    ↓         ↓         ↓          ↓
    └────────┴─────────┴──────────┴──→ Disable → Uninstall
                        ↓
                    Stop → Stopped
```

### 3.2 状态定义与转换规则

插件在系统中具备以下状态：

#### StatusDiscovered（已发现）
* 插件在编译期通过 Go 注册，系统启动时完成元信息加载
* 可执行：Install
* 不可执行：Enable、Disable、Start、Stop、Uninstall

#### StatusInstalled（已安装）
* 数据库表已创建，基础数据已初始化
* 可执行：Enable、Uninstall
* 不可执行：Install、Disable、Start、Stop

#### StatusEnabled（已启用）
* 配置级别：菜单可见、API 可访问、前端 AMIS JSON 可加载
* 可执行：Start、Disable、Uninstall
* 不可执行：Install、Enable

#### StatusRunning（运行中）
* 运行时级别：后台任务已启动，定时任务执行中
* 可执行：Stop、Disable、Uninstall
* 不可执行：Install、Enable、Start

#### StatusStopped（已停止）
* 后台任务已停止，但插件仍处于启用状态
* 可执行：Start、Disable、Uninstall
* 不可执行：Install、Enable、Stop

#### StatusDisabled（已禁用）
* 菜单隐藏、API 不可访问，但数据和权限定义保留
* 可执行：Enable、Uninstall
* 不可执行：Install、Disable、Start、Stop

### 3.3 生命周期方法说明

#### Install（安装）
* **职责**：创建数据库表、初始化基础数据、注册权限模型
* **特性**：只执行一次，必须具有幂等性
* **状态变化**：Discovered → Installed

#### Upgrade（升级）
* **职责**：执行数据库迁移（表结构变更、数据迁移）、权限模型更新、版本兼容性处理
* **特性**：版本号变化时触发，不改变插件状态，必须具有幂等性
* **状态变化**：无（可在任何状态触发）

#### Enable（启用）
* **职责**：注册路由、暴露菜单、使 API 可访问
* **特性**：配置级能力暴露，不启动后台任务
* **状态变化**：Installed/Disabled → Enabled

#### Disable（禁用）
* **职责**：隐藏菜单、撤销路由、使 API 不可访问
* **特性**：不删除数据和权限定义，自动停止后台任务（如正在运行）
* **状态变化**：Enabled/Stopped → Disabled

#### Start（启动后台任务）
* **职责**：启动非阻塞后台协程、监听 EventBus 事件
* **调用时机**：系统启动时按依赖顺序启动、手动调用 StartPlugin API
* **特性**：不可阻塞，使用 context.Context 实现优雅停止
* **状态变化**：Enabled/Stopped → Running

#### Stop（停止后台任务）
* **职责**：停止后台协程、清理资源
* **调用时机**：手动调用 StopPlugin API、禁用插件、卸载插件
* **特性**：不可阻塞
* **状态变化**：Running → Stopped

#### StartCron（执行定时任务）
* **职责**：执行插件定义的定时任务逻辑
* **调用时机**：插件运行时（StatusRunning），根据 metadata 中的 cron 表达式触发
* **特性**：不可阻塞，每个定时任务独立执行

#### Uninstall（卸载）
* **职责**：根据 keepData 参数决定是否删除数据库表和数据、清理插件注册信息
* **特性**：自动停止后台任务（如正在运行），支持保留数据选项
* **状态变化**：Enabled/Disabled/Running/Stopped → Discovered

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
     ├── controller/          # 控制器（可选）
     │   └── *.go
     ├── service/             # 服务层（可选）
     │   └── *.go
     ├── admin/               # 插件管理员类操作实现（可选）
     ├── cluster/             # 集群类操作实现（可选）
     ├── mgm/                 # 管理类操作实现（可选）
     └── ...                  # 其他业务逻辑
```

---

## 6. 插件元信息与能力声明

### 6.1 元信息（Meta）

插件必须声明基础元信息，用于插件管理、版本控制、依赖判断、启动顺序控制：

* **Name**：插件唯一标识（系统级唯一，必填）
* **Title**：插件展示名称
* **Version**：插件版本号（用于触发 Upgrade）
* **Description**：插件功能描述

### 6.2 能力声明字段

* **Menus**：菜单声明（0..n），定义前端导航入口
* **Tables**：插件使用的数据库表名列表
* **Crons**：定时任务调度表达式（5段 cron 格式）
* **Dependencies**：强依赖插件列表，启用前必须确保所有依赖插件均已启用
* **RunAfter**：启动顺序约束，不依赖这些插件，但必须在它们之后启动
* **Lifecycle**：生命周期接口实现
* **ClusterRouter**：集群类操作路由注册回调（`/k8s/cluster/<cluster-id>/plugins/<plugin-name>/`）
* **ManagementRouter**：管理类操作路由注册回调（`/mgm/plugins/<plugin-name>/`）
* **PluginAdminRouter**：平台管理员类操作路由注册回调（`/admin/plugins/<plugin-name>/`）
* **RootRouter**：根路由注册回调（公开 API，一般不建议使用）

### 6.3 依赖与启动顺序

* **Dependencies**：强依赖关系
  * 启用插件前，必须确保所有依赖插件均已启用
  * 禁用插件前，必须确保没有其他插件依赖于当前插件
  * 系统启动时按依赖顺序启动插件（拓扑排序）
  * 插件之间不得直接相互调用内部实现，只能通过核心提供的公共能力交互

* **RunAfter**：启动顺序约束
  * 不表示依赖关系，仅表示启动顺序
  * 插件会在 RunAfter 列表中的插件之后启动
  * 系统启动时会综合考虑 Dependencies 和 RunAfter 进行拓扑排序

示例：
```go
Dependencies: []string{"plugin1", "plugin2"},  // 强依赖
RunAfter: []string{"leader"},                  // 仅顺序约束
```

---

## 7. 菜单与权限模型

### 7.1 菜单定义

插件可声明 0 个或多个菜单，用于前端导航、权限绑定、页面入口定位。

菜单字段说明：

* **Key**：菜单唯一标识
* **Title**：菜单展示标题
* **Icon**：图标（Font Awesome 类名）
* **URL**：跳转地址
* **EventType**：事件类型（'url' 或 'custom'）
* **CustomEvent**：自定义事件，如 `() => loadJsonPage("/path")`
* **Order**：排序号
* **Children**：子菜单
* **Show**：显示表达式（字符串形式的 JS 表达式，控制菜单可见性）
  * `isPlatformAdmin()`：判断是否为平台管理员
  * `isUserHasRole('role')`：判断用户是否有指定角色（guest/platform_admin）
  * `isUserInGroup('group')`：判断用户是否在指定组

> 菜单仅在插件 **Enable** 后可见。Show 表达式是菜单的显示权限。

### 7.2 路由与权限体系

插件通过四类 API 路由实现不同权限级别的操作：

#### 集群类操作（ClusterRouter）
* **访问路径**：`/k8s/cluster/<cluster-id>/plugins/<plugin-name>/xxxx`
* **权限要求**：必须是登录用户（系统自动注入登录校验）
* **适用场景**：针对集群操作的插件，如集群监控、集群配置等
* **特点**：路径自动注入具体的集群 ID

#### 管理类操作（ManagementRouter）
* **访问路径**：`/mgm/plugins/<plugin-name>/xxxx`
* **权限要求**：必须是登录用户（系统自动注入登录校验）
* **适用场景**：一般的管理类插件，如巡检配置等
* **特点**：无法获取到集群 ID

#### 平台管理员类操作（PluginAdminRouter）
* **访问路径**：`/admin/plugins/<plugin-name>/xxxx`
* **权限要求**：必须是平台管理员用户（系统自动注入登录校验和角色校验）
* **适用场景**：对整个平台进行操作的插件，如分布式功能
* **特点**：无法获取到集群 ID

#### 根路由 API（RootRouter）
* **访问路径**：`/xxxx`
* **权限要求**：必须是登录用户
* **适用场景**：公开的 API 接口
* **特点**：一般不建议使用，如需使用要特别注意注册路由的正确性

> **路由注册时机**：API 在 Enable 阶段注册，在 Disable 阶段不可访问

路由注册示例：

```go
ClusterRouter: func(cluster chi.Router) {
    g := cluster.Group("/plugins/" + pluginName)
    g.GET("/items", handler.List)
    g.POST("/items", handler.Create)
},

ManagementRouter: func(mgm chi.Router) {
    g := mgm.Group("/plugins/" + pluginName)
    g.GET("/config", handler.GetConfig)
    g.POST("/config", handler.SetConfig)
},

PluginAdminRouter: func(admin chi.Router) {
    g := admin.Group("/plugins/" + pluginName)
    g.GET("/settings", handler.GetSettings)
    g.POST("/settings", handler.SetSettings)
},
```

---

## 8. 前端模型（AMIS）定义

### 8.1 前端技术约束

* 插件前端 **只允许 AMIS JSON**
* 禁止插件引入 React / Vue 代码
* 禁止插件执行任意 JS 逻辑

### 8.2 前端加载方式

* 前端通过统一 API 获取 AMIS JSON
* 插件启用前，请求返回 404
* AMIS JSON 仅用于描述界面结构，不参与权限决策

---

## 9. 数据库管理规范

### 9.1 表名规范

* 表名必须包含插件名前缀，避免命名冲突
* 使用下划线分隔单词
* 示例：`plugin_name_items`
* 不允许修改其他插件或核心表结构

### 9.2 数据库操作工具

* 使用 GORM 进行数据库操作
* 使用 AutoMigrate 进行表结构管理
* 使用 Migrator.DropTable 删除表
* 所有数据库操作都必须具备幂等性

### 9.3 初始化（Install）

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

### 9.4 升级（Upgrade）

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

### 9.5 卸载删除数据（Uninstall with KeepData=false）

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

### 9.6 卸载保留数据（Uninstall with KeepData=true）

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

### 9.7 数据库操作示例

```go
type Item struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"size:255;not null" json:"name"`
    Description string    `gorm:"type:text" json:"description"`
    CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
    UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (Item) TableName() string {
    return "plugin_name_items"
}

func InitDB() error {
    return dao.DB().AutoMigrate(&Item{})
}

func DropDB() error {
    db := dao.DB()
    if db.Migrator().HasTable(&Item{}) {
        return db.Migrator().DropTable(&Item{})
    }
    return nil
}
```

---

## 10. 插件管理原则（重要）

* 插件不是热插拔组件
* 插件启停属于**运维级操作**，且启用、关闭需要重启k8m生效。
* 稳定性优先于动态性

---

## 11. 反设计原则（明确禁止）

以下行为在插件体系中**明确禁止**：

* 插件直接修改核心代码
* 插件私自注册全局路由
* 插件返回任意前端代码
* 插件绕过 RBAC 鉴权
* 插件跨模块访问数据库表

----

## 12. 运行上下文与事件总线

### 12.1 Context 设计

插件在生命周期方法中只能通过 Context 与系统交互。Context 是插件访问系统能力的**唯一入口**，用于隔离插件与核心实现。

> 插件不得直接操作核心内部对象。

### 12.2 EventBus 事件总线
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

**支持的事件类型：**
* EventLeaderElected：选举成为 Leader
* EventLeaderLost：失去 Leader 身份

**EventBus 特性：**
* Subscribe 返回一个只读 channel，用于接收事件
* Publish 会向所有订阅者发送事件，慢消费者的事件会被丢弃
* 每个订阅者的 channel 缓冲大小为 1，防止阻塞

---

## 13. 插件管理器职责

插件管理器是插件体系的唯一调度者，负责：

* **插件注册**：在编译期注册插件元信息
* **生命周期调度**：Install、Upgrade、Enable、Disable、Uninstall、Start、Stop、StartCron
* **插件状态管理**：维护插件状态，插件本身不得修改状态
* **插件依赖校验**：Dependencies（强依赖）和 RunAfter（启动顺序）
* **拓扑排序**：按依赖顺序启动插件
* **定时任务调度**：基于 cron 表达式调度 StartCron 方法
* **EventBus 管理**：为每个生命周期提供独立的事件总线实例

> Manager 不包含具体业务逻辑，仅负责流程与约束。

---

## 14. 插件开发最佳实践

### 14.1 后台任务管理

**最佳实践：**

* 使用 context.Context 实现优雅停止
* 在 Start 方法中保存 context.CancelFunc，在 Stop 方法中调用
* 后台任务应该监听 context.Done() 信号，及时退出
* 避免在后台任务中使用阻塞操作
* 使用 klog.V(6).Infof 打印日志

示例代码：

```go
type PluginLifecycle struct {
    cancelStart context.CancelFunc
}

func (l *PluginLifecycle) Start(ctx plugins.BaseContext) error {
    klog.V(6).Infof("启动插件后台任务")

    startCtx, cancel := context.WithCancel(context.Background())
    l.cancelStart = cancel

    go func(meta plugins.Meta) {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                klog.V(6).Infof("插件后台任务运行中，插件: %s，版本: %s", meta.Name, meta.Version)
            case <-startCtx.Done():
                klog.V(6).Infof("插件启动 goroutine 退出")
                return
            }
        }
    }(ctx.Meta())

    return nil
}

func (l *PluginLifecycle) Stop(ctx plugins.BaseContext) error {
    klog.V(6).Infof("停止插件后台任务")

    if l.cancelStart != nil {
        l.cancelStart()
        l.cancelStart = nil
    }

    return nil
}
```

### 14.2 定时任务管理

**声明定时任务：**

```go
Crons: []string{
    "* * * * *",      // 每分钟执行一次
    "*/2 * * * *",    // 每2分钟执行一次
}
```

**最佳实践：**

* 避免在 StartCron 中执行耗时操作，使用 goroutine 处理耗时任务
* 确保定时任务具有幂等性
* 使用 klog.V(6).Infof 打印日志

示例代码：

```go
func (l *PluginLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
    klog.V(6).Infof("执行插件定时任务，表达式: %s", spec)

    go func() {
        // 执行定时任务逻辑
    }()

    return nil
}
```

### 14.3 生命周期方法实现

* 确保所有生命周期方法具有幂等性
* 使用 klog.V(6).Infof 打印日志（使用中文）
* 返回明确的错误信息

### 14.4 错误处理

* 返回明确的错误信息
* 使用 klog.V(6).Infof 打印错误日志
* 避免使用 panic

### 14.5 资源管理

* 在 Start 方法中分配资源
* 在 Stop 方法中释放资源
* 使用 defer 确保资源释放

### 14.6 并发安全

* 使用 sync.Mutex 保护共享资源
* 避免在生命周期方法中使用阻塞操作
* 使用 goroutine 处理耗时任务

### 14.7 测试

* 编写单元测试
* 测试生命周期方法的幂等性
* 测试插件的依赖关系

---

## 15. 完整插件示例

```go
package demo

import (
    "context"
    "time"

    "github.com/weibaohui/k8m/pkg/plugins"
    "github.com/weibaohui/k8m/pkg/plugins/modules/demo/models"
    "github.com/weibaohui/k8m/pkg/plugins/modules/demo/route"
    "k8s.io/klog/v2"
)

var Metadata = plugins.Module{
    Meta: plugins.Meta{
        Name:        "demo",
        Title:       "演示插件",
        Version:     "1.0.0",
        Description: "演示插件功能",
    },
    Tables: []string{
        "demo_items",
    },
    Crons: []string{
        "* * * * *",
    },
    Menus: []plugins.Menu{
        {
            Key:   "plugin_demo_index",
            Title: "演示插件",
            Icon:  "fa-solid fa-cube",
            Order: 1,
            Children: []plugins.Menu{
                {
                    Key:         "plugin_demo_cluster",
                    Title:       "演示插件Cluster",
                    Icon:        "fa-solid fa-puzzle-piece",
                    EventType:   "custom",
                    CustomEvent: `() => loadJsonPage("/plugins/demo/cluster")`,
                    Order:       100,
                },
            },
        },
    },
    Dependencies: []string{},
    RunAfter: []string{
        "leader",
    },
    Lifecycle: &DemoLifecycle{},
    ClusterRouter: route.RegisterClusterRoutes,
    ManagementRouter: route.RegisterManagementRoutes,
    PluginAdminRouter: route.RegisterPluginAdminRoutes,
}

type DemoLifecycle struct {
    cancelStart context.CancelFunc
}

func (d *DemoLifecycle) Install(ctx plugins.InstallContext) error {
    if err := models.InitDB(); err != nil {
        klog.V(6).Infof("安装Demo插件失败: %v", err)
        return err
    }
    klog.V(6).Infof("安装Demo插件成功")
    return nil
}

func (d *DemoLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
    klog.V(6).Infof("升级Demo插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
    if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
        return err
    }
    return nil
}

func (d *DemoLifecycle) Enable(ctx plugins.EnableContext) error {
    klog.V(6).Infof("启用Demo插件")
    return nil
}

func (d *DemoLifecycle) Disable(ctx plugins.BaseContext) error {
    klog.V(6).Infof("禁用Demo插件")
    return nil
}

func (d *DemoLifecycle) Uninstall(ctx plugins.UninstallContext) error {
    klog.V(6).Infof("卸载Demo插件")
    if !ctx.KeepData() {
        if err := models.DropDB(); err != nil {
            return err
        }
        klog.V(6).Infof("卸载Demo插件完成，已删除相关表及数据")
    } else {
        klog.V(6).Infof("卸载Demo插件完成，保留相关表及数据")
    }
    return nil
}

func (d *DemoLifecycle) Start(ctx plugins.BaseContext) error {
    klog.V(6).Infof("启动Demo插件后台任务")

    startCtx, cancel := context.WithCancel(context.Background())
    d.cancelStart = cancel

    go func(meta plugins.Meta) {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                klog.V(6).Infof("Demo插件后台任务运行中，插件: %s，版本: %s", meta.Name, meta.Version)
            case <-startCtx.Done():
                klog.V(6).Infof("Demo 插件启动 goroutine 退出")
                return
            }
        }
    }(ctx.Meta())

    return nil
}

func (d *DemoLifecycle) Stop(ctx plugins.BaseContext) error {
    klog.V(6).Infof("停止Demo插件后台任务")

    if d.cancelStart != nil {
        d.cancelStart()
        d.cancelStart = nil
    }

    return nil
}

func (d *DemoLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
    klog.V(6).Infof("执行Demo插件定时任务，表达式: %s", spec)
    return nil
}
```

---

## 16. 总结

k8m 插件体系是一个完整、可控、可扩展的插件架构，具有以下特点：

1. **完整生命周期管理**：从发现到卸载，支持完整的插件生命周期
2. **配置与运行分离**：区分配置级别的启用/禁用和运行时级别的运行/停止
3. **依赖管理**：支持插件依赖声明和启动顺序控制
4. **权限控制**：通过菜单显示表达式和后端 API 显式校验实现权限控制
5. **后台任务管理**：支持后台协程和定时任务
6. **数据库管理**：支持数据库表创建、升级和删除
7. **路由管理**：支持多种类型的路由注册和权限控制
8. **事件总线**：支持插件间事件通信

通过遵循本文档的规范，开发者可以创建高质量、可维护、可扩展的插件。
