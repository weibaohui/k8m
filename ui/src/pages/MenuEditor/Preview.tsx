import React, {useEffect, useState} from 'react';
import {message, Tree} from 'antd';
import {MenuItem} from '@/types/menu';
import {Parser} from 'expr-eval'; // 引入 expr-eval
import {fetcher} from '@/components/Amis/fetcher';
import {useNavigate, NavigateFunction} from "react-router-dom";

interface PreviewProps {
    menuData: MenuItem[];
    onMenuClick?: (key: string) => void;
    navigate?: NavigateFunction; // 添加可选的 navigate 属性
}

// 定义用户角色接口
interface UserRoleResponse {
    role: string;  // 根据实际数据结构调整类型
    cluster: string;
}

interface CRDSupportedStatus {
    IsGatewayAPISupported: boolean;
    IsOpenKruiseSupported: boolean;
    IsIstioSupported: boolean;
}


const Preview: React.FC<PreviewProps> = ({menuData, navigate: propNavigate}) => {
    // 尝试使用传入的 navigate 或者 useNavigate hook
    let navigateFunc: NavigateFunction | undefined;

    try {
        // 尝试使用 React Router 的 useNavigate hook
        const routerNavigate = useNavigate();
        navigateFunc = propNavigate || routerNavigate;
    } catch (error) {
        // 如果不在 Router 上下文中，使用传入的 navigate 或者提供一个空函数
        navigateFunc = propNavigate || (path => {
            console.warn('Navigation attempted outside Router context to:', path);
            // 可以在这里添加备用导航逻辑，例如使用 window.location
        });
    }

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


    const handleClick = (key: string) => {
        console.log('点击菜单项:', key)
        const item = findMenuItem(menuData, key);
        if (item) {
            if (item.eventType === 'url' && item.url) {
                window.open(item.url, '_blank');
            } else if (item.eventType === 'custom' && item.customEvent) {
                try {
                    // 创建一个函数执行上下文
                    const context = {
                        loadJsonPage: (path: string) => {
                            if (navigateFunc) {
                                navigateFunc(path);
                            } else {
                                console.warn('Navigation attempted but navigate function is not available');
                            }
                        }
                    };

                    // 构建并执行自定义函数
                    const func = new Function(...Object.keys(context), `return ${item.customEvent}`);
                    const result = func(...Object.values(context));

                    // 如果是函数，执行它
                    if (typeof result === 'function') {
                        result();
                    }
                } catch (error) {
                    message.error('自定义事件执行错误:' + error);
                }
            }
        }
    };
    // 递归查找菜单项
    const findMenuItem = (data: MenuItem[], key: string): MenuItem | null => {
        for (const item of data) {
            if (item.key === key) return item;
            if (item.children) {
                const found = findMenuItem(item.children, key);
                if (found) return found;
            }
        }
        return null;
    };


    // 评估show属性
    const shouldShowItem = (item: MenuItem): boolean => {
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
                    user: {role: 'user'}, // 示例数据
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

                return Boolean(result);
            } catch (error) {
                console.error('评估显示表达式错误:', error);
                return false;
            }
        }

        return true;
    };

    const convertToTreeData = (data: MenuItem[]) => {
        // 过滤掉不显示的菜单项
        const filteredData = data.filter(item => shouldShowItem(item));

        return filteredData.map((item): {
            key: string;
            title: React.ReactNode;
            children?: ReturnType<typeof convertToTreeData>
        } => ({
            key: item.key as string,
            title: (
                <span onClick={() => handleClick(item.key)}>
                    {item.icon && <i className={`fa-solid ${item.icon}`} style={{marginRight: '4px'}}></i>}
                    {item.title}
                </span>
            ),
            children: item.children ? convertToTreeData(item.children) : undefined
        }));
    };

    return (
        <Tree
            treeData={convertToTreeData(menuData)}
            defaultExpandAll
            showLine
            blockNode
        />
    );
};

export default Preview;
