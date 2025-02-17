import React from 'react';
import { Tree } from 'antd';
import { DownOutlined, FolderOpenFilled } from '@ant-design/icons';
import { EventDataNode } from 'antd/es/tree';

const { DirectoryTree } = Tree;

interface FileNode {
    name: string;
    type: string;
    permissions: string;
    owner: string;
    group: string;
    size: number;
    modTime: string;
    path: string;
    isDir: boolean;
    children?: FileNode[];
    isLeaf?: boolean;
    title: string;
    icon?: React.ReactNode | ((props: any) => React.ReactNode);
    disabled?: boolean;
    key: string;
}

interface FileTreeProps {
    treeData: FileNode[];
    selected: FileNode | undefined;
    onSelect: (selectedKeys: React.Key[], info: {
        event: "select";
        selected: boolean;
        node: EventDataNode<FileNode>;
        selectedNodes: FileNode[];
        nativeEvent: MouseEvent;
    }) => void;
    onExpand: (expandedKeys: React.Key[], info: {
        node: EventDataNode<FileNode>;
        expanded: boolean;
        nativeEvent: MouseEvent;
    }) => void;
    onRightClick: (info: { event: React.MouseEvent; node: EventDataNode<FileNode> }) => void;
}

const FileTree: React.FC<FileTreeProps> = ({
    treeData,
    selected,
    onSelect,
    onExpand,
    onRightClick
}) => {
    if (treeData.length === 0) {
        return (
            <div style={{
                textAlign: 'center',
                padding: '20px',
                color: '#999',
                fontSize: '14px'
            }}>
                <FolderOpenFilled style={{ fontSize: '32px', marginBottom: '8px', color: '#d9d9d9' }} />
                <div>暂无文件数据</div>
            </div>
        );
    }

    return (
        <DirectoryTree
            className='mt-4'
            treeData={treeData}
            showLine={true}
            checkStrictly={true}
            onSelect={onSelect}
            onExpand={onExpand}
            switcherIcon={<DownOutlined />}
            onRightClick={onRightClick}
            selectedKeys={selected ? [selected.key] : []}
        />
    );
};

export type { FileNode };
export default FileTree;