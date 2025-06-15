import React, { useEffect, useState } from 'react';
import { Button, List, Space, Tag, Typography, Alert, Card, Spin } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { fetcher } from '@/components/Amis/fetcher';
import WebSocketMarkdownViewerComponent from './WebSocketMarkdownViewer';

const { Title } = Typography;

interface InspectionEventListProps {
    record_id: number | string;
}

const statusColorMap: Record<string, string> = {
    '正常': 'green',
    '失败': 'red',
    '警告': 'orange',
};

const InspectionEventList: React.FC<InspectionEventListProps> = ({ record_id }) => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [data, setData] = useState<any[]>([]);
    const [count, setCount] = useState<number>(0);
    const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});

    useEffect(() => {
        if (!record_id) return;
        setLoading(true);
        setError(null);
        fetcher({
            url: `/admin/inspection/schedule/record/id/${record_id}/event/list`,
            method: 'get',
        })
            .then((res: any) => {
                if (res?.data?.data?.rows) {
                    setData(res.data.data.rows);
                    setCount(res.data.data.count || res.data.data.rows.length);
                } else {
                    setData([]);
                    setCount(0);
                    setError('未获取到事件数据');
                }
            })
            .catch((err: any) => {
                setError(err.message || '未知错误');
                setData([]);
                setCount(0);
            })
            .finally(() => setLoading(false));
    }, [record_id]);

    const toggleExplanation = (itemKey: string) => {
        setExpandedItems(prev => ({
            ...prev,
            [itemKey]: !prev[itemKey]
        }));
    };

    return (
        <Card style={{ marginTop: 24 }}>
            <Title level={5} style={{ marginBottom: 16 }}>事件明细（共 {count} 条）</Title>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} showIcon />}
            <Spin spinning={loading} tip="加载中...">
                <List
                    dataSource={data}
                    renderItem={item => {
                        const itemKey = `${item.kind}-${item.name}-${item.id}`;
                        return (
                            <List.Item style={{ padding: '24px 0', borderBottom: '1px solid #f0f0f0', display: 'block' }}>
                                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                                    <Space wrap>
                                        <Tag color={statusColorMap[item.event_status] || 'default'}>{item.event_status}</Tag>
                                        <Typography.Text strong>{item.kind}:</Typography.Text>
                                        <Typography.Text>{item.name}</Typography.Text>
                                        <Tag color="blue">{item.namespace}</Tag>
                                        <Tag color="purple">{item.script_name}</Tag>
                                        <Tag color="geekblue">{item.cluster}</Tag>
                                        <Tag color="default">{item.created_at ? dayjs(item.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}</Tag>
                                    </Space>
                                    <Alert
                                        style={{ margin: '8px 0', background: item.event_status === '失败' ? '#fff1f0' : undefined }}
                                        message={<span style={{ fontWeight: 500 }}>{item.event_msg}</span>}
                                        type={item.event_status === '失败' ? 'error' : (item.event_status === '警告' ? 'warning' : 'success')}
                                        showIcon
                                        action={
                                            <Button
                                                icon={<QuestionCircleOutlined />}
                                                onClick={() => toggleExplanation(itemKey)}
                                                type="link"
                                                style={{ padding: 0 }}
                                            >
                                                AI解释
                                            </Button>
                                        }
                                    />
                                    {expandedItems[itemKey] && (
                                        <div style={{ marginTop: 8, marginBottom: 8 }}>
                                            <WebSocketMarkdownViewerComponent
                                                url="/ai/chat/k8s_gpt/resource"
                                                params={{
                                                    kind: item.kind,
                                                    name: item.name,
                                                    data: item.event_msg,
                                                    field: item.check_desc
                                                }}
                                                data={{}}
                                            />
                                        </div>
                                    )}
                                </Space>
                            </List.Item>
                        );
                    }}
                />
            </Spin>
        </Card>
    );
};

export default InspectionEventList;
