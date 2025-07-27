import React from 'react';
import {Tree} from 'antd';
import {MenuItem} from '@/types/menu';

interface PreviewProps {
    menuData: MenuItem[];
    onMenuClick?: (key: string) => void;
}

const Preview: React.FC<PreviewProps> = ({menuData, onMenuClick}) => {
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
                    user: { role: 'admin' }, // 示例数据
                    // 可以根据实际需求添加更多上下文
                };
                
                // 构建并执行自定义函数
                const func = new Function(...Object.keys(context), `return ${item.show}`);
                const result = func(...Object.values(context));
                
                return Boolean(result);
            } catch (error) {
                console.error('显示表达式执行错误:', error);
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
