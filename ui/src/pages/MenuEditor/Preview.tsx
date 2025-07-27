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

    const convertToTreeData = (data: MenuItem[]) => {
        return data.map(item => ({
            key: item.key,
            title: (
                <span onClick={() => handleClick(item.key)}>
                    {item.icon && <i className={`fa-solid ${item.icon}`} style={{marginRight: '4px'}}></i>}
                    {item.title}
                </span>
            ),
            children: item.children?.map(child => ({
                key: child.key,
                title: (
                    <span onClick={() => handleClick(child.key)}>
                        {child.icon && <i className={`fa-solid ${child.icon}`} style={{marginRight: '4px'}}></i>}
                        {child.title}
                    </span>
                ),
                children: child.children?.map(grandChild => ({
                    key: grandChild.key,
                    title: (
                        <span onClick={() => handleClick(grandChild.key)}>
                            {grandChild.icon &&
                                <i className={`fa-solid ${grandChild.icon}`} style={{marginRight: '4px'}}></i>}
                            {grandChild.title}
                        </span>
                    )
                }))
            }))
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
