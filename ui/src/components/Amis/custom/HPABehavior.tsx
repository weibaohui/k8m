import { HPA } from '@/store/hpa';
import React from 'react';
import { Table, Tag, Tooltip } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

interface HPABehaviorProps {
    data: HPA;
}

const HPABehaviorComponent = React.forwardRef<HTMLSpanElement, HPABehaviorProps>(({ data }, ref) => {
    const getTypeColor = (type: string) => {
        switch (type) {
            case 'Pods':
                return 'blue';
            case 'Percent':
                return 'green';
            default:
                return 'default';
        }
    };

    const getSelectPolicyColor = (policy: string | undefined) => {
        switch (policy) {
            case 'Max':
                return 'orange';
            case 'Min':
                return 'cyan';
            case 'Disabled':
                return 'red';
            default:
                return 'default';
        }
    };

    const columns = [
        {
            title: '方向',
            dataIndex: 'direction',
            key: 'direction',
            render: (direction: string) => (
                <Tag color={direction === 'scaleUp' ? 'red' : 'green'}
                    icon={direction === 'scaleUp' ? <ArrowUpOutlined /> : <ArrowDownOutlined />}>
                    {direction === 'scaleUp' ? '扩容' : '缩容'}
                </Tag>
            ),
        },
        {
            title: '策略类型',
            dataIndex: 'type',
            key: 'type',
            render: (type: string) => (
                <Tag color={getTypeColor(type)}>
                    {type}
                </Tag>
            ),
        },
        {
            title: '数值',
            dataIndex: 'value',
            key: 'value',
            render: (value: number, record: any) => (
                <Tooltip title={`${value}${record.type === 'Percent' ? '%' : '个Pod'}`}>
                    <span style={{ color: '#1890ff' }}>
                        {value}{record.type === 'Percent' ? '%' : ''}
                    </span>
                </Tooltip>
            ),
        },
        {
            title: '周期(秒)',
            dataIndex: 'periodSeconds',
            key: 'periodSeconds',
            render: (seconds: number) => (
                <Tag color="processing">{seconds}s</Tag>
            ),
        },
        {
            title: '选择策略',
            dataIndex: 'selectPolicy',
            key: 'selectPolicy',
            render: (policy: string | undefined) => (
                <Tag color={getSelectPolicyColor(policy)}>
                    {policy || 'N/A'}
                </Tag>
            ),
        },
    ];

    const getBehaviorData = () => {
        const behavior = data?.spec?.behavior;
        if (!behavior) return [];

        const dataSource: any[] = [];

        if (behavior.scaleUp?.policies) {
            behavior.scaleUp.policies.forEach((policy, index) => {
                dataSource.push({
                    key: `up-${index}`,
                    direction: 'scaleUp',
                    type: policy.type,
                    value: policy.value,
                    periodSeconds: policy.periodSeconds,
                    selectPolicy: behavior.scaleUp?.selectPolicy,
                });
            });
        }

        if (behavior.scaleDown?.policies) {
            behavior.scaleDown.policies.forEach((policy, index) => {
                dataSource.push({
                    key: `down-${index}`,
                    direction: 'scaleDown',
                    type: policy.type,
                    value: policy.value,
                    periodSeconds: policy.periodSeconds,
                    selectPolicy: behavior.scaleDown?.selectPolicy,
                });
            });
        }

        return dataSource;
    };

    return (
        <span ref={ref}>
            <Table
                columns={columns}
                dataSource={getBehaviorData()}
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

export default HPABehaviorComponent;