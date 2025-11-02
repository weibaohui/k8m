# 集群与命名空间切换说明

本文档说明前端关于“集群切换”和“命名空间（Namespace）切换”的实现方式，包括：
- 工具函数（`ui/src/utils/utils.ts`）如何获取/设置集群与命名空间；
- AMIS 过滤器 `selectedCluster` 与 `selectedNs` 的用法与示例；
- 在 AMIS Schema 与普通 React 组件中如何串接使用。

## 背景与目标
- 集群与命名空间是强相关的：不同集群拥有不同的命名空间集合。
- 为了在前端统一行为：
  - 集群 ID 统一通过 URL 路径解析与设置（保证与路由一致）。
  - 命名空间在 `localStorage` 中按“集群维度”隔离存储（避免跨集群污染）。

---

## 工具函数（utils.ts）
文件路径：`ui/src/utils/utils.ts`

- `getCurrentClusterId(): string`
  - 从当前 URL 的路径中解析已选集群 ID（形如：`/cluster/<base64ClusterId>/...`）。
  - 解析流程：从路径段提取 Base64（URL 安全）编码的集群 ID，解码后返回。
  - 若无法解析或未选择集群，返回空字符串。

- `setCurrentClusterId(clusterId: string): void`
  - 设置当前集群，并进行跳转：把 `clusterId` 进行 Base64（URL 安全）编码，拼接到路径 `/cluster/<encoded>/...`。
  - 若当前路径已有集群段，则替换该段并保持后续路径与哈希不变；否则跳转到集群首页。

- `getSelectedNS(overrideClusterId?: string): string`
  - 获取当前选中的命名空间，按集群维度隔离：
    - 读取 `localStorage` 键：`selectedNS_${clusterId}`。
    - `clusterId` 默认取 `getCurrentClusterId()`，也可通过 `overrideClusterId` 指定。
  - 若未设置或无法读取，返回空字符串。

- `setSelectedNS(ns: string, overrideClusterId?: string): void`
  - 设置当前选中的命名空间，按集群维度隔离：
    - 写入 `localStorage` 键：`selectedNS_${clusterId}`。
    - `clusterId` 默认取 `getCurrentClusterId()`，也可通过 `overrideClusterId` 指定。

- 方法已暴露到 `window`（浏览器环境）以便脚本直接调用：
  - `window.getCurrentClusterId`
  - `window.setCurrentClusterId`
  - `window.getSelectedNS`
  - `window.setSelectedNS`

> 说明：命名空间键名采用 `selectedNS_${clusterId}`，确保不同集群的命名空间选择互不影响。

---

## AMIS 过滤器

### SelectedCluster 过滤器
文件路径：`ui/src/components/Amis/custom/SelectedCluster.ts`

- 作用：返回当前 URL 中的集群 ID。
- 用法：
  - 基本：`${''|selectedCluster}` 获取当前集群 ID；未选择时返回空字符串。
  - 兜底：`${'my-cluster'|selectedCluster}` 未选择时返回 `my-cluster`。
- 实现：内部调用 `getCurrentClusterId()`。

### SelectedNs 过滤器
文件路径：`ui/src/components/Amis/custom/SelectedNs.ts`

- 作用：返回当前集群维度下已选命名空间。
- 用法：
  - 基本：`${''|selectedNs}` 获取当前命名空间；未设置时返回空字符串。
  - 兜底：`${'default'|selectedNs}` 未设置时返回 `default`。
- 实现：内部调用 `getSelectedNS()`，从 `localStorage` 读取 `selectedNS_${clusterId}`。

> 这两个过滤器已在 `ui/src/components/Amis/index.tsx` 通过 `registerFilter` 注册（`selectedCluster`、`selectedNs`），可直接在 AMIS Schema 中使用。

---

## 在 AMIS Schema 中的示例

### 作为接口参数
```json
{
  "type": "service",
  "api": {
    "url": "/k8s/pod/list",
    "method": "get",
    "params": {
      "cluster": "${''|selectedCluster}",
      "namespace": "${''|selectedNs}"
    }
  }
}
```

### 作为表单默认值
```json
{
  "type": "form",
  "title": "创建资源",
  "controls": [
    {
      "type": "text",
      "name": "cluster",
      "label": "集群",
      "value": "${''|selectedCluster}"
    },
    {
      "type": "text",
      "name": "namespace",
      "label": "命名空间",
      "value": "${'default'|selectedNs}"
    }
  ]
}
```

### 与占位符拼接或兜底
```json
{
  "type": "service",
  "api": {
    "url": "/k8s/cluster/${''|selectedCluster}/ns/${'default'|selectedNs}/summary",
    "method": "get"
  }
}
```

---

## 在 React 组件中的示例
```ts
import { getCurrentClusterId, setCurrentClusterId, getSelectedNS, setSelectedNS } from "@/utils/utils";

// 读取当前集群与命名空间
const clusterId = getCurrentClusterId();
const ns = getSelectedNS();

// 设置命名空间（按集群隔离）
setSelectedNS("kube-system");

// 切换集群并保留当前页面路径/哈希
setCurrentClusterId("prod-cluster-01");
```

---

## 注意事项
- 建议在 AMIS 过滤器调用中提供兜底值，以避免空字符串导致接口参数不完整。
- 命名空间存储与读取均以“当前集群”为前提；切换集群后会读取该集群对应的命名空间记录。
- 如果历史逻辑中曾直接使用 `localStorage.getItem('selectedNs')`，请迁移为 `getSelectedNS()` 与过滤器 `selectedNs`，以获得按集群隔离的行为。

---

## 相关文件路径
- `ui/src/utils/utils.ts`
- `ui/src/components/Amis/custom/SelectedNs.ts`
- `ui/src/components/Amis/custom/SelectedCluster.ts`
- `ui/src/components/Amis/index.tsx`