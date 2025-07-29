export interface MenuItem {
    key: string;
    title: string;
    icon?: string;
    url?: string;
    eventType?: 'url' | 'custom';
    customEvent?: string;
    order?: number;
    children?: MenuItem[];
    show?: string; // 修改这里，只保留字符串类型
}
