// 定义需要过滤的 key 列表
const immutableKeys = [
    "cpu.request",
    "cpu.requestFraction",
    "cpu.limit",
    "cpu.limitFraction",
    "cpu.total",
    "memory.request",
    "memory.requestFraction",
    "memory.limit",
    "memory.limitFraction",
    "memory.total",
    "ip.usage.total",
    "ip.usage.used",
    "ip.usage.available",
    "pod.count.total",
    "pod.count.used",
    "pod.count.available",
    "kubectl.kubernetes.io/last-applied-configuration",
    "kom.kubernetes.io/restartedAt"
];
const FilterAnnotations = (input: Record<string, string>) => {
    // 过滤掉 immutableKeys 中的 key
    return Object.fromEntries(
        Object.entries(input).filter(([key]) => !immutableKeys.includes(key))
    );
};
export default FilterAnnotations;
