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
 * 获取当前选中的集群ID，从URL路径中提取并解码
 * @returns {string} 当前集群ID，如果未选择则返回空字符串
 */
export function getCurrentClusterId(): string {
    if (typeof window !== 'undefined') {
        const currentPath = window.location.pathname;
        
        // 检查路径是否以 /cluster/ 开头
        if (currentPath.startsWith('/cluster/')) {
            const pathParts = currentPath.split('/');
            if (pathParts.length >= 3 && pathParts[2]) {
                // 提取base64编码的集群ID并解码
                const clusterIdBase64 = pathParts[2];
                try {
                    return fromUrlSafeBase64(clusterIdBase64);
                } catch (error) {
                    console.warn('无法解码集群ID:', clusterIdBase64, error);
                    return '';
                }
            }
        }
    }
    
    return  '';
}

/**
 * 设置当前选中的集群ID，并跳转到对应的集群页面，保持当前页面路径
 * @param {string} clusterId - 要设置的集群ID
 */
export function setCurrentClusterId(clusterId: string): void {
    localStorage.setItem('cluster', clusterId);
    
    // 将集群ID进行base64编码并跳转，保持当前页面路径
    if (typeof window !== 'undefined' && clusterId) {
        const clusterIdBase64 = toUrlSafeBase64(clusterId);
        const currentPath = window.location.pathname;
        const currentHash = window.location.hash;
        
        // 如果当前路径已经包含 /cluster/，则替换集群ID部分
        if (currentPath.startsWith('/cluster/')) {
            // 提取当前路径中集群ID后面的部分
            const pathParts = currentPath.split('/');
            if (pathParts.length > 2) {
                // 重新构建路径：/cluster/新集群ID/原有路径
                const remainingPath = pathParts.slice(3).join('/');
                const newPath = `/cluster/${clusterIdBase64}/${remainingPath}`;
                window.location.href = newPath + currentHash;
            } else {
                // 如果只有 /cluster/集群ID，则直接替换
                window.location.href = `/cluster/${clusterIdBase64}${currentHash}`;
            }
        } else {
            // 如果当前路径不包含集群信息，则跳转到集群首页
            window.location.href = `/cluster/${clusterIdBase64}${currentHash}`;
        }
        
        console.log("执行集群切换跳转，新集群ID:", clusterId);
    }
}

 

// 将方法暴露到window对象上，以便在脚本中使用
declare global {
    interface Window {
        getCurrentClusterId: typeof getCurrentClusterId;
        setCurrentClusterId: typeof setCurrentClusterId;
    }
}

if (typeof window !== 'undefined') {
    window.getCurrentClusterId = getCurrentClusterId;
    window.setCurrentClusterId = setCurrentClusterId;
}