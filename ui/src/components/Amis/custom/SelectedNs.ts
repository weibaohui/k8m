/**
 * 读取当前选择的命名空间（selectedNs）。
 *
 * 作为 AMIS 过滤器使用：
 * - 在表达式中写 `${''|selectedNs}` 可直接返回 localStorage 中的命名空间；
 * - 可传入一个兜底值：`${'default'|selectedNs}`，当未设置或为空时返回该兜底值；
 *
 * 设计目标：替代 `${ls:selectedNs||'default'}`，并更易扩展（可接收兜底值）。
 *
 * @param fallback 可选的兜底命名空间字符串，例如 'default'
 * @returns 命名空间字符串；优先返回 localStorage('selectedNs')，否则返回 fallback 或 'default'
 */
const SelectedNs = (fallback?: unknown): string => {
    try {
        console.log('selectedNs fallback:', fallback);
        const raw = (typeof window !== 'undefined') ? window.localStorage.getItem('selectedNs') : null;
        const ns = (raw ?? '').trim();
        if (ns) return ns;

        const fb = (typeof fallback === 'string') ? fallback.trim() : '';
        return fb || 'default';
    } catch {
        const fb = (typeof fallback === 'string') ? (fallback as string).trim() : '';
        return fb || 'default';
    }
}

export default SelectedNs;