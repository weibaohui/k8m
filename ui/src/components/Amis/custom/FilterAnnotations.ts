// 定义需要过滤的 key 列表
const immutableKeys = [
    "cpu.request",
    "cpu.requestFraction",
    "cpu.limit",
    "cpu.limitFraction",
    "cpu.total",
    "cpu.realtime",
    "memory.request",
    "memory.requestFraction",
    "memory.limit",
    "memory.limitFraction",
    "memory.total",
    "memory.realtime",
    "ip.usage.total",
    "ip.usage.used",
    "ip.usage.available",
    "pod.count.total",
    "pod.count.used",
    "pod.count.available",
    "kubectl.kubernetes.io/last-applied-configuration",
    "kom.kubernetes.io/restartedAt",
    "pvc.count",
    "pv.count",
    "ingress.count",
];
const FilterAnnotations = (input: Record<string, string>) => {
    // 如果 input 不存在，则返回空对象
    if (!input) return {};
    // 如果是undefinded，则返回空对象
    if (input === undefined) return {};
    // 如果 input 是空对象，则返回空对象
    if (Object.keys(input).length === 0) return {};
    // 如果 input 不是 Record<string, string> 类型，则返回空对象
    if (typeof input !== "object") return {};

    // 过滤掉 immutableKeys 中的 key
    return Object.fromEntries(
        Object.entries(input).filter(([key]) => !immutableKeys.includes(key))
    );
};
export default FilterAnnotations;
