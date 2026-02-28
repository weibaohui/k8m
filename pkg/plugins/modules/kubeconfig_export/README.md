# Kubeconfig 导出插件

## 功能说明

本插件提供了 Kubernetes 集群的 kubeconfig 生成和导出功能，方便用户管理和分发集群访问凭证。

## 主要功能

1. **Kubeconfig 管理**
   - 查看所有已导入的 kubeconfig 模板列表
   - 显示集群的基本信息（服务器地址、用户、命名空间等）

2. **Kubeconfig 导出**
   - 支持按集群 ID 生成 kubeconfig
   - 支持限制命名空间（Namespace）
   - 支持导出为 YAML 格式文件
   - 支持自定义文件名

## API 路由

### 管理类路由 (mgm)

- `GET /mgm/plugins/kubeconfig_export/templates` - 获取 kubeconfig 模板列表
- `GET /mgm/plugins/kubeconfig_export/cluster/{clusterID}/kubeconfig` - 获取指定集群的 kubeconfig 信息
- `GET /mgm/plugins/kubeconfig_export/kubeconfig/{id}` - 根据 ID 获取 kubeconfig

### 集群类路由 (cluster)

- `POST /k8s/cluster/{clusterID}/plugins/kubeconfig_export/generate` - 为集群生成 kubeconfig
- `GET /k8s/cluster/{clusterID}/plugins/kubeconfig_export/export` - 导出集群 kubeconfig 文件

## 插件依赖

- `leader` 插件（用于集群管理）

## 使用方法

1. 启用插件后，在左侧菜单会出现 "Kubeconfig 导出" 菜单
2. 点击 "Kubeconfig 管理" 可以查看所有 kubeconfig 模板
3. 在列表中点击 "导出 Kubeconfig" 按钮，可以选择导出参数并下载 kubeconfig 文件

## 未来扩展

- 支持根据角色（admin/edit/view）生成有限制的 kubeconfig
- 支持设置有效期（Duration）
- 支持批量导出
- 支持自定义 kubeconfig 内容