import React, { useEffect, useState } from 'react';
import { Card, Table, Tag, Typography, Spin, Alert } from 'antd';
import dayjs from 'dayjs';
import { fetcher } from '@/components/Amis/fetcher';

const { Title } = Typography;

interface InspectionEventListProps {
    record_id: number | string;
}

const statusColorMap: Record<string, string> = {
    '正常': 'green',
    '失败': 'red',
    '警告': 'orange',
};

const columns = [
    {
        title: '状态',
        dataIndex: 'event_status',
        key: 'event_status',
        render: (status: string) => <Tag color={statusColorMap[status] || 'default'}>{status}</Tag>,
    },
    {
        title: '消息',
        dataIndex: 'event_msg',
        key: 'event_msg',
    },
    {
        title: '资源类型',
        dataIndex: 'kind',
        key: 'kind',
    },
    {
        title: '命名空间',
        dataIndex: 'namespace',
        key: 'namespace',
    },
    {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
    },
    {
        title: '脚本',
        dataIndex: 'script_name',
        key: 'script_name',
    },
    {
        title: '检查描述',
        dataIndex: 'check_desc',
        key: 'check_desc',
    },
    {
        title: '集群',
        dataIndex: 'cluster',
        key: 'cluster',
    },
    {
        title: '创建时间',
        dataIndex: 'created_at',
        key: 'created_at',
        render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss'),
    },
];

const InspectionEventList: React.FC<InspectionEventListProps> = ({ record_id }) => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [data, setData] = useState<any[]>([]);
    const [count, setCount] = useState<number>(0);

    useEffect(() => {
        if (!record_id) return;
        setLoading(true);
        setError(null);
        fetcher({
            url: `/admin/inspection/schedule/record/id/${record_id}/event/list`,
            method: 'get',
        })
            .then((res: any) => {
                if (res?.data?.rows) {
                    setData(res.data.rows);
                    setCount(res.data.count || res.data.rows.length);
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

    return (
        <Card style={{ marginTop: 24 }}>
            <Title level={5} style={{ marginBottom: 16 }}>事件明细（共 {count} 条）</Title>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} showIcon />}
            <Spin spinning={loading} tip="加载中...">
                <Table
                    columns={columns}
                    dataSource={data}
                    rowKey="id"
                    size="small"
                    pagination={{ pageSize: 20 }}
                    scroll={{ x: 'max-content' }}
                />
            </Spin>
        </Card>
    );
};

export default InspectionEventList;
