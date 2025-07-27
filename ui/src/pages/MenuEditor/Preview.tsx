import React, { useEffect, useState } from 'react';
import {Tree} from 'antd';
import {MenuItem} from '@/types/menu';
import { Parser } from 'expr-eval'; // 引入 expr-eval
import {fetcher} from '@/components/Amis/fetcher';

interface PreviewProps {
    menuData: MenuItem[];
    onMenuClick?: (key: string) => void;
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


const Preview: React.FC<PreviewProps> = ({menuData, onMenuClick}) => {
    
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
        onMenuClick?.(key);
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
                    user: { role: 'user' }, // 示例数据
                 };
                
                // 创建 expr-eval 解析器实例
                const parser = new Parser();
                
                // 注入预定义的方法
                // 例如，添加一个名为 'contains' 的自定义函数
                // 用法：contains('admin', user.role)
                parser.functions.contains = function(str: string | string[], substr: string) {
                    if (typeof str !== 'string' || typeof substr !== 'string') {
                        return false;
                    }
                    return str.includes(substr);
                };
                
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
        
        return filteredData.map((item): { key: string; title: React.ReactNode; children?: ReturnType<typeof convertToTreeData> } => ({
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
