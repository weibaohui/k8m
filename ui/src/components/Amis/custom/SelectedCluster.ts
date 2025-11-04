import { getCurrentClusterId, getCurrentClusterIdInBase64 } from "@/utils/utils.ts";

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

/**
 * 读取当前选择的集群ID（Base64 编码）。
 *
 * 作为 AMIS 过滤器使用时，可直接返回 `getCurrentClusterIdInBase64()` 的结果，
 * 无兜底参数，未选择集群时返回空字符串。
 *
 * @returns Base64（URL 安全）编码的当前集群ID；未选择时返回空字符串
 */
export const SelectedClusterBase64 = (): string => {
    try {
        return (typeof window !== 'undefined') ? getCurrentClusterIdInBase64() : '';
    } catch {
        return '';
    }
}

export default SelectedCluster;