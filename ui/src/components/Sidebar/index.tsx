import useStore from '@/store/layout'
import { Menu } from 'antd';
import { useNavigate } from 'react-router-dom';
import { MenuItem } from '@/types/menu';
import { initialMenu } from '@/pages/MenuEditor/menuData';
import type { MenuProps } from 'antd';
import { useMemo, useEffect, useState } from 'react';
import { Parser } from 'expr-eval';
import { fetcher } from '@/components/Amis/fetcher';

type AntdMenuItem = Required<MenuProps>['items'][number];

// 定义用户角色接口
interface UserRoleResponse {
    role: string;
    cluster: string;
}

interface CRDSupportedStatus {
    IsGatewayAPISupported: boolean;
    IsOpenKruiseSupported: boolean;
    IsIstioSupported: boolean;
}

const Sidebar = () => {
    const { collapse } = useStore(state => state)
    const navigate = useNavigate();

    // 状态管理
    const [userRole, setUserRole] = useState<string>('');
    const [isGatewayAPISupported, setIsGatewayAPISupported] = useState<boolean>(false);
    const [isOpenKruiseSupported, setIsOpenKruiseSupported] = useState<boolean>(false);
    const [isIstioSupported, setIsIstioSupported] = useState<boolean>(false);

    useEffect(() => {
        const fetchUserRole = async () => {
            try {
                const response = await fetcher({
                    url: '/params/user/role',
                    method: 'get'
                });
                // 检查 response.data 是否存在，并确保其类型正确
                if (response.data && typeof response.data === 'object') {
                    const role = response.data.data as UserRoleResponse;
                    setUserRole(role.role);

                    const originCluster = localStorage.getItem('cluster') || '';
                    if (originCluster == "" && role.cluster != "") {
                        localStorage.setItem('cluster', role.cluster);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch user role:', error);
            }
        };

        const fetchCRDSupportedStatus = async () => {
            try {
                const response = await fetcher({
                    url: '/k8s/crd/status',
                    method: 'get'
                });
                if (response.data && typeof response.data === 'object') {
                    const status = response.data.data as CRDSupportedStatus;
                    setIsGatewayAPISupported(status.IsGatewayAPISupported);
                    setIsOpenKruiseSupported(status.IsOpenKruiseSupported);
                    setIsIstioSupported(status.IsIstioSupported);
                }
            } catch (error) {
                console.error('Failed to fetch Gateway API status:', error);
            }
        };


        fetchUserRole();
        fetchCRDSupportedStatus();
    }, []);

    // 评估show属性
    const shouldShowItem = (item: MenuItem): boolean => {
        console.log("itemshow", item.show)
        if (item.show === undefined || item.show === null) {
            return true;
        }

        if (typeof item.show === 'boolean') {
            return item.show;
        }

        if (typeof item.show === 'string') {
            try {
                // 创建一个函数执行上下文
                const context = {
                    // 这里可以添加一些上下文变量，例如用户信息等
                    user: { role: 'user' }, // 示例数据
                };

                // 创建 expr-eval 解析器实例
                const parser = new Parser();

                // 注入预定义的方法
                // 例如，添加一个名为 'contains' 的自定义函数
                // 用法：contains('admin', user.role)
                parser.functions.contains = function (str: string | string[], substr: string) {
                    if (typeof str !== 'string' || typeof substr !== 'string') {
                        return false;
                    }
                    return str.includes(substr);
                };
                //增加几个方法，判断是否支持gateway api、istio、kruise
                parser.functions.isGatewayAPISupported = function () {
                    return isGatewayAPISupported;
                };
                parser.functions.isIstioSupported = function () {
                    return isIstioSupported;
                };
                parser.functions.isOpenKruiseSupported = function () {
                    return isOpenKruiseSupported;
                };
                //userRole==platform_admin
                parser.functions.isPlatformAdmin = function () {
                    return userRole == 'platform_admin';
                };
                parser.functions.isUserHasRole = function (role: string) {
                    return userRole == role;
                };

                //增加几个方法，
                // 解析表达式
                const expr = parser.parse(item.show);

                // 评估表达式
                const result = expr.evaluate(context);
                console.log("expr", expr)
                console.log("result", result)
                return Boolean(result);
            } catch (error) {
                console.error('评估显示表达式错误:', error);
                return false;
            }
        }

        return true;
    };


    // 转换函数：将 initialMenu 格式转换为 Antd Menu 格式
    const convertMenuItems = (menuItems: MenuItem[]): AntdMenuItem[] => {
        return menuItems
            .filter(item => shouldShowItem(item)) // 第一层过滤：根据show属性过滤
            .sort((a, b) => (a.order || 0) - (b.order || 0))
            .map((item): AntdMenuItem => {
                const loadJsonPage = (path: string) => {
                    navigate(path);
                };

                // 解析 customEvent 中的路径
                const getPathFromCustomEvent = (customEvent?: string): string => {
                    if (!customEvent) return '';
                    const match = customEvent.match(/loadJsonPage\("([^"]+)"\)/);
                    return match ? match[1] : '';
                };

                const menuItem: AntdMenuItem = {
                    key: item.key,
                    label: item.title,
                    icon: item.icon ? <i className={item.icon}></i> : undefined,
                };

                // 如果有 customEvent，添加 onClick 处理
                if (item.customEvent) {
                    const path = getPathFromCustomEvent(item.customEvent);
                    if (path) {
                        (menuItem as any).onClick = () => loadJsonPage(path);
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

    // 使用 useMemo 缓存转换结果，依赖状态变化
    const menuItems = useMemo(() => convertMenuItems(initialMenu), [
        navigate,
        userRole,
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
