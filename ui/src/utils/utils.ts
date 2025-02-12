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

