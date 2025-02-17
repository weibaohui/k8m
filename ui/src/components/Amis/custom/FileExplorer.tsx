import React, { ReactNode, useEffect, useState } from 'react';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Button, message, Modal, Popconfirm, PopconfirmProps, Select, Splitter, Tag, Tree, Upload } from 'antd';
import { FileFilled, FolderOpenFilled } from '@ant-design/icons';
import XTermComponent from './XTerm';
import { EventDataNode } from 'antd/es/tree';
import FormatBytes from './FormatBytes';
import formatLsShortDate from './FormatLsShortDate';
import MonacoEditorWithForm from './MonacoEditorWithForm';

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

        const handleRightClick = ({ event, node }: { event: React.MouseEvent; node: EventDataNode<FileNode> }) => {
            event.preventDefault();
            setContextMenu({
                visible: true,
                x: event.clientX,
                y: event.clientY,
                node: node,
            });
        };

        const renderContextMenu = () => {
            if (!contextMenu.visible || !contextMenu.node) return null;
            return (
                <div
                    style={{
                        position: 'absolute',
                        top: contextMenu.y,
                        left: contextMenu.x,
                        backgroundColor: 'white',
                        border: '1px solid #ccc',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
                        zIndex: 1000,
                        padding: '8px',
                    }}
                >
                    <div onClick={() =>
                        //@ts-ignore
                        console.log('Copy Path:', contextMenu.node.path)}>Copy Path</div>
                    <div onClick={() =>
                        //@ts-ignore
                        console.log('Download:', contextMenu.node.path)}>Download</div>
                    <div onClick={() =>
                        //@ts-ignore
                        console.log('Upload to:', contextMenu.node.path)}>Upload</div>
                </div>
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
        const downloadFile: PopconfirmProps['onConfirm'] = async () => {
            try {
                // 构建查询参数
                const queryParams = new URLSearchParams({
                    containerName: selectedContainer,
                    podName: podName,
                    namespace: namespace,
                    path: selected?.path || "",
                    token: localStorage.getItem('token') || "",
                }).toString();

                // 构建完整的 GET 请求 URL
                const url = `/k8s/file/download?${queryParams}`;

                // 创建一个隐形的 <a> 标签
                const a = document.createElement('a');
                a.href = url; // 设置文件下载的 URL

                // 模拟用户点击 <a> 标签来触发文件下载
                a.click();

                message.success('文件正在下载...');
            } catch (e) {
                message.error('下载失败，请重试');
            }
        };
        const downloadTarFile: PopconfirmProps['onConfirm'] = async () => {
            try {
                // 构建查询参数
                const queryParams = new URLSearchParams({
                    containerName: selectedContainer,
                    podName: podName,
                    namespace: namespace,
                    path: selected?.path || "",
                    token: localStorage.getItem('token') || "",
                    type: 'tar'
                }).toString();

                // 构建完整的 GET 请求 URL
                const url = `/k8s/file/download?${queryParams}`;
                // 创建一个隐形的 <a> 标签
                const a = document.createElement('a');
                a.href = url; // 设置文件下载的 URL
                // 模拟用户点击 <a> 标签来触发文件下载
                a.click();
                message.success('文件正在下载...');
            } catch (e) {
                message.error('下载失败，请重试');
            }
        };
        const fileInfo = () => {
            if (!selected) {
                return null;
            }
            const size = FormatBytes(selected?.size || 0)
            const time = formatLsShortDate(selected?.modTime)
            return (
                <>
                    <div className='mt-10' style={{
                        marginTop: '8px',
                        fontFamily: 'monospace',
                        whiteSpace: 'nowrap',
                        display: 'flex',
                        gap: '8px',
                        alignItems: 'center'
                    }}>
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
                            <Button color="danger" variant="solid"
                                disabled={!(selected.type == 'file' || selected.isDir)}>删除</Button>
                        </Popconfirm>
                        <Button className='ml-2' color="primary" variant="solid"
                            onClick={async () => {
                                if (selected.type === 'file') {
                                    try {
                                        const response = await fetcher({
                                            url: `/k8s/file/show`,
                                            method: 'post',
                                            data: {
                                                "containerName": selectedContainer,
                                                "podName": podName,
                                                "namespace": namespace,
                                                "path": selected?.path
                                            }
                                        });
                                        if (response.data?.status !== 0) {
                                            //@ts-ignore
                                            message.error(response.data?.msg || '非文本文件，不可在线编辑。请下载编辑后上传。');
                                            return;
                                        }
                                        //@ts-ignore
                                        const fileContent = response.data?.data?.content || '';
                                        // const decodedString = atob(fileContent);
                                        const decodedString = fileContent;
                                        let language = selected?.path?.split('.').pop() || 'plaintext';
                                        //根据文件名后缀判断语言
                                        switch (language) {
                                            case 'yaml':
                                                language = 'yaml';
                                                break;
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
                                            title: '编辑' + selected?.path + ' （ESC 关闭）',
                                            width: '80%',
                                            content: (
                                                <div style={{
                                                    border: '1px solid #e5e6eb',
                                                    borderRadius: '4px'
                                                }}>
                                                    <MonacoEditorWithForm
                                                        text={decodedString}
                                                        componentId="fileContext"
                                                        saveApi={`/k8s/file/save`}
                                                        data={{
                                                            params: {
                                                                containerName: selectedContainer,
                                                                podName: podName,
                                                                namespace: namespace,
                                                                path: selected?.path || '',
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
                                            onOk() {
                                            },
                                            okText: '取消',
                                            okType: 'default',
                                        });
                                    } catch (error) {
                                        message.error('获取文件内容失败');
                                    }
                                }
                            }}
                            disabled={selected.type != 'file'}
                        >编辑</Button>
                        <Button className='ml-2' color="primary" variant="solid"
                            onClick={downloadFile} disabled={selected.type != 'file'}
                        >下载</Button>
                        <Button className='ml-2' color="primary" variant="solid"
                            onClick={downloadTarFile} disabled={selected.type != 'file'}
                        >压缩下载</Button>
                        <Upload
                            name='file'
                            action={`/k8s/file/upload`}
                            showUploadList={false}
                            headers={{
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }}
                            data={{
                                containerName: selectedContainer,
                                podName: podName,
                                namespace: namespace,
                                isDir: selected?.isDir || false,
                                path: selected?.path || '',
                                fileName: selected?.name || ''
                            }}
                            onChange={(info) => {

                                if (info.file.response?.data?.file?.status === 'done') {
                                    message.success('上传成功');
                                } else if (info.file.response?.data?.file?.status === 'error') {
                                    message.error(info.file.response?.data?.file?.error);
                                }
                            }}
                        >
                            <Button className='ml-2' color="primary" variant="solid"
                                disabled={!selected.isDir}>上传</Button>
                        </Upload>

                    </div>
                </>
            );
        }
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
                        <div style={{ padding: '8px' }}>
                            {fileInfo()}
                        </div>
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