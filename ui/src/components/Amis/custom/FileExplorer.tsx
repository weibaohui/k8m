import React, { ReactNode, useEffect, useState } from 'react';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Button, message, Popconfirm, PopconfirmProps, Splitter, Tag, Tree } from 'antd';
import { FileFilled, FolderOpenFilled } from '@ant-design/icons';
import XTermComponent from './XTerm';
import { EventDataNode } from 'antd/es/tree';
import FormatBytes from './FormatBytes';
import formatLsShortDate from './FormatLsShortDate';
import { saveAs } from 'file-saver';
import { replacePlaceholders } from '@/utils/utils';

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
    icon?: ReactNode | ((props: any) => ReactNode);
    disabled?: boolean;
}

interface FileExplorerProps {
    containerName: string,
    podName: string,
    namespace: string,
    data: Record<string, any>
}

const FileExplorerComponent = React.forwardRef<HTMLDivElement, FileExplorerProps>(
    ({ containerName, podName, namespace, data }, _) => {
        console.log(containerName, podName, namespace, data)
        podName = replacePlaceholders(podName, data);
        namespace = replacePlaceholders(namespace, data);
        containerName = replacePlaceholders(containerName, data);
        console.log(containerName, podName, namespace, data)


        const [treeData, setTreeData] = useState<FileNode[]>([]);
        const [selected, setSelected] = useState<FileNode>();

        const fetchData = async (path: string = '/', isDir: boolean): Promise<FileNode[]> => {
            try {
                const response = await fetcher({
                    url: `/k8s/file/list?path=${encodeURIComponent(path)}`,
                    method: 'post',
                    data: {
                        "containerName": containerName,
                        "podName": podName,
                        "namespace": namespace,
                        "isDir": isDir,
                        "path": path
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
                return result;
            } catch (error) {
                console.error('Failed to fetch file tree:', error);
                return [];
            }
        };

        useEffect(() => {
            const initializeTree = async () => {
                const rootData = await fetchData("/", true);
                setTreeData(rootData);
            };
            initializeTree();
        }, [containerName, podName, namespace]);

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
        const renderIcon = (node: any) => {
            if (node.isDir) {
                return <i className={`${node.isDir} mr-2`} style={{ color: '#666' }} />;
            }
            if (!node.isDir) {
                return <FileFilled style={{ color: '#666', marginRight: 8 }} />;
            }
            return <FolderOpenFilled style={{ color: '#4080FF', marginRight: 8 }} />;
        };



        const onExpand: (expandedKeys: React.Key[], info: {
            node: EventDataNode<FileNode>;
            expanded: boolean;
            nativeEvent: MouseEvent;
        }) => void = (_, info) => {
            if (info.expanded) {
                fetchData(info.node.path, true).then((children) => {
                    setTreeData((origin) => updateTreeData(origin, info.node.path, children));
                });
            }
        };
        const onSelect: (selectedKeys: React.Key[], info: {
            event: "select";
            selected: boolean;
            node: EventDataNode<FileNode>;
            selectedNodes: FileNode[];
            nativeEvent: MouseEvent;
        }) => void = (_, info) => {
            setSelected(info.node);
        };


        const dirTree = () => {
            return <DirectoryTree
                treeData={treeData}
                showLine={true}
                style={{ width: '30vh', maxWidth: '200px' }}
                checkStrictly={true}
                onSelect={onSelect}
                onExpand={onExpand}
                showIcon={true}

            />
        }
        const confirmDeleteFile: PopconfirmProps['onConfirm'] = async (e) => {
            const response = await fetcher({
                url: '/k8s/file/delete',
                method: 'post',
                data: {
                    "containerName": containerName,
                    "podName": podName,
                    "namespace": namespace,
                    "path": selected?.path
                }
            });
            message.success(response.data?.msg);
        };
        const downloadFile: PopconfirmProps['onConfirm'] = async (e) => {
            try {
                const response = await fetcher({
                    url: '/k8s/file/download',
                    method: 'post',
                    data: {
                        "containerName": containerName,
                        "podName": podName,
                        "namespace": namespace,
                        "path": selected?.path
                    },
                    //@ts-ignore
                    responseType: 'blob' // 设置响应类型为blob
                });

                if (response && response.data) {
                    // 使用file-saver保存文件
                    //@ts-ignore
                    saveAs(new Blob([response.data]), selected?.name || 'download');
                    message.success('正在下载文件...');
                } else {
                    message.error('下载失败，请重试');
                }
            } catch (e) {
                message.error('下载失败，请重试');
            }

        };
        const fileInfo = () => {
            if (!selected) { return null; }
            const size = FormatBytes(selected?.size || 0)
            const time = formatLsShortDate(selected?.modTime)
            return (
                <>
                    <div className='mt-10' style={{ marginTop: '8px', fontFamily: 'monospace', whiteSpace: 'nowrap', display: 'flex', gap: '8px', alignItems: 'center' }}>
                        <Tag color="geekblue">
                            {selected?.path || ''}
                        </Tag>
                        <Tag color="blue">
                            {selected?.permissions || '-'}
                        </Tag>
                        <Tag color="green">
                            {selected?.owner || 'root'}
                        </Tag>
                        <Tag color="orange">
                            {selected?.group || 'root'}
                        </Tag>
                        <Tag color="purple">
                            {size}
                        </Tag>
                        <Tag color="red">
                            {time}
                        </Tag>


                        <Popconfirm
                            title='请确认'
                            description={`是否确认删除文件：${selected?.path} ？`}
                            onConfirm={confirmDeleteFile}
                            okText="删除"
                            cancelText="否"
                        >
                            <Button color="danger" variant="solid" disabled={!(selected.type == 'file' || selected.isDir)}>删除</Button>
                        </Popconfirm>
                        <Button className='ml-2' color="primary" variant="solid"
                            onClick={downloadFile} disabled={selected.type != 'file'}
                        >下载</Button>

                    </div >
                </>
            );
        }
        return (

            <Splitter style={{ height: '100%', boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)' }}>
                <Splitter.Panel collapsible defaultSize='20%'>
                    <div style={{ padding: '8px' }}>
                        {dirTree()}
                    </div>
                </Splitter.Panel>
                <Splitter.Panel>
                    {fileInfo()}
                    <XTermComponent
                        url={`/k8s/pod/xterm/ns/${namespace}/pod_name/${podName}` + ''}
                        params={{
                            "container_name": containerName
                        }}
                        data={{}}
                        height='100%'
                    ></XTermComponent>

                </Splitter.Panel>
            </Splitter>

        );
    });

export default FileExplorerComponent;