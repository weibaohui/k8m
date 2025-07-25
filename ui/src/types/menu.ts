export interface MenuItem {
    key: string;
    title: string;
    icon?: string;
    url?: string;
    eventType?: 'url' | 'custom';
    order?: number;
    children?: MenuItem[];
}
