import React, { ReactNode, useEffect, useState } from 'react';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { message, Modal, PopconfirmProps, Select, Splitter, Tree, Menu, MenuProps } from 'antd';
import { CopyOutlined, DeleteOutlined, DownloadOutlined, EditOutlined, FileFilled, FileZipOutlined, FolderOpenFilled, UploadOutlined, ReloadOutlined } from '@ant-design/icons';
import XTermComponent from './XTerm';
import { EventDataNode } from 'antd/es/tree';
import MonacoEditorWithForm from './MonacoEditorWithForm';
type MenuItem = Required<MenuProps>['items'][number];

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
    key: string;
}

interface FileExplorerProps {
    data: Record<string, any>
}

const FileExplorerComponent = React.forwardRef<HTMLDivElement, FileExplorerProps>(
    ({ data }, _) => {
        const podName = data?.metadata?.name
        const namespace = data?.metadata?.namespace


        const [treeData, setTreeData] = useState<FileNode[]>([]);
        const [selected, setSelected] = useState<FileNode>();
        const [selectedContainer, setSelectedContainer] = React.useState('');

        const [contextMenu, setContextMenu] = useState<{ visible: boolean; x: number; y: number; node: FileNode | null }>({
            visible: false,
            x: 0,
            y: 0,
            node: null,
        });


        const items: MenuItem[] = [
            {
                label: '刷新',
                key: 'refresh',
                icon: <ReloadOutlined style={{ color: '#1890ff' }} />
            },
            {
                label: '复制路径',
                key: 'copy',
                icon: <CopyOutlined style={{ color: '#1890ff' }} />
            },
            {
                label: '删除',
                key: 'delete',
                icon: <DeleteOutlined style={{ color: '#ff4d4f' }} /> // 使用红色表示删除操作
            }, {
                label: '编辑',
                key: 'edit',
                disabled: !(contextMenu.node?.type == 'file'),
                icon: <EditOutlined style={{ color: '#52c41a' }} /> // 使用绿色表示编辑操作
            }, {
                label: '下载',
                key: 'download',
                disabled: !(contextMenu.node?.type == 'file'),
                icon: <DownloadOutlined style={{ color: '#722ed1' }} /> // 使用紫色表示下载操作
            }, {
                label: '压缩下载',
                key: 'downloadZip',
                disabled: !(contextMenu.node?.type == 'file'),
                icon: <FileZipOutlined style={{ color: '#faad14' }} /> // 使用金色表示压缩下载
            },
            {
                label: '上传',
                key: 'upload',
                disabled: !contextMenu.node?.isDir,
                icon: <UploadOutlined style={{ color: '#13c2c2' }} /> // 使用青色表示上传操作
            },
        ];

        const handleRightClick = ({ event, node }: { event: React.MouseEvent; node: EventDataNode<FileNode> }) => {
            event.preventDefault();
            setContextMenu({
                visible: true,
                x: event.clientX,
                y: event.clientY,
                node: node,
            });
            // 同时选中该节点
            setSelected(node);
        };
        const onClick: MenuProps['onClick'] = async (e) => {
            if (!contextMenu.node) return;
            switch (e.key) {
                case 'refresh':
                    if (contextMenu.node) {
                        fetchData(contextMenu.node.path, contextMenu.node.isDir).then((children) => {
                            if (contextMenu.node?.isDir) {
                                setTreeData((origin) => updateTreeData(origin, String(contextMenu.node?.path), children));
                            }
                        });
                        message.success('刷新成功');
                    }
                    break;
                case 'copy':
                    navigator.clipboard.writeText(contextMenu.node.path);
                    message.success('路径已复制到剪贴板');
                    break;
                case 'delete':
                    Modal.confirm({
                        title: '请确认',
                        content: `是否确认删除文件：${contextMenu.node.path} ？`,
                        okText: '删除',
                        cancelText: '取消',
                        onOk: confirmDeleteFile
                    });
                    break;
                case 'edit':
                    if (contextMenu.node.type === 'file') {
                        try {
                            const response = await fetcher({
                                url: `/k8s/file/show`,
                                method: 'post',
                                data: {
                                    "containerName": selectedContainer,
                                    "podName": podName,
                                    "namespace": namespace,
                                    "path": contextMenu.node.path
                                }
                            });
                            if (response.data?.status !== 0) {
                                message.error(response.data?.msg || '非文本文件，不可在线编辑。请下载编辑后上传。');
                                return;
                            }
                            //@ts-ignore
                            const fileContent = response.data?.data?.content || '';
                            let language = contextMenu.node.path?.split('.').pop() || 'plaintext';
                            switch (language) {
                                case 'yaml':
                                case 'yml':
                                    language = 'yaml';
                                    break;
                                case 'json':
                                    language = 'json';
                                    break;
                                case 'py':
                                    language = 'python';
                                    break;
                                default:
                                    language = 'shell';
                                    break;
                            }

                            Modal.info({
                                title: '编辑' + contextMenu.node.path + ' （ESC 关闭）',
                                width: '80%',
                                content: (
                                    <div style={{
                                        border: '1px solid #e5e6eb',
                                        borderRadius: '4px'
                                    }}>
                                        <MonacoEditorWithForm
                                            text={fileContent}
                                            componentId="fileContext"
                                            saveApi={`/k8s/file/save`}
                                            data={{
                                                params: {
                                                    containerName: selectedContainer,
                                                    podName: podName,
                                                    namespace: namespace,
                                                    path: contextMenu.node.path || '',
                                                }
                                            }}
                                            options={{
                                                language: language,
                                                wordWrap: "on",
                                                scrollbar: {
                                                    "vertical": "auto"
                                                }
                                            }}
                                        />
                                    </div>
                                ),
                                onOk() { },
                                okText: '取消',
                                okType: 'default',
                            });
                        } catch (error) {
                            message.error('获取文件内容失败');
                        }
                    } else {
                        message.error('只能编辑文件类型');
                    }
                    break;
                case 'download':
                    try {
                        const queryParams = new URLSearchParams({
                            containerName: selectedContainer,
                            podName: podName,
                            namespace: namespace,
                            path: contextMenu.node.path || "",
                            token: localStorage.getItem('token') || "",
                        }).toString();
                        const url = `/k8s/file/download?${queryParams}`;
                        const a = document.createElement('a');
                        a.href = url;
                        a.click();
                        message.success('文件正在下载...');
                    } catch (e) {
                        message.error('下载失败，请重试');
                    }
                    break;
                case 'downloadZip':
                    try {
                        const queryParams = new URLSearchParams({
                            containerName: selectedContainer,
                            podName: podName,
                            namespace: namespace,
                            path: contextMenu.node.path || "",
                            token: localStorage.getItem('token') || "",
                            type: 'tar'
                        }).toString();
                        const url = `/k8s/file/download?${queryParams}`;
                        const a = document.createElement('a');
                        a.href = url;
                        a.click();
                        message.success('文件正在下载...');
                    } catch (e) {
                        message.error('下载失败，请重试');
                    }
                    break;
                case 'upload':
                    if (!contextMenu.node.isDir) {
                        message.error('只能在目录下上传文件');
                        return;
                    }
                    const uploadInput = document.createElement('input');
                    uploadInput.type = 'file';
                    uploadInput.onchange = async (e) => {
                        const file = (e.target as HTMLInputElement).files?.[0];
                        if (!file) return;

                        const formData = new FormData();
                        formData.append('file', file);
                        formData.append('containerName', selectedContainer);
                        formData.append('podName', podName);
                        formData.append('namespace', namespace);
                        formData.append('isDir', String(contextMenu.node?.isDir));
                        formData.append('path', String(contextMenu.node?.path));
                        formData.append('fileName', file.name);

                        try {
                            const response = await fetch('/k8s/file/upload', {
                                method: 'POST',
                                headers: {
                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                },
                                body: formData
                            });
                            const result = await response.json();
                            if (result.data?.file?.status === 'done') {
                                message.success('上传成功');
                            } else {
                                message.error(result.data?.file?.error || '上传失败');
                            }
                        } catch (error) {
                            message.error('上传失败');
                        }
                    };
                    uploadInput.click();
                    break;
            }
            setContextMenu({ ...contextMenu, visible: false });
        };
        const renderContextMenu = () => {
            if (!contextMenu.visible || !contextMenu.node) return null;
            return (
                <Menu
                    style={{
                        position: 'absolute',
                        top: contextMenu.y,
                        left: contextMenu.x,
                        zIndex: 1000,
                        minWidth: '120px',
                        backgroundColor: '#ffffff',
                        boxShadow: '0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
                        border: '1px solid #f0f0f0',
                        borderRadius: '2px',
                    }}
                    onClick={onClick}
                    items={items}
                />
            );
        };

        useEffect(() => {
            const handleClickOutside = () => {
                if (contextMenu.visible) {
                    setContextMenu({ ...contextMenu, visible: false });
                }
            };

            window.addEventListener('click', handleClickOutside);
            return () => {
                window.removeEventListener('click', handleClickOutside);
            };
        }, [contextMenu]);
        const containerOptions = () => {
            const options = [];
            for (const container of data.spec.containers) {
                options.push({
                    label: container.name,
                    value: container.name
                });
            }
            return options;
        };
        // Initialize selected container
        useEffect(() => {
            const options = containerOptions();
            if (options.length > 0) {
                setSelectedContainer(options[0].value);
            }
        }, [data.spec.containers]);
        const fetchData = async (path: string = '/', isDir: boolean): Promise<FileNode[]> => {
            try {
                const response = await fetcher({
                    url: `/k8s/file/list?path=${encodeURIComponent(path)}`,
                    method: 'post',
                    data: {
                        "containerName": selectedContainer,
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
                    //key改成随机值
                    key: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
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
        }, [selectedContainer, podName, namespace]);


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
            // 当数据为空时显示骨架屏
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
            // 有数据时显示正常树
            return <DirectoryTree className='mt-4'
                treeData={treeData}
                showLine={true}
                checkStrictly={true}
                onSelect={onSelect}
                onExpand={onExpand}
                showIcon={true}
                onRightClick={handleRightClick}
                selectedKeys={selected ? [selected.key] : []}
            />
        }
        const confirmDeleteFile: PopconfirmProps['onConfirm'] = async () => {
            const response = await fetcher({
                url: '/k8s/file/delete',
                method: 'post',
                data: {
                    "containerName": selectedContainer,
                    "podName": podName,
                    "namespace": namespace,
                    "path": selected?.path
                }
            });
            message.success(response.data?.msg);
        };


        const handleContainerChange = (value: string) => {
            setSelectedContainer(value)
        };

        return (

            <>

                <Splitter style={{ height: '100%', boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)' }}>
                    <Splitter.Panel collapsible defaultSize='20%'>

                        <div style={{ padding: '8px' }}>
                            <Select
                                prefix='容器：'
                                value={selectedContainer}
                                onChange={handleContainerChange}
                                options={containerOptions()}
                            />
                            {dirTree()}
                            {renderContextMenu()}
                        </div>
                    </Splitter.Panel>
                    <Splitter.Panel>
                        {selectedContainer && (
                            <XTermComponent
                                url={`/k8s/pod/xterm/ns/${namespace}/pod_name/${podName}`}
                                params={{
                                    "container_name": selectedContainer
                                }}
                                data={{ data }}
                                height='calc(100vh - 100px)'
                                width='96%'
                            ></XTermComponent>
                        )}

                    </Splitter.Panel>
                </Splitter>
            </>


        );
    });

export default FileExplorerComponent;