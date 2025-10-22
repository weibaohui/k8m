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
    Badge,
    Empty,
    Tooltip,
    Row,
    Col
} from 'antd';
import {
    ReloadOutlined,
    CloudUploadOutlined,
    ContainerOutlined,
    AppstoreOutlined,
    CheckCircleOutlined,
    InfoCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { Deployment } from '../../../store/deployment';
import { Container } from '../../../store/pod';

const { Title, Text } = Typography;

interface ContainerUpdateInfo {
    deploymentName: string;
    namespace: string;
    containerName: string;
    currentImage: string;
    shouldUpdate: boolean;
    newImage: string;
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

const K8sBatchUpdateImages: React.FC<K8sBatchUpdateImagesProps> = ({ selectedDeployments, data }) => {
    const [containerUpdates, setContainerUpdates] = useState<Record<string, ContainerUpdateInfo>>({});
    const [loading, setLoading] = useState(false);

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
                    newImage: ''
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
        setContainerUpdates(prev => ({
            ...prev,
            [key]: {
                ...prev[key],
                [field]: value
            }
        }));
    }, []);

    // 批量更新处理
    const handleBatchUpdate = useCallback(async () => {
        const selectedContainers = Object.values(containerUpdates).filter(c => c.shouldUpdate);

        if (selectedContainers.length === 0) {
            message.warning('请至少选择一个容器进行更新');
            return;
        }

        // 验证新镜像地址
        const invalidContainers = selectedContainers.filter(c => !c.newImage.trim());
        if (invalidContainers.length > 0) {
            message.error('请为所有选中的容器填写新的镜像地址');
            return;
        }

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
                groups[key].containers.push({
                    name: container.containerName,
                    image: container.newImage.trim()
                });
                return groups;
            }, {} as Record<string, any>);

            const batchRequest: BatchUpdateImageRequest = {
                deployments: Object.values(deploymentGroups)
            };

            console.log('批量更新请求:', batchRequest);

            // 这里应该调用实际的API
            // await api.batchUpdateImages(batchRequest);

            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 2000));

            message.success(`成功更新 ${selectedContainers.length} 个容器的镜像`);

            // 重置选择状态
            setContainerUpdates(prev => {
                const updated = { ...prev };
                Object.keys(updated).forEach(key => {
                    updated[key] = {
                        ...updated[key],
                        shouldUpdate: false,
                        newImage: ''
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
    }, [containerUpdates]);

    // 重置所有选择
    const handleReset = useCallback(() => {
        setContainerUpdates(prev => {
            const updated = { ...prev };
            Object.keys(updated).forEach(key => {
                updated[key] = {
                    ...updated[key],
                    shouldUpdate: false,
                    newImage: ''
                };
            });
            return updated;
        });
        message.info('已重置所有选择');
    }, []);

    // 表格列定义
    const columns: ColumnsType<ContainerUpdateInfo> = [
        {
            title: (
                <Space>
                    <AppstoreOutlined />
                    <span>Deployment</span>
                </Space>
            ),
            key: 'deployment',
            width: 200,
            render: (_, record) => (
                <div>
                    <div style={{ fontWeight: 500, color: '#1890ff' }}>
                        {record.deploymentName}
                    </div>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                        {record.namespace}
                    </Text>
                </div>
            ),
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
            width: 150,
            render: (name) => (
                <Tag color="blue" style={{ margin: 0 }}>
                    {name}
                </Tag>
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
            title: '新镜像',
            key: 'newImage',
            width: 300,
            render: (_: any, record) => {
                const key = `${record.namespace}-${record.deploymentName}-${record.containerName}`;
                const containerUpdate = containerUpdates[key];

                return (
                    <Input
                        placeholder="输入新的镜像地址"
                        value={containerUpdate?.newImage || ''}
                        disabled={!containerUpdate?.shouldUpdate}
                        onChange={(e) => updateContainerInfo(key, 'newImage', e.target.value)}
                        style={{
                            backgroundColor: containerUpdate?.shouldUpdate ? '#fff' : '#f5f5f5'
                        }}
                    />
                );
            },
        },
    ];

    // 表格数据
    const tableData = Object.values(containerUpdates);

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
            {/* 统计信息卡片 */}
            <Card
                style={{ marginBottom: '16px' }}
                bodyStyle={{ padding: '16px 24px' }}
            >
                <Row gutter={24} align="middle">
                    <Col flex="auto">
                        <Title level={4} style={{ margin: 0, color: '#1890ff' }}>
                            <CloudUploadOutlined style={{ marginRight: '8px' }} />
                            批量更新容器镜像
                        </Title>
                    </Col>
                    <Col>
                        <Space size="large">
                            <div style={{ textAlign: 'center' }}>
                                <Badge count={stats.deploymentCount} color="#52c41a">
                                    <div style={{ padding: '4px 8px' }}>
                                        <AppstoreOutlined style={{ fontSize: '16px', color: '#52c41a' }} />
                                    </div>
                                </Badge>
                                <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                                    Deployments
                                </div>
                            </div>
                            <div style={{ textAlign: 'center' }}>
                                <Badge count={stats.totalContainers} color="#1890ff">
                                    <div style={{ padding: '4px 8px' }}>
                                        <ContainerOutlined style={{ fontSize: '16px', color: '#1890ff' }} />
                                    </div>
                                </Badge>
                                <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                                    容器总数
                                </div>
                            </div>
                            <div style={{ textAlign: 'center' }}>
                                <Badge count={stats.selectedForUpdate} color="#f5222d">
                                    <div style={{ padding: '4px 8px' }}>
                                        <CheckCircleOutlined style={{ fontSize: '16px', color: '#f5222d' }} />
                                    </div>
                                </Badge>
                                <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                                    待更新
                                </div>
                            </div>
                        </Space>
                    </Col>
                </Row>
            </Card>

            {/* 容器列表卡片 */}
            <Card
                title={
                    <Space>
                        <InfoCircleOutlined />
                        <span>容器镜像列表</span>
                        <Text type="secondary">({tableData.length} 个容器)</Text>
                    </Space>
                }
                extra={
                    <Space>
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
                            批量更新 {stats.selectedForUpdate > 0 && `(${stats.selectedForUpdate})`}
                        </Button>
                    </Space>
                }
                bodyStyle={{ padding: '0' }}
            >
                <Table
                    columns={columns}
                    dataSource={tableData}
                    rowKey={(record) => `${record.namespace}-${record.deploymentName}-${record.containerName}`}
                    pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
                    }}
                    scroll={{ x: 1000 }}
                    size="middle"
                    rowClassName={(_, index) => index % 2 === 0 ? 'table-row-light' : 'table-row-dark'}
                />
            </Card>

            <style dangerouslySetInnerHTML={{
                __html: `
          .ant-table-tbody > tr.table-row-light > td {
            background-color: #fafafa;
          }
          .ant-table-tbody > tr.table-row-dark > td {
            background-color: #ffffff;
          }
          .ant-table-tbody > tr:hover > td {
            background-color: #e6f7ff !important;
          }
        `
            }} />
        </div>
    );
};

export default K8sBatchUpdateImages;
