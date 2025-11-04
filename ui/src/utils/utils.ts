export const sleep = (time = 10, fn?: () => void) => {
    return new Promise((resolve) => {
        const timer = setTimeout(() => {
            clearTimeout(timer);
            fn && typeof fn === 'function' && fn();
            resolve(true);
        }, time);
    });
};

export const isDev = import.meta.env.MODE === 'development'

export function replacePlaceholders(url: string, data: Record<string, any>): string {
    return url.replace(/\$\{([^}]+)\}/g, (_, key) => {
        const keys = key.split('.');
        let value: any = data;

        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                return _; // 如果找不到值，返回原始占位符
            }
        }

        return String(value); // 返回字符串形式的值
    });
}


export function appendQueryParam(url: string, params: Record<string, string>): string {
    const queryString = Object.keys(params)
        .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
        .join('&');
    return url.includes('?') ? `${url}&${queryString}` : `${url}?${queryString}`;
}

export function formatFinalGetUrl(props: {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>
}): string {
    let url = props.url;
    if (props.data != null && props.data.length != 0) {
        url = replacePlaceholders(props.url, props.data);
    }
    //如果param 为空，则直接返回url
    if (props.params == null || Object.keys(props.params).length === 0) {
        return url;
    }
    const params = Object.keys(props.params).reduce<Record<string, string>>((acc, key) => {
        const value = props.params[key];

        if (value.startsWith("${ls:")) {
            const localStorageValue = parseLocalStorageExpression(value);
            if (localStorageValue !== null) {
                acc[key] = localStorageValue;
            }
        } else {
            acc[key] = replacePlaceholders(value, props.data);
        }
        return acc;
    }, {});

    return appendQueryParam(url, params);
}

function parseLocalStorageExpression(expression: string): string | null {
    const match = expression.match(/^\${ls:(.+)}$/);
    return match ? localStorage.getItem(match[1]) : null;
}


export function toUrlSafeBase64(str: string) {
// 先转 UTF-8，再 btoa
    const utf8Str = unescape(encodeURIComponent(str));
    const base64 = btoa(utf8Str);
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function fromUrlSafeBase64(str: string): string {
    try {
        const base64 = str.replace(/-/g, '+').replace(/_/g, '/');
        // 补足 '='
        const padLength = (4 - (base64.length % 4)) % 4;
        const padded = base64 + '='.repeat(padLength);
        return atob(padded);
    } catch (e) {
        return '';
    }
}

export function ProcessK8sUrlWithCluster(url: string, overrideCluster?: string): string {
    // 仅处理 /k8s 开头的接口
    if (!url.startsWith('/k8s')) {
        return url;
    }
    // 已经带有 /k8s/cluster/:cluster 的，避免重复插入
    if (url.startsWith('/k8s/cluster/')) {
        return url;
    }

    // 选择覆盖的 cluster，否则使用本地已选 cluster
    const originCluster = (overrideCluster && String(overrideCluster)) || getCurrentClusterId();
    const cluster = originCluster ? toUrlSafeBase64(originCluster) : '';
    // 未选择集群时，不插入 cluster 段，避免生成 /k8s/cluster//...
    if (!cluster) {
        return url;
    }
    const parts = url.split('/');
    parts.splice(2, 0, 'cluster', cluster);
    return parts.join('/');
}

// 解析路径,逐层获取值
// obj: 要解析的对象
// path: 路径，例如 'a.b.c'
// 返回值: 路径对应的值，如果路径不存在则返回 undefined
export function GetValueByPath<T = any>(obj: any, path: string, defaultValue?: T): T {
    if (!obj || typeof path !== 'string') return defaultValue as T;

    const keys = path.replace(/\[(\d+)\]/g, '.$1').split('.');

    function traverse(current: any, index: number): any {
        if (current == null) {
            return undefined;
        }

        const key = keys[index];

        if (Array.isArray(current)) {
            // 当前是数组，批量处理每个元素
            const results = current.map(item => traverse(item, index));
            return results.flat(); // 把结果拍平
        }

        if (index === keys.length - 1) {
            // 最后一个key
            return current?.[key];
        }

        return traverse(current[key], index + 1);
    }

    const result = traverse(obj, 0);

    // 返回默认值或者实际值
    if (result === undefined) {
        return defaultValue as T;
    }
    return result;
}

/**
 * 获取当前选中的集群ID（从 URL 哈希路径中解析）
 * 解析位置形如：`#/k/ClusterID/xxxx/yyy`，其中第二段为集群ID。
 * 使用 URL 安全 Base64 解码，不兼容旧的 `#/cluster/...` 路径。
 * @returns {string} 当前集群ID，未选择时返回空字符串
 */
export function getCurrentClusterId(): string {
    if (typeof window === 'undefined') return '';

    // 读取哈希并去除查询参数
    const rawHash = window.location.hash || '';
    const hashBody = rawHash.startsWith('#') ? rawHash.slice(1) : rawHash;
    const pathOnly = hashBody.split('?')[0] || '';

    // 统一成以 '/' 开头的路径，便于分段
    const normPath = pathOnly.startsWith('/') ? pathOnly : `/${pathOnly}`;
    const parts = normPath.split('/');
    const idx = parts.indexOf('k');

    if (idx >= 0 && parts.length > idx + 1 && parts[idx + 1]) {
        const encoded = parts[idx + 1];
        const decoded = fromUrlSafeBase64(encoded);
        // 严格按 Base64 解码，失败则视为未选择
        return decoded || '';
    }
    return '';
}

export function getCurrentClusterIdInBase64(): string {
    return  getCurrentClusterId() ? toUrlSafeBase64(getCurrentClusterId()) : '';
}

/**
 * 设置当前选中的集群ID（写入到 URL 哈希路径）
 * 目标位置：`#/k/ClusterID/xxxx/yyy`。若已有 k 段则替换其后 ID；
 * 若不存在，则在现有哈希路径前插入 `k/ClusterID`，保留剩余路径与查询参数。
 * @param {string} clusterId 要设置的集群ID
 */
export function setCurrentClusterId(clusterId: string): void {
    if (typeof window === 'undefined' || !clusterId) return;

    const encoded = toUrlSafeBase64(clusterId);
    const rawHash = window.location.hash || '';
    const hashBody = rawHash.startsWith('#') ? rawHash.slice(1) : rawHash;

    const hasQuery = hashBody.includes('?');
    const queryPart = hasQuery ? hashBody.slice(hashBody.indexOf('?')) : '';
    let pathOnly = (hasQuery ? hashBody.slice(0, hashBody.indexOf('?')) : hashBody) || '';

    // 统一为以 '/' 开头的路径
    pathOnly = pathOnly.startsWith('/') ? pathOnly : `/${pathOnly}`;
    // 去掉前导 '/'
    const segs = pathOnly.replace(/^\/+/,'').split('/').filter(s => s.length > 0);
    const idx = segs.indexOf('k');
    if (idx >= 0) {
        // 如果存在 k 段，移除该段以及紧随其后的 ID 段（若存在）
        segs.splice(idx, (segs.length > idx + 1) ? 2 : 1);
    }
    // 始终将 k/encoded 放到最前面
    const newSegs = ['k', encoded, ...segs];
    const newPath = '/' + newSegs.join('/');

    window.location.hash = `#${newPath}${queryPart}`;
    console.info('已切换到指定集群，更新哈希路径');
}

/**
 * 获取当前选中的命名空间（按集群维度隔离）
 * - 从 localStorage 读取 `selectedNS_${clusterId}`
 * - 若未设置或无法读取则返回空字符串
 * @param {string} [overrideClusterId] 可选，指定集群ID，默认读取当前URL中的集群ID
 * @returns {string} 当前选中的命名空间
 */
export function getSelectedNS(overrideClusterId?: string): string {
    const clusterId = (overrideClusterId && String(overrideClusterId)) || getCurrentClusterId();
    if (!clusterId) return '';
    const key = `selectedNS_${clusterId}`;
    try {
        const value = localStorage.getItem(key);
        return value || '';
    } catch (e) {
        console.warn('无法读取选中的命名空间:', e);
        return '';
    }
}

/**
 * 设置当前选中的命名空间（按集群维度隔离）
 * - 将命名空间写入 localStorage 的 `selectedNS_${clusterId}` 键
 * @param {string} ns 要设置的命名空间
 * @param {string} [overrideClusterId] 可选，指定集群ID，默认读取当前URL中的集群ID
 */
export function setSelectedNS(ns: string, overrideClusterId?: string): void {
    const clusterId = (overrideClusterId && String(overrideClusterId)) || getCurrentClusterId();
    if (!clusterId) return;
    const key = `selectedNS_${clusterId}`;
    try {
        localStorage.setItem(key, ns);
    } catch (e) {
        console.warn('无法保存选中的命名空间:', e);
    }
}



// 将方法暴露到window对象上，以便在脚本中使用
declare global {
    interface Window {
        getCurrentClusterId: typeof getCurrentClusterId;
        setCurrentClusterId: typeof setCurrentClusterId;
        getSelectedNS: typeof getSelectedNS;
        setSelectedNS: typeof setSelectedNS;
    }
}

if (typeof window !== 'undefined') {
    window.getCurrentClusterId = getCurrentClusterId;
    window.setCurrentClusterId = setCurrentClusterId;
    window.getSelectedNS = getSelectedNS;
    window.setSelectedNS = setSelectedNS;
}