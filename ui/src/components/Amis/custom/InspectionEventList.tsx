import React, { useEffect, useState } from 'react';
import { Button, List, Space, Tag, Typography, Alert, Card, Spin } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { fetcher } from '@/components/Amis/fetcher';
import WebSocketMarkdownViewerComponent from './WebSocketMarkdownViewer';
import { replacePlaceholders } from "@/utils/utils.ts";

const { Title } = Typography;

interface InspectionEventListComponentProps {
    record_id: string;
    data?: Record<string, any>;
}

const statusColorMap: Record<string, string> = {
    '正常': 'green',
    '失败': 'red',
    '警告': 'orange',
};

const InspectionEventListComponent: React.FC<InspectionEventListComponentProps> = (props) => {
    const { record_id: initialRecordId, data } = props;
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [eventData, setEventData] = useState<any[]>([]);
    const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});

    const record_id = replacePlaceholders(initialRecordId, data!) || "";

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
                    setEventData(res.data.data.rows);
                } else {
                    setEventData([]);
                    setError('未获取到事件数据');
                }
            })
            .catch((err: any) => {
                setError(err.message || '未知错误');
                setEventData([]);
            })
            .finally(() => setLoading(false));
    }, [record_id]);

    const toggleExplanation = (itemKey: string) => {
        setExpandedItems(prev => ({
            ...prev,
            [itemKey]: !prev[itemKey]
        }));
    };

    const filteredData = eventData.filter(item => item.event_status === '失败');

    return (
        <Card>
            <Title level={5} >异常事件明细（共 {filteredData.length} 条）</Title>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} showIcon />}
            <Spin spinning={loading} tip="加载中...">
                <List
                    dataSource={filteredData}
                    renderItem={item => {
                        const itemKey = `${item.kind}-${item.name}-${item.id}`;
                        return (
                            <List.Item style={{ padding: '24px 0', borderBottom: '1px solid #f0f0f0', display: 'block' }}>
                                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                                    <Space wrap>
                                        <Tag
                                            color={statusColorMap[item.event_status] || 'default'}>{item.event_status}</Tag>
                                        <Typography.Text strong>{item.kind}:</Typography.Text>
                                        <Typography.Text>{item.namespace}/{item.name}</Typography.Text>
                                        <Tag color="geekblue">{item.cluster}</Tag>
                                        <Tag
                                            color="default">{item.created_at ? dayjs(item.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}</Tag>
                                    </Space>
                                    <Alert
                                        style={{
                                            margin: '8px 0',
                                            background: item.event_status === '失败' ? '#fff1f0' : undefined
                                        }}
                                        message={<span style={{ fontWeight: 500 }}>{item.event_msg}</span>}
                                        type={item.event_status === '失败' ? 'error' : (item.event_status === '警告' ? 'warning' : 'success')}
                                        showIcon
                                        action={
                                            item.event_status !== '正常' && (
                                                <Button
                                                    icon={<QuestionCircleOutlined />}
                                                    onClick={() => toggleExplanation(itemKey)}
                                                    type="link"
                                                    style={{ padding: 0 }}
                                                >
                                                    AI解释
                                                </Button>
                                            )
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

export default InspectionEventListComponent;
