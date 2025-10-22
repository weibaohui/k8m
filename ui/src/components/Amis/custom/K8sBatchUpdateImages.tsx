import React, { useCallback, useState, useMemo } from 'react';
import { Button, message, Table, Checkbox, Input, Space, Typography, Divider } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';
import { Deployment, BatchUpdateImageRequest } from '@/store/deployment';
import { Container } from '@/store/pod';

const { Title, Text } = Typography;

interface ImageBatchUpdateProps {
    data: Record<string, any>;
}

interface ContainerUpdateInfo {
    deploymentName: string;
    deploymentNamespace: string;
    containerName: string;
    currentImage: string;
    newImage: string;
    shouldUpdate: boolean;
    key: string;
}

const ImageBatchUpdateComponent = React.forwardRef<HTMLDivElement, ImageBatchUpdateProps>(
    ({ data }, _) => {
        const selectedItems: Deployment[] = data.selectedItems || [];
        const [loading, setLoading] = useState(false);
        const [containerUpdates, setContainerUpdates] = useState<Record<string, ContainerUpdateInfo>>({});

        // 解析所有容器信息
        const containerList = useMemo(() => {
            const containers: ContainerUpdateInfo[] = [];
            
            selectedItems.forEach((deployment) => {
                deployment.spec.template.spec.containers.forEach((container: Container) => {
                    const key = `${deployment.metadata.namespace}-${deployment.metadata.name}-${container.name}`;
                    containers.push({
                        deploymentName: deployment.metadata.name,
                        deploymentNamespace: deployment.metadata.namespace,
                        containerName: container.name,
                        currentImage: container.image,
                        newImage: container.image,
                        shouldUpdate: false,
                        key
                    });
                });
            });
            
            return containers;
        }, [selectedItems]);

        // 初始化容器更新状态
        React.useEffect(() => {
            const initialUpdates: Record<string, ContainerUpdateInfo> = {};
            containerList.forEach(container => {
                initialUpdates[container.key] = { ...container };
            });
            setContainerUpdates(initialUpdates);
        }, [containerList]);

        // 处理更新选择
        const handleUpdateToggle = useCallback((key: string, checked: boolean) => {
            setContainerUpdates(prev => ({
                ...prev,
                [key]: {
                    ...prev[key],
                    shouldUpdate: checked
                }
            }));
        }, []);

        // 处理新镜像输入
        const handleImageChange = useCallback((key: string, newImage: string) => {
            setContainerUpdates(prev => ({
                ...prev,
                [key]: {
                    ...prev[key],
                    newImage
                }
            }));
        }, []);

        // 批量更新镜像
        const handleBatchUpdate = useCallback(async () => {
            const updatesToProcess = Object.values(containerUpdates).filter(update => update.shouldUpdate);
            
            if (updatesToProcess.length === 0) {
                message.warning('请至少选择一个容器进行更新');
                return;
            }

            // 验证新镜像不为空
            const invalidUpdates = updatesToProcess.filter(update => !update.newImage.trim());
            if (invalidUpdates.length > 0) {
                message.error('请为所有选中的容器填写新镜像地址');
                return;
            }

            setLoading(true);
            
            try {
                // 按 deployment 分组
                const deploymentGroups: Record<string, ContainerUpdateInfo[]> = {};
                updatesToProcess.forEach(update => {
                    const deploymentKey = `${update.deploymentNamespace}-${update.deploymentName}`;
                    if (!deploymentGroups[deploymentKey]) {
                        deploymentGroups[deploymentKey] = [];
                    }
                    deploymentGroups[deploymentKey].push(update);
                });

                // 构建批量更新请求
                const batchRequest: BatchUpdateImageRequest = {
                    deployments: Object.values(deploymentGroups).map(group => ({
                        name: group[0].deploymentName,
                        namespace: group[0].deploymentNamespace,
                        containers: group.map(update => ({
                            name: update.containerName,
                            image: update.newImage
                        }))
                    }))
                };

                console.log('批量更新请求:', batchRequest);

                // 这里应该调用实际的API
                // const response = await fetcher('/api/deploy/batch/update-images', {
                //     method: 'POST',
                //     body: JSON.stringify(batchRequest)
                // });

                // 模拟API调用
                await new Promise(resolve => setTimeout(resolve, 2000));
                
                message.success(`成功更新 ${updatesToProcess.length} 个容器的镜像`);
                
                // 重置选择状态
                setContainerUpdates(prev => {
                    const updated = { ...prev };
                    Object.keys(updated).forEach(key => {
                        updated[key].shouldUpdate = false;
                        updated[key].newImage = updated[key].currentImage;
                    });
                    return updated;
                });
                
            } catch (error) {
                console.error('批量更新失败:', error);
                message.error('批量更新失败，请重试');
            } finally {
                setLoading(false);
            }
        }, [containerUpdates]);

        // 表格列定义
        const columns = [
            {
                title: 'Deployment',
                dataIndex: 'deploymentName',
                key: 'deploymentName',
                render: (text: string, record: ContainerUpdateInfo) => (
                    <div>
                        <div><strong>{text}</strong></div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                            {record.deploymentNamespace}
                        </Text>
                    </div>
                ),
                width: 200,
            },
            {
                title: '容器名称',
                dataIndex: 'containerName',
                key: 'containerName',
                width: 150,
            },
            {
                title: '当前镜像',
                dataIndex: 'currentImage',
                key: 'currentImage',
                render: (text: string) => (
                    <Text code style={{ fontSize: '12px', wordBreak: 'break-all' }}>
                        {text}
                    </Text>
                ),
                width: 300,
            },
            {
                title: '更新',
                key: 'shouldUpdate',
                render: (_: any, record: ContainerUpdateInfo) => (
                    <Checkbox
                        checked={containerUpdates[record.key]?.shouldUpdate || false}
                        onChange={(e) => handleUpdateToggle(record.key, e.target.checked)}
                    />
                ),
                width: 80,
                align: 'center' as const,
            },
            {
                title: '新镜像',
                key: 'newImage',
                render: (_: any, record: ContainerUpdateInfo) => (
                    <Input
                        placeholder="输入新镜像地址"
                        value={containerUpdates[record.key]?.newImage || ''}
                        onChange={(e) => handleImageChange(record.key, e.target.value)}
                        disabled={!containerUpdates[record.key]?.shouldUpdate}
                        style={{ fontSize: '12px' }}
                    />
                ),
                width: 300,
            },
        ];

        const selectedCount = Object.values(containerUpdates).filter(update => update.shouldUpdate).length;

        return (
            <div style={{ padding: '16px' }}>
                <Title level={4}>批量更新容器镜像</Title>
                
                <div style={{ marginBottom: '16px' }}>
                    <Text>
                        选中的 Deployment 数量: <strong>{selectedItems.length}</strong>
                        {containerList.length > 0 && (
                            <>
                                {' | '}
                                容器总数: <strong>{containerList.length}</strong>
                                {' | '}
                                待更新容器: <strong style={{ color: selectedCount > 0 ? '#1890ff' : undefined }}>
                                    {selectedCount}
                                </strong>
                            </>
                        )}
                    </Text>
                </div>

                {containerList.length > 0 ? (
                    <>
                        <Table
                            columns={columns}
                            dataSource={containerList}
                            rowKey="key"
                            pagination={false}
                            size="small"
                            scroll={{ x: 1000 }}
                            style={{ marginBottom: '16px' }}
                        />
                        
                        <Divider />
                        
                        <Space>
                            <Button
                                type="primary"
                                onClick={handleBatchUpdate}
                                loading={loading}
                                disabled={selectedCount === 0}
                            >
                                批量更新镜像 ({selectedCount})
                            </Button>
                            <Button
                                onClick={() => {
                                    setContainerUpdates(prev => {
                                        const updated = { ...prev };
                                        Object.keys(updated).forEach(key => {
                                            updated[key].shouldUpdate = false;
                                            updated[key].newImage = updated[key].currentImage;
                                        });
                                        return updated;
                                    });
                                }}
                            >
                                重置
                            </Button>
                        </Space>
                    </>
                ) : (
                    <div style={{ textAlign: 'center', padding: '40px' }}>
                        <Text type="secondary">请先选择要更新的 Deployment</Text>
                    </div>
                )}
            </div>
        );
    }
);

export default ImageBatchUpdateComponent;
