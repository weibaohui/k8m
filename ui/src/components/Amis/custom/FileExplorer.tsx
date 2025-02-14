import React, { useState } from 'react';
import { Tree } from '@arco-design/web-react';
import { IconFolder, IconFile } from '@arco-design/web-react/icon';
const TreeData = [
    {
        title: 'Trunk 1',
        key: '0-0',
        children: [
            {
                title: 'Trunk 1-0',
                key: '0-0-0',
                children: [
                    {
                        title: 'leaf',
                        key: '0-0-0-0',
                    },
                    {
                        title: 'leaf',
                        key: '0-0-0-1',
                        children: [
                            {
                                title: 'leaf',
                                key: '0-0-0-1-0',
                            },
                        ],
                    },
                    {
                        title: 'leaf',
                        key: '0-0-0-2',
                    },
                ],
            },
            {
                title: 'Trunk 1-1',
                key: '0-0-1',
            },
            {
                title: 'Trunk 1-2',
                key: '0-0-2',
                children: [
                    {
                        title: 'leaf',
                        key: '0-0-2-0',
                    },
                    {
                        title: 'leaf',
                        key: '0-0-2-1',
                    },
                ],
            },
        ],
    },
    {
        title: 'Trunk 2',
        key: '0-1',
    },
    {
        title: 'Trunk 3',
        key: '0-2',
        children: [
            {
                title: 'Trunk 3-0',
                key: '0-2-0',
                children: [
                    {
                        title: 'leaf',
                        key: '0-2-0-0',
                    },
                    {
                        title: 'leaf',
                        key: '0-2-0-1',
                    },
                ],
            },
        ],
    },
];
interface FileNode {
    key: string;
    title: string;
    isLeaf?: boolean;
    children?: FileNode[];
    icon?: string;
}

interface FileExplorerProps {
    data: FileNode[];
}

const FileExplorerComponent = React.forwardRef<HTMLDivElement, FileExplorerProps>(({ data }, ref) => {
    const [treeData, setTreeData] = useState(TreeData);

    const renderIcon = (node: FileNode) => {
        // 如果节点有自定义图标，使用自定义图标
        if (node.icon) {
            return <i className={`${node.icon} mr-2`} style={{ color: '#666' }} />;
        }
        // 否则使用默认的文件夹或文件图标
        if (node.isLeaf) {
            return <IconFile style={{ color: '#666', marginRight: 8 }} />;
        }
        return <IconFolder style={{ color: '#4080FF', marginRight: 8 }} />;
    };

    return (
        <div ref={ref} style={{ padding: '8px' }}>
            <Tree
                treeData={treeData}
                showLine={true}
                size='mini'
                style={{
                    backgroundColor: '#1E1E1E',
                    color: '#CCCCCC'
                }}
                blockNode
                autoExpandParent
            />
        </div>
    );
});

export default FileExplorerComponent;