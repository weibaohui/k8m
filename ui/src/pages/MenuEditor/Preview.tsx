import React from 'react';
import { message, Tree } from 'antd';
import { MenuItem } from '@/types/menu';
import { useNavigate, NavigateFunction } from "react-router-dom";
import { useUserRole } from '@/hooks/useUserRole';
import { useCRDStatus } from '@/hooks/useCRDStatus';
import { shouldShowMenuItem } from '@/utils/menuVisibility';

interface PreviewProps {
    menuData: MenuItem[];
    onMenuClick?: (key: string) => void;
    navigate?: NavigateFunction;
}


const Preview: React.FC<PreviewProps> = ({ menuData, navigate: propNavigate }) => {
    // 尝试使用传入的 navigate 或者 useNavigate hook
    let navigateFunc: NavigateFunction | undefined;

    try {
        const routerNavigate = useNavigate();
        navigateFunc = propNavigate || routerNavigate;
    } catch (error) {
        navigateFunc = propNavigate || (path => {
            console.warn('Navigation attempted outside Router context to:', path);
        });
    }

    // 使用自定义hooks
    const { userRole, menuData: _menuData } = useUserRole();
    const { isGatewayAPISupported, isOpenKruiseSupported, isIstioSupported } = useCRDStatus();

    // 创建菜单可见性上下文
    const visibilityContext = {
        userRole,
        _menuData,
        isGatewayAPISupported,
        isOpenKruiseSupported,
        isIstioSupported
    };


    const handleClick = (key: string) => {
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


    const convertToTreeData = (data: MenuItem[]) => {
        // 过滤掉不显示的菜单项
        const filteredData = data.filter(item => shouldShowMenuItem(item, {
            ...visibilityContext,
            groups: Array.isArray(_menuData) ? _menuData.map((item: { key: string }) => item.key) as string[] : []
        }));

        return filteredData.map((item): {
            key: string;
            title: React.ReactNode;
            children?: ReturnType<typeof convertToTreeData>
        } => ({
            key: item.key as string,
            title: (
                <span onClick={() => handleClick(item.key)}>
                    {item.icon && <i className={`fa-solid ${item.icon}`} style={{ marginRight: '4px' }}></i>}
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
