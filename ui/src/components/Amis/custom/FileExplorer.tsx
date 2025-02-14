import React, { useEffect, useState } from 'react';
import { Tree } from '@arco-design/web-react';
import { IconFolder, IconFile } from '@arco-design/web-react/icon';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { NodeProps } from '@arco-design/web-react/es/Cascader';

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
}

interface FileExplorerProps {
    data: FileNode[];
}

const FileExplorerComponent = React.forwardRef<HTMLDivElement, FileExplorerProps>(({ data }, _) => {
    const [treeData, setTreeData] = useState<FileNode[]>([]);

    const fetchData = async (path: string = '/'): Promise<FileNode[]> => {
        try {
            const response = await fetcher({
                url: `/k8s/file/list?path=${encodeURIComponent(path)}`,
                method: 'post',
                data: {
                    "containerName": "k8m",
                    "podName": "k8m-d478997d5-x4wdp",
                    "namespace": "k8m"
                }
            });

            // @ts-ignore
            const rows = response.data?.data?.rows || [];
            const result = rows.map((item: any): FileNode => ({
                name: item.name || '',
                type: item.type || '',
                permissions: item.permissions || '',
                owner: item.owner || '',
                group: item.group || '',
                size: item.size || 0,
                modTime: item.modTime || '',
                path: item.path || '',
                isDir: item.isDir || false,
                isLeaf: !item.isDir,
                title: item.name,
            }));
            console.log(result)
            console.log(result)
            console.log(result)
            console.log(result)
            return result;
        } catch (error) {
            console.error('Failed to fetch file tree:', error);
            return [];
        }
    };

    useEffect(() => {
        const initializeTree = async () => {
            const rootData = await fetchData();
            setTreeData(rootData);
        };
        initializeTree();
    }, []);

    const updateTreeData = (list: FileNode[], key: string, children: FileNode[]): FileNode[] => {
        return list.map((node) => {
            if (node.path === key) {
                return { ...node, children };
            }
            if (node.children) {
                return { ...node, children: updateTreeData(node.children, key, children) };
            }
            return node;
        });
    };


    // @ts-ignore
    const renderIcon = (node: NodeProps) => {
        if (node.isDir) {
            return <i className={`${node.isDir} mr-2`} style={{ color: '#666' }} />;
        }
        if (!node.isDir) {
            return <IconFile style={{ color: '#666', marginRight: 8 }} />;
        }
        return <IconFolder style={{ color: '#4080FF', marginRight: 8 }} />;
    };

    const onLoadData = async (node: FileNode) => {
        const children = await fetchData(node.path);
        setTreeData((origin) => updateTreeData(origin, node.path, children));
    };

    return (
        <div style={{ padding: '8px' }}>
            <Tree
                treeData={treeData}
                showLine={true}
                size='mini'
                style={{ width: '30vh', maxWidth: '200px' }}
                blockNode
                autoExpandParent
                onSelect={(value, info) => {
                    console.log(value, info);
                }}
                renderExtra={renderIcon}
            />
        </div>
    );
});

export default FileExplorerComponent;