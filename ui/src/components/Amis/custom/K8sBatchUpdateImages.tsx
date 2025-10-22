import React, { useState, useCallback, useMemo, useEffect } from 'react';
import {
    Table,
    Checkbox,
    Input,
    Button,
    message,
    Card,
    Space,
    Typography,
    Tag,
    Empty,
    Row,
    Col
} from 'antd';
import {
    ReloadOutlined,
    CloudUploadOutlined,
    ContainerOutlined,
    AppstoreOutlined,
    CheckCircleOutlined,
    InfoCircleOutlined,
    SettingOutlined,
    TagOutlined,
    SelectOutlined,
    UndoOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { Deployment } from '@/store/deployment';
import { Container } from '@/store/pod';
import { fetcher } from '@/components/Amis/fetcher';



const { Text } = Typography;

interface ContainerUpdateInfo {
    deploymentName: string;
    namespace: string;
    containerName: string;
    currentImage: string;
    shouldUpdate: boolean;
    newImage: string;
    imageAddress: string;
    imageTag: string;
}

interface BatchUpdateImageRequest {
    deployments: Array<{
        name: string;
        namespace: string;
        containers: Array<{
            name: string;
            image: string;
        }>;
    }>;
}

interface K8sBatchUpdateImagesProps {
    selectedDeployments?: Deployment[];
    data?: {
        selectedItems?: Deployment[];
    };
}

// 解析镜像地址和标签的工具函数
const parseImageAddressAndTag = (image: string): { address: string; tag: string } => {
    if (!image) {
        return { address: '', tag: '' };
    }

    const lastColonIndex = image.lastIndexOf(':');

    // 如果没有冒号，整个字符串都是地址，标签为空
    if (lastColonIndex === -1) {
        return { address: image, tag: '' };
    }

    // 检查冒号后面的部分是否包含斜杠，如果包含则可能是端口号而不是标签
    const potentialTag = image.substring(lastColonIndex + 1);
    if (potentialTag.includes('/')) {
        return { address: image, tag: '' };
    }

    // 拆分地址和标签
    const address = image.substring(0, lastColonIndex);
    const tag = potentialTag;

    return { address, tag };
};

// 合并镜像地址和标签的工具函数
const combineImageAddressAndTag = (address: string, tag: string): string => {
    if (!address) {
        return '';
    }

    if (!tag) {
        return address;
    }

    return `${address}:${tag}`;
};

const K8sBatchUpdateImages: React.FC<K8sBatchUpdateImagesProps> = ({ selectedDeployments, data }) => {
    const [containerUpdates, setContainerUpdates] = useState<Record<string, ContainerUpdateInfo>>({});
    const [loading, setLoading] = useState(false);
    const [batchTagValue, setBatchTagValue] = useState<string>('');

    // Get deployments from either prop or data object
    const deployments = useMemo(() => {
        return selectedDeployments || data?.selectedItems || [];
    }, [selectedDeployments, data]);

    // 解析容器信息
    const containerInfos = useMemo(() => {
        const infos: ContainerUpdateInfo[] = [];

        if (!deployments || !Array.isArray(deployments)) {
            return infos;
        }

        deployments.forEach(deployment => {
            if (!deployment || !deployment.spec) {
                return;
            }

            const containers: Container[] = deployment.spec?.template?.spec?.containers || [];
            containers.forEach(container => {
                if (!container) {
                    return;
                }

                const key = `${deployment.metadata?.namespace || 'default'}-${deployment.metadata?.name || 'unknown'}-${container.name || 'unknown'}`;
                infos.push({
                    deploymentName: deployment.metadata?.name || 'Unknown',
                    namespace: deployment.metadata?.namespace || 'default',
                    containerName: container.name || 'Unknown',
                    currentImage: container.image || 'Unknown',
                    shouldUpdate: false,
                    newImage: '',
                    imageAddress: '',
                    imageTag: ''
                });
            });
        });

        return infos;
    }, [deployments]);

    // 初始化容器更新状态
    useEffect(() => {
        const updates: Record<string, ContainerUpdateInfo> = {};
        containerInfos.forEach(info => {
            const key = `${info.namespace}-${info.deploymentName}-${info.containerName}`;
            updates[key] = { ...info };
        });
        setContainerUpdates(updates);
    }, [containerInfos]);

    // 统计信息
    const stats = useMemo(() => {
        const totalContainers = Object.keys(containerUpdates).length;
        const selectedForUpdate = Object.values(containerUpdates).filter(c => c.shouldUpdate).length;
        const deploymentCount = deployments.length;

        return {
            deploymentCount,
            totalContainers,
            selectedForUpdate
        };
    }, [containerUpdates, deployments]);

    // 更新容器信息
    const updateContainerInfo = useCallback((key: string, field: keyof ContainerUpdateInfo, value: any) => {
        setContainerUpdates(prev => {
            const currentUpdate = prev[key];

            // 如果是切换shouldUpdate状态，需要特殊处理
            if (field === 'shouldUpdate' && value && !currentUpdate?.shouldUpdate) {
                // 当启用更新时，自动解析当前镜像的地址和标签
                const container = containerInfos.find(c =>
                    `${c.namespace}-${c.deploymentName}-${c.containerName}` === key
                );
                if (container) {
                    const { address, tag } = parseImageAddressAndTag(container.currentImage);
                    return {
                        ...prev,
                        [key]: {
                            ...currentUpdate,
                            shouldUpdate: value,
                            imageAddress: address,
                            imageTag: tag
                        }
                    };
                }
            }

            const updated = { ...prev };
            updated[key] = {
                ...updated[key],
                [field]: value
            };

            // 当更新imageAddress或imageTag时，同时更新newImage
            if (field === 'imageAddress' || field === 'imageTag') {
                const container = updated[key];
                const address = field === 'imageAddress' ? value : container.imageAddress;
                const tag = field === 'imageTag' ? value : container.imageTag;
                updated[key].newImage = combineImageAddressAndTag(address, tag);
            }

            return updated;
        });
    }, [containerInfos]);

    // 批量更新处理
    const handleBatchUpdate = useCallback(async () => {
        // 过滤出需要更新的容器
        const containersToUpdate = containerInfos.filter(container => {
            const key = `${container.namespace}-${container.deploymentName}-${container.containerName}`;
            const updateInfo = containerUpdates[key];
            return updateInfo?.shouldUpdate && updateInfo?.imageAddress?.trim();
        });

        if (containersToUpdate.length === 0) {
            message.warning('请选择要更新的容器并输入镜像地址');
            return;
        }

        // 验证镜像地址格式
        const invalidContainers = containersToUpdate.filter(container => {
            const key = `${container.namespace}-${container.deploymentName}-${container.containerName}`;
            const updateInfo = containerUpdates[key];
            const imageAddress = updateInfo?.imageAddress?.trim();
            return !imageAddress || imageAddress.length === 0;
        });

        if (invalidContainers.length > 0) {
            message.error('请为所有选中的容器输入有效的镜像地址');
            return;
        }

        const selectedContainers = Object.values(containerUpdates).filter(c => c.shouldUpdate);

        setLoading(true);

        try {
            // 按 deployment 分组
            const deploymentGroups = selectedContainers.reduce((groups, container) => {
                const key = `${container.namespace}-${container.deploymentName}`;
                if (!groups[key]) {
                    groups[key] = {
                        name: container.deploymentName,
                        namespace: container.namespace,
                        containers: []
                    };
                }

                // 合并镜像地址和标签为完整镜像名
                const newImage = combineImageAddressAndTag(
                    container.imageAddress || '',
                    container.imageTag || ''
                );

                groups[key].containers.push({
                    name: container.containerName,
                    image: newImage
                });
                return groups;
            }, {} as Record<string, any>);

            const batchRequest: BatchUpdateImageRequest = {
                deployments: Object.values(deploymentGroups)
            };

            console.log('批量更新请求:', batchRequest);

            // 这里应该调用实际的API
            // await api.batchUpdateImages(batchRequest);

            console.log('批量更新请求:', JSON.stringify(batchRequest));
            // 调用后端API
            const response = await fetcher({
                url: '/k8s/deployment/batch_update_images',
                method: 'post',
                data: batchRequest
            });
            message.success(`成功更新 ${containersToUpdate.length} 个容器的镜像`);

            // 重置选择状态
            setContainerUpdates(prev => {
                const updated = { ...prev };
                Object.keys(updated).forEach(key => {
                    updated[key] = {
                        ...updated[key],
                        shouldUpdate: false,
                        newImage: '',
                        imageAddress: '',
                        imageTag: ''
                    };
                });
                return updated;
            });

        } catch (error) {
            console.error('批量更新失败:', error);
            message.error('批量更新失败，请重试');
        } finally {
            setLoading(false);
        }
    }, [containerUpdates, containerInfos]);

    // 重置所有选择
    const handleReset = useCallback(() => {
        setContainerUpdates(prev => {
            const updated = { ...prev };
            Object.keys(updated).forEach(key => {
                updated[key] = {
                    ...updated[key],
                    shouldUpdate: false,
                    newImage: '',
                    imageAddress: '',
                    imageTag: ''
                };
            });
            return updated;
        });
        setBatchTagValue('');
        message.info('已重置所有选择');
    }, []);

    // 批量设置标签
    const handleBatchSetTag = useCallback(() => {
        if (!batchTagValue.trim()) {
            message.warning('请输入要设置的标签值');
            return;
        }

        const selectedContainers = Object.entries(containerUpdates).filter(([_, update]) => update.shouldUpdate);

        if (selectedContainers.length === 0) {
            message.warning('请先选择要更新的容器');
            return;
        }

        setContainerUpdates(prev => {
            const updated = { ...prev };
            selectedContainers.forEach(([key, _]) => {
                if (updated[key]) {
                    updated[key] = {
                        ...updated[key],
                        imageTag: batchTagValue.trim()
                    };
                }
            });
            return updated;
        });

        message.success(`已为 ${selectedContainers.length} 个容器设置标签: ${batchTagValue.trim()}`);
    }, [batchTagValue, containerUpdates]);

    // 全选/取消全选处理函数
    const handleSelectAll = useCallback(() => {
        const allSelected = Object.values(containerUpdates).every(update => update.shouldUpdate);

        setContainerUpdates(prev => {
            const updated = { ...prev };
            Object.keys(updated).forEach(key => {
                const container = containerInfos.find(c =>
                    `${c.namespace}-${c.deploymentName}-${c.containerName}` === key
                );

                if (container) {
                    if (allSelected) {
                        // 如果全部已选中，则取消全选
                        updated[key] = {
                            ...updated[key],
                            shouldUpdate: false,
                            imageAddress: '',
                            imageTag: ''
                        };
                    } else {
                        // 如果未全选，则全选并自动解析镜像地址和标签
                        const { address, tag } = parseImageAddressAndTag(container.currentImage);
                        updated[key] = {
                            ...updated[key],
                            shouldUpdate: true,
                            imageAddress: address,
                            imageTag: tag
                        };
                    }
                }
            });
            return updated;
        });

        if (allSelected) {
            message.info('已取消全选');
        } else {
            message.success(`已全选 ${Object.keys(containerUpdates).length} 个容器`);
        }
    }, [containerUpdates, containerInfos]);

    // 判断是否全选状态
    const isAllSelected = useMemo(() => {
        const totalContainers = Object.keys(containerUpdates).length;
        if (totalContainers === 0) return false;
        return Object.values(containerUpdates).every(update => update.shouldUpdate);
    }, [containerUpdates]);

    // 计算每个Deployment的容器数量，用于rowSpan
    const deploymentContainerCounts = useMemo(() => {
        const counts: Record<string, number> = {};
        Object.values(containerUpdates).forEach(container => {
            const key = `${container.namespace}-${container.deploymentName}`;
            counts[key] = (counts[key] || 0) + 1;
        });
        return counts;
    }, [containerUpdates]);

    // 为表格数据添加分组信息
    const tableDataWithGrouping = useMemo(() => {
        const data = Object.values(containerUpdates);
        const groupedData: (ContainerUpdateInfo & {
            isFirstInGroup?: boolean;
            groupRowSpan?: number;
            deploymentKey?: string;
        })[] = [];

        // 按Deployment分组
        const deploymentGroups: Record<string, ContainerUpdateInfo[]> = {};
        data.forEach(container => {
            const key = `${container.namespace}-${container.deploymentName}`;
            if (!deploymentGroups[key]) {
                deploymentGroups[key] = [];
            }
            deploymentGroups[key].push(container);
        });

        // 为每组的第一个容器标记分组信息
        Object.entries(deploymentGroups).forEach(([deploymentKey, containers]) => {
            containers.forEach((container, index) => {
                const enhancedContainer = {
                    ...container,
                    isFirstInGroup: index === 0,
                    groupRowSpan: index === 0 ? containers.length : 0,
                    deploymentKey
                };
                groupedData.push(enhancedContainer);
            });
        });

        return groupedData;
    }, [containerUpdates]);

    // 表格列定义
    const columns: ColumnsType<ContainerUpdateInfo & {
        isFirstInGroup?: boolean;
        groupRowSpan?: number;
        deploymentKey?: string;
    }> = [
            {
                title: (
                    <Space>
                        <AppstoreOutlined />
                        <span>Deployment</span>
                    </Space>
                ),
                key: 'deployment',
                width: 200,
                render: (_, record) => {


                    const containerCount = deploymentContainerCounts[record.deploymentKey || ''] || 1;

                    return {
                        children: (
                            <div >
                                <div style={{
                                    fontWeight: 600,
                                    color: '#1890ff',
                                    fontSize: '14px',
                                    marginBottom: '4px'
                                }}>
                                    {record.deploymentName}
                                </div>
                                <div style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    marginBottom: '4px'
                                }}>
                                    <Text type="secondary" style={{ fontSize: '12px' }}>
                                        {record.namespace}
                                    </Text>
                                    <Tag color="blue" style={{ fontSize: '11px' }}>
                                        {containerCount} 个容器
                                    </Tag>
                                </div>
                            </div>
                        ),
                        props: {
                            rowSpan: record.groupRowSpan,
                        },
                    };
                },
            },
            {
                title: (
                    <Space>
                        <ContainerOutlined />
                        <span>容器名称</span>
                    </Space>
                ),
                dataIndex: 'containerName',
                key: 'containerName',
                width: 180,
                render: (containerName: string, record) => (
                    <div >
                        <ContainerOutlined style={{
                            color: '#52c41a',
                            fontSize: '14px'
                        }} />
                        <Text strong style={{
                            color: '#52c41a',
                            margin: 0,
                            padding: 0
                        }}>
                            {containerName}
                        </Text>
                    </div>
                ),
            },
            {
                title: '当前镜像',
                dataIndex: 'currentImage',
                key: 'currentImage',
                width: 300,
                render: (image) => (
                    <>
                        {image}
                    </>
                ),
            },
            {
                title: (
                    <Space>
                        <CheckCircleOutlined />
                        <span>更新</span>
                    </Space>
                ),
                key: 'shouldUpdate',
                width: 80,
                align: 'center',
                render: (_: any, record) => {
                    const key = `${record.namespace}-${record.deploymentName}-${record.containerName}`;
                    const containerUpdate = containerUpdates[key];

                    return (
                        <Checkbox
                            checked={containerUpdate?.shouldUpdate || false}
                            onChange={(e) => updateContainerInfo(key, 'shouldUpdate', e.target.checked)}
                        />
                    );
                },
            },
            {
                title: '镜像地址',
                key: 'imageAddress',
                width: 250,
                render: (_: any, record) => {
                    const key = `${record.namespace}-${record.deploymentName}-${record.containerName}`;
                    const containerUpdate = containerUpdates[key];

                    return (
                        <Input
                            placeholder="输入镜像地址"
                            value={containerUpdate?.imageAddress || ''}
                            disabled={!containerUpdate?.shouldUpdate}
                            onChange={(e) => updateContainerInfo(key, 'imageAddress', e.target.value)}
                            style={{
                                backgroundColor: containerUpdate?.shouldUpdate ? '#fff' : '#f5f5f5'
                            }}
                        />
                    );
                },
            },
            {
                title: '标签',
                key: 'imageTag',
                width: 150,
                render: (_: any, record) => {
                    const key = `${record.namespace}-${record.deploymentName}-${record.containerName}`;
                    const containerUpdate = containerUpdates[key];

                    return (
                        <Input
                            placeholder="输入标签"
                            value={containerUpdate?.imageTag || ''}
                            disabled={!containerUpdate?.shouldUpdate}
                            onChange={(e) => updateContainerInfo(key, 'imageTag', e.target.value)}
                            style={{
                                backgroundColor: containerUpdate?.shouldUpdate ? '#fff' : '#f5f5f5'
                            }}
                        />
                    );
                },
            },
        ];

    if (!deployments || deployments.length === 0) {
        return (
            <Card>
                <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={
                        <div>
                            <Text type="secondary">请先选择需要更新的 Deployment</Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: '12px' }}>
                                选择后将显示所有容器的镜像信息
                            </Text>
                        </div>
                    }
                />
            </Card>
        );
    }

    return (
        <div style={{ padding: '16px' }}>

            {/* 批量操作工具栏 */}
            <Card
                title={
                    <Space>
                        <SettingOutlined />
                        批量操作
                    </Space>
                }
                size="small"
                style={{ marginBottom: '16px' }}
            >
                <Row gutter={16} align="middle">
                    <Col span={6}>
                        <Typography.Text strong>批量设置标签:</Typography.Text>
                    </Col>
                    <Col span={8}>
                        <Input
                            placeholder="输入标签值 (如: v0.6)"
                            value={batchTagValue}
                            onChange={(e) => setBatchTagValue(e.target.value)}
                            onPressEnter={handleBatchSetTag}
                        />
                    </Col>
                    <Col span={4}>
                        <Button
                            type="primary"
                            ghost
                            onClick={handleBatchSetTag}
                            disabled={!batchTagValue.trim() || stats.selectedForUpdate === 0}
                        >
                            应用
                        </Button>
                    </Col>

                </Row>
            </Card>

            {/* 容器列表卡片 */}
            <Card
                title={
                    <Space>
                        <InfoCircleOutlined />
                        <span>容器镜像列表</span>
                        <Text type="secondary">({tableDataWithGrouping.length} 个容器)</Text>
                    </Space>
                }
                extra={
                    <Space>
                        <Button
                            type={isAllSelected ? "default" : "primary"}
                            ghost={!isAllSelected}
                            onClick={handleSelectAll}
                            icon={isAllSelected ? <UndoOutlined /> : <SelectOutlined />}
                            disabled={Object.keys(containerUpdates).length === 0}
                        >
                            {isAllSelected ? '取消全选' : '全选'}
                        </Button>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={handleReset}
                            disabled={loading}
                        >
                            重置
                        </Button>
                        <Button
                            type="primary"
                            icon={<CloudUploadOutlined />}
                            loading={loading}
                            onClick={handleBatchUpdate}
                            disabled={stats.selectedForUpdate === 0}
                        >
                            确定更新 {stats.selectedForUpdate > 0 && `(${stats.selectedForUpdate})`}
                        </Button>
                    </Space>
                }
                bodyStyle={{ padding: '0' }}
            >
                <Table
                    columns={columns}
                    dataSource={tableDataWithGrouping}
                    rowKey={(record) => `${record.namespace}-${record.deploymentName}-${record.containerName}`}
                    pagination={false}
                    scroll={{ x: 1000 }}
                    size="middle"
                    rowClassName={(record, index) => {
                        let className = record.shouldUpdate ? 'selected-row' : '';
                        if (record.isFirstInGroup) {
                            className += ' deployment-group-header';
                        } else {
                            className += ' container-row';
                        }
                        return className;
                    }}
                />
            </Card>


        </div>
    );
};

export default K8sBatchUpdateImages;
