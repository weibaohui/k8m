import { HPA } from '@/store/hpa';
import React from 'react';
import { Table, Tag, Tooltip } from 'antd';
import { PercentageOutlined, DashboardOutlined, ApiOutlined, CloudOutlined, ContainerOutlined } from '@ant-design/icons';

interface HPAMetricsProps {
    data: HPA;
}

const HPAMetricsComponent = React.forwardRef<HTMLSpanElement, HPAMetricsProps>(({ data }, ref) => {
    const getTypeIcon = (type: string) => {
        switch (type) {
            case 'Resource':
                return <DashboardOutlined />;
            case 'Pods':
                return <ContainerOutlined />;
            case 'Object':
                return <ApiOutlined />;
            case 'External':
                return <CloudOutlined />;
            case 'ContainerResource':
                return <ContainerOutlined />;
            default:
                return null;
        }
    };

    const getTypeColor = (type: string) => {
        switch (type) {
            case 'Resource':
                return 'blue';
            case 'Pods':
                return 'green';
            case 'Object':
                return 'purple';
            case 'External':
                return 'orange';
            case 'ContainerResource':
                return 'cyan';
            default:
                return 'default';
        }
    };

    const columns = [
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            render: (type: string) => (
                <Tag color={getTypeColor(type)} icon={getTypeIcon(type)}>
                    {type}
                </Tag>
            ),
        },
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            render: (name: string) => (
                <Tooltip title={name}>
                    <span style={{ color: '#1890ff' }}>{name}</span>
                </Tooltip>
            ),
        },
        {
            title: '目标值',
            dataIndex: 'target',
            key: 'target',
            render: (target: any) => {
                if (target.type === 'Utilization' && target.averageUtilization) {
                    return (
                        <Tag color="processing">
                            {target.averageUtilization}%
                        </Tag>
                    );
                } else if (target.type === 'AverageValue' && target.averageValue) {
                    return <Tag color="success">{target.averageValue}</Tag>;
                } else if (target.type === 'Value' && target.value) {
                    return <Tag color="warning">{target.value}</Tag>;
                }
                return <Tag color="default">-</Tag>;
            },
        },
        {
            title: '当前值',
            dataIndex: 'current',
            key: 'current',
            render: (current: any) => {
                if (!current) return <Tag color="default">-</Tag>;
                if (current.averageUtilization) {
                    return (
                        <Tag color="processing">
                            {current.averageUtilization}%
                        </Tag>
                    );
                } else if (current.averageValue) {
                    return <Tag color="success">{current.averageValue}</Tag>;
                } else if (current.value) {
                    return <Tag color="warning">{current.value}</Tag>;
                }
                return <Tag color="default">-</Tag>;
            },
        },
    ];

    const getMetricsData = () => {
        if (!data?.spec?.metrics) return [];

        return data.spec.metrics.map((metric, index) => {
            let name = '';
            let target = null;
            let current = null;

            if (metric.type === 'Resource' && metric.resource) {
                name = metric.resource.name;
                target = metric.resource.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Resource' && m.resource?.name === name)?.resource?.current;
            } else if (metric.type === 'ContainerResource' && metric.containerResource) {
                name = `${metric.containerResource.container}/${metric.containerResource.name}`;
                target = metric.containerResource.target;
                current = data?.status?.currentMetrics?.find(m => 
                    m.type === 'ContainerResource' && 
                    m.containerResource?.container === metric.containerResource?.container && 
                    m.containerResource?.name === metric.containerResource?.name
                )?.containerResource?.current;
            } else if (metric.type === 'Pods' && metric.pods) {
                name = metric.pods.metric.name;
                target = metric.pods.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Pods' && m.pods?.metric.name === name)?.pods?.current;
            } else if (metric.type === 'External' && metric.external) {
                name = metric.external.metric.name;
                target = metric.external.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'External' && m.external?.metric.name === name)?.external?.current;
            } else if (metric.type === 'Object' && metric.object) {
                name = metric.object.metric.name;
                target = metric.object.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Object' && m.object?.metric.name === name)?.object?.current;
            }

            return {
                key: index,
                type: metric.type,
                name,
                target,
                current,
            };
        });
    };

    return (
        <span ref={ref}>
            <Table
                columns={columns}
                dataSource={getMetricsData()}
                pagination={false}
                size="small"
                style={{
                    backgroundColor: '#fff',
                    borderRadius: '8px',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                }}
            />
        </span>
    );
});

export default HPAMetricsComponent;
