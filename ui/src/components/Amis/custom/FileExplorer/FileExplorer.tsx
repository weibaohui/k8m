import React, { useEffect, useState } from 'react';
import { Modal, Splitter } from 'antd';
import XTermComponent from '@/components/Amis/custom/XTerm';
import { EventDataNode } from 'antd/es/tree';
import MonacoEditorWithForm from '@/components/Amis/custom/MonacoEditorWithForm';
import FileTree, { FileNode } from '@/components/Amis/custom/FileExplorer/components/FileTree';
import ContextMenu from '@/components/Amis/custom/FileExplorer/components/ContextMenu';
import ContainerSelector from '@/components/Amis/custom/FileExplorer/components/ContainerSelector';
import { FileOperations } from '@/components/Amis/custom/FileExplorer/components/FileOperations';

interface FileExplorerProps {
    data: Record<string, any>
    remove?: string //关闭界面后是否删除Pod
}

const FileExplorerComponent = React.forwardRef<HTMLDivElement, FileExplorerProps>(
    ({ data, remove }, _) => {
        const podName = data?.metadata?.name;
        const namespace = data?.metadata?.namespace;

        const [treeData, setTreeData] = useState<FileNode[]>([]);
        const [selected, setSelected] = useState<FileNode>();
        const [selectedContainer, setSelectedContainer] = useState('');

        const [contextMenu, setContextMenu] = useState<{
            visible: boolean;
            x: number;
            y: number;
            node: FileNode | null
        }>({
            visible: false,
            x: 0,
            y: 0,
            node: null,
        });

        const fileOperations = new FileOperations({
            selectedContainer,
            podName,
            namespace
        });

        const handleRightClick = ({ event, node }: { event: React.MouseEvent; node: EventDataNode<FileNode> }) => {
            event.preventDefault();
            setContextMenu({
                visible: true,
                x: event.clientX,
                y: event.clientY,
                node: node,
            });
            setSelected(node);
        };

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

        const handleMenuClick = async (e: any) => {
            if (!contextMenu.node) return;

            switch (e.key) {
                case 'refresh':
                    await fileOperations.handleRefresh(contextMenu.node, (children) => {
                        setTreeData((origin) => updateTreeData(origin, contextMenu.node!.path, children));
                    });
                    break;
                case 'copy':
                    await fileOperations.handleCopy(contextMenu.node);
                    break;
                case 'delete':
                    await fileOperations.handleDelete(contextMenu.node, () => {
                        setTreeData((origin) => {
                            const removeNode = (list: FileNode[]): FileNode[] => {
                                return list.filter(node => {
                                    if (node.path === contextMenu.node?.path) {
                                        return false;
                                    }
                                    if (node.children) {
                                        node.children = removeNode(node.children);
                                    }
                                    return true;
                                });
                            };
                            return removeNode(origin);
                        });
                    });
                    break;
                case 'edit':
                    await fileOperations.handleEditFile(contextMenu.node, (content, language) => {
                        Modal.info({
                            title: '编辑' + contextMenu.node!.path + ' （ESC 关闭）',
                            width: '80%',
                            content: (
                                <div style={{
                                    border: '1px solid #e5e6eb',
                                    borderRadius: '4px'
                                }}>
                                    <MonacoEditorWithForm
                                        text={content}
                                        componentId="fileContext"
                                        saveApi="/k8s/file/save"
                                        data={{
                                            params: {
                                                containerName: selectedContainer,
                                                podName: podName,
                                                namespace: namespace,
                                                path: contextMenu.node!.path || '',
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
                    });
                    break;
                case 'download':
                    await fileOperations.downloadFile(contextMenu.node);
                    break;
                case 'downloadZip':
                    await fileOperations.downloadFile(contextMenu.node, 'tar');
                    break;
                case 'upload':
                    await fileOperations.handleUpload(contextMenu.node, async () => {
                        await fileOperations.handleRefresh(contextMenu.node!, (children) => {
                            setTreeData((origin) => updateTreeData(origin, contextMenu.node!.path, children));
                        });
                    });
                    break;
            }
            setContextMenu({ ...contextMenu, visible: false });
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

        const containerOptions = data.spec.containers.map((container: any) => ({
            name: container.name
        }));

        useEffect(() => {
            if (containerOptions.length > 0) {
                setSelectedContainer(containerOptions[0].name);
            }
        }, [data.spec.containers]);

        useEffect(() => {
            const initializeTree = async () => {
                const rootData = await fileOperations.fetchData("/", true);
                setTreeData(rootData);
            };
            if (selectedContainer) {
                initializeTree();
            }
        }, [selectedContainer, podName, namespace]);

        const onExpand = async (_: React.Key[], info: {
            node: EventDataNode<FileNode>;
            expanded: boolean;
            nativeEvent: MouseEvent;
        }) => {
            if (info.expanded) {
                const children = await fileOperations.fetchData(info.node.path, true);
                setTreeData((origin) => updateTreeData(origin, info.node.path, children));
            }
        };

        const onSelect = (_: React.Key[], info: {
            event: "select";
            selected: boolean;
            node: EventDataNode<FileNode>;
            selectedNodes: FileNode[];
            nativeEvent: MouseEvent;
        }) => {
            setSelected(info.node);
        };

        return (
            <>
                <Splitter style={{ height: '100%', boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)' }}>
                    <Splitter.Panel defaultSize='20%' collapsible={false}>
                        <div style={{ padding: '8px' }}>
                            <ContainerSelector
                                selectedContainer={selectedContainer}
                                containers={containerOptions}
                                onContainerChange={setSelectedContainer}
                            />
                            <div style={{ height: 'calc(100vh - 150px)', overflowY: 'auto' }}>
                                <FileTree
                                    treeData={treeData}
                                    selected={selected}
                                    onSelect={onSelect}
                                    onExpand={onExpand}
                                    onRightClick={handleRightClick}
                                />
                                <ContextMenu
                                    visible={contextMenu.visible}
                                    x={contextMenu.x}
                                    y={contextMenu.y}
                                    node={contextMenu.node}
                                    onMenuClick={handleMenuClick}
                                />
                            </div>
                        </div>
                    </Splitter.Panel>
                    <Splitter.Panel  >
                        {selectedContainer && (
                            <XTermComponent
                                url={`/k8s/pod/xterm/ns/${namespace}/pod_name/${podName}`}
                                params={{
                                    "container_name": selectedContainer,
                                    "remove": remove || "",
                                }}
                                data={{ data }}
                                height='calc(100vh - 100px)'
                                width='96%'
                            />
                        )}
                    </Splitter.Panel>
                </Splitter>
            </>
        );
    }
);

export default FileExplorerComponent;