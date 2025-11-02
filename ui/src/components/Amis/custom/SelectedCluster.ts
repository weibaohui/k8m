import { getCurrentClusterId } from "@/utils/utils.ts";

/**
 * 读取当前选择的集群ID（selectedCluster）。
 *
 * 作为 AMIS 过滤器使用：
 * - 在表达式中写 `${''|selectedCluster}` 可直接返回当前URL中的集群ID；
 * - 可传入一个兜底值：`${'my-cluster'|selectedCluster}`，当未选择时返回该兜底值；
 *
 * 设计目标：统一从 URL 解析已选集群，便于与路由保持一致。
 *
 * @param fallback 可选的兜底集群ID字符串
 * @returns 集群ID字符串；优先返回 `getCurrentClusterId()`，否则返回 fallback 或空字符串
 */
const SelectedCluster = (fallback?: unknown): string => {
    try {
        const raw = (typeof window !== 'undefined') ? getCurrentClusterId() : '';
        const cluster = (raw ?? '').trim();
        if (cluster) return cluster;

        const fb = (typeof fallback === 'string') ? fallback.trim() : '';
        return fb || '';
    } catch {
        const fb = (typeof fallback === 'string') ? (fallback as string).trim() : '';
        return fb || '';
    }
}

export default SelectedCluster;