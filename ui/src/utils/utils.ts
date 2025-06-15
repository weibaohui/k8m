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
    const base64 = btoa(str); // 标准 Base64
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, ''); // 转为 URL-safe
}

export function ProcessK8sUrlWithCluster(url: string): string {
    const originCluster = localStorage.getItem('cluster') || '';
    const cluster = originCluster ? toUrlSafeBase64(originCluster) : '';

    if (url.startsWith('/k8s')) {
        const parts = url.split('/');
        parts.splice(2, 0, 'cluster', cluster);
        return parts.join('/');
    }
    return url;
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