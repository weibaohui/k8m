import useStore from '@/store/layout'
import { Menu } from 'antd';
import { useNavigate } from 'react-router-dom';
import { MenuItem } from '@/types/menu';
import { initialMenu } from '@/pages/MenuEditor/menuData';
import type { MenuProps } from 'antd';
import { useMemo } from 'react';
import { useUserRole } from '@/hooks/useUserRole';
import { useCRDStatus } from '@/hooks/useCRDStatus';
import { shouldShowMenuItem } from '@/utils/menuVisibility';
import { getCurrentClusterId, toUrlSafeBase64 } from '@/utils/utils';

type AntdMenuItem = Required<MenuProps>['items'][number];

const Sidebar = () => {
    const { collapse } = useStore(state => state)
    const navigate = useNavigate();

    // 使用自定义hooks
    const { userRole, menuData } = useUserRole();
    const { isGatewayAPISupported, isOpenKruiseSupported, isIstioSupported } = useCRDStatus();

    // 创建菜单可见性上下文
    const visibilityContext = {
        userRole,
        menuData,
        isGatewayAPISupported,
        isOpenKruiseSupported,
        isIstioSupported
    };


    // 转换函数：将 initialMenu 格式转换为 Antd Menu 格式
    const convertMenuItems = (menuItems: MenuItem[]): AntdMenuItem[] => {
        return menuItems
            .filter(item => shouldShowMenuItem(item, visibilityContext)) // 第一层过滤：根据show属性过滤
            .sort((a, b) => (a.order || 0) - (b.order || 0))
            .map((item): AntdMenuItem => {
                /**
                 * 跳转到指定页面路径
                 * 在跳转前，始终在路径前加上 `k/<clusterID>` 前缀（clusterID 为 URL 安全 Base64 编码）。
                 * 例如：传入 `/admin`，最终跳转为 `#/k/<clusterID>/admin`。
                 * 若未选择集群，则按原路径跳转。
                 */
                const loadJsonPage = (path: string) => {
                    const clusterId = getCurrentClusterId();
                    if (clusterId) {
                        const encoded = toUrlSafeBase64(clusterId);
                        const purePath = path.startsWith('/') ? path : `/${path}`;
                        navigate(`/k/${encoded}${purePath}`);
                    } else {
                        navigate(path);
                    }
                };

                /**
                 * 解析 customEvent 中的路径
                 * @param customEvent - 自定义事件字符串
                 * @returns 解析出的路径
                 */
                const getPathFromCustomEvent = (customEvent?: string): string => {
                    if (!customEvent) return '';
                    const match = customEvent.match(/loadJsonPage\("([^"]+)"\)/);
                    return match ? match[1] : '';
                };

                /**
                 * 解析 customEvent 中的 open 函数调用
                 * @param customEvent - 自定义事件字符串
                 * @returns 解析出的 URL
                 */
                const getUrlFromCustomEvent = (customEvent?: string): string => {
                    if (!customEvent) return '';
                    const match = customEvent.match(/open\("([^"]+)"\)/);
                    return match ? match[1] : '';
                };

                const menuItem: AntdMenuItem = {
                    key: item.key,
                    label: item.title,
                    icon: item.icon ? <i className={`fa-solid ${item.icon}`}></i> : undefined,
                };

                // 如果有 customEvent，添加 onClick 处理
                if (item.customEvent) {
                    const path = getPathFromCustomEvent(item.customEvent);
                    const url = getUrlFromCustomEvent(item.customEvent);

                    if (path) {
                        // 处理 loadJsonPage 类型的事件
                        (menuItem as any).onClick = () => loadJsonPage(path);
                    } else if (url) {
                        // 处理 open 类型的事件
                        (menuItem as any).onClick = () => window.open(url, '_blank');
                    }
                }

                // 如果有子菜单，递归转换（每个层级的子菜单都会执行过滤）
                if (item.children && item.children.length > 0) {
                    const filteredChildren = convertMenuItems(item.children);
                    // 只有当过滤后还有子菜单时才添加children属性
                    if (filteredChildren.length > 0) {
                        (menuItem as any).children = filteredChildren;
                    }
                }

                return menuItem;
            })
            .filter((menuItem): menuItem is AntdMenuItem => {
                // 第二层过滤：如果是父菜单但没有子菜单且没有点击事件，则过滤掉
                const hasChildren = (menuItem as any).children && (menuItem as any).children.length > 0;
                const hasClickEvent = (menuItem as any).onClick;

                // 如果有子菜单或有点击事件，则保留
                return hasChildren || hasClickEvent;
            });
    };

    // 解析菜单数据，优先使用 menuData，如果无效则使用 initialMenu
    const getMenuData = (): MenuItem[] => {
        if (menuData) {
            // 如果是数组，直接使用
            if (Array.isArray(menuData) && menuData.length > 0) {
                return menuData as MenuItem[];
            }

            // 如果是字符串，尝试解析
            if (typeof menuData === 'string') {
                const raw = menuData.trim();
                if (raw && raw !== '[object Object]') {
                    try {
                        const parsed = JSON.parse(raw);
                        if (Array.isArray(parsed) && parsed.length > 0) {
                            return parsed as MenuItem[];
                        }
                    } catch (error) {
                        console.warn('Failed to parse menuData string, falling back to initialMenu:', error);
                    }
                }
            }
        }
        return initialMenu;
    };

    // 使用 useMemo 缓存转换结果，依赖状态变化
    const menuItems = useMemo(() => {
        const menuDataToUse = getMenuData();
        return convertMenuItems(menuDataToUse);
    }, [
        navigate,
        userRole,
        menuData,
        isGatewayAPISupported,
        isOpenKruiseSupported,
        isIstioSupported
    ]);

    return (
        <div style={{ height: 'calc(100vh - 110px)', minWidth: 0, flex: "auto", overflow: 'auto' }}>

            <Menu
                mode="inline"
                inlineCollapsed={collapse}
                items={menuItems}
                style={{ height: '100%', borderRight: 0 }}
            >
            </Menu>
        </div>
    )
}

export default Sidebar
