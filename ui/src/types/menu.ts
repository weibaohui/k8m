export interface MenuItem {
    key: string;
    title: string;
    icon?: string;
    url?: string;
    eventType?: 'url' | 'custom';
    customEvent?: string;
    order?: number;
    children?: MenuItem[];
}
