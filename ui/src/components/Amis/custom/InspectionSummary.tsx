import React, { useState, useEffect } from 'react';
import { fetcher } from '@/components/Amis/fetcher';
import { Alert, Button, Card, DatePicker, Form, Spin, Table } from "antd";
import dayjs from 'dayjs';


interface InspectionSummaryComponentProps {
    id: string;
}

/**
 * InspectionSummaryComponent 组件
 * 使用antd控件，动态根据表单参数从API获取数据并展示amis风格页面，使用fetcher封装请求
 */
const InspectionSummaryComponent = React.forwardRef<HTMLDivElement, InspectionSummaryComponentProps>(({ id }, _) => {
    // 表单状态
    const [form] = Form.useForm();
    const [startTime, setStartTime] = useState(() => dayjs().startOf('day'));
    const [endTime, setEndTime] = useState(() => dayjs().add(1, 'day').startOf('day'));
    const [loading, setLoading] = useState(false);
    const [data, setData] = useState<any>({});
    const [error, setError] = useState<string | null>(null);
    if (id === undefined) {
        id = ""
    }
    // 查询API，使用fetcher
    const fetchSummary = (values?: any) => {
        setLoading(true);
        setError(null);
        const sTime = (values?.startTime || startTime).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
        const eTime = (values?.endTime || endTime).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
        const url = `/admin/inspection/schedule/id/${id}/summary/start_time/${encodeURIComponent(sTime)}/end_time/${encodeURIComponent(eTime)}`;
        fetcher({ url, method: 'post' })
            .then((response: any) => {
                if (response?.data?.data) {
                    setData(response.data.data);
                } else {
                    setData({});
                    setError('未获取到数据');
                }
            })
            .catch((err: any) => {
                setError(err.message || '未知错误');
                setData({});
            })
            .finally(() => {
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchSummary();
    }, []);

    const { total_runs, total_clusters, latest_run = {}, clusters = [] } = data || {};

    // antd表格列定义
    const latestRunColumns = [
        { title: '资源类型', dataIndex: 'kind', key: 'kind' },
        { title: '正常数', dataIndex: 'normal_count', key: 'normal_count' },
        { title: '异常数', dataIndex: 'error_count', key: 'error_count' }
    ];
    const clusterColumns = [
        { title: '资源类型', dataIndex: 'kind', key: 'kind' },
        { title: '总数', dataIndex: 'count', key: 'count' },
        { title: '异常数', dataIndex: 'error_count', key: 'error_count' }
    ];

    return (
        <div>
            {/* 查询表单 */}
            <Form
                form={form}
                layout="inline"
                initialValues={{ startTime, endTime }}
                onFinish={values => {
                    setStartTime(values.startTime);
                    setEndTime(values.endTime);
                    fetchSummary(values);
                }}
                style={{ marginBottom: 16 }}
            >
                <Form.Item label="起始时间" name="startTime" rules={[{ required: true, message: '请选择起始时间' }]}>
                    <DatePicker showTime format="YYYY-MM-DD HH:mm" value={startTime} />
                </Form.Item>
                <Form.Item label="结束时间" name="endTime" rules={[{ required: true, message: '请选择结束时间' }]}>
                    <DatePicker showTime format="YYYY-MM-DD HH:mm" value={endTime} />
                </Form.Item>
                <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading}>查询</Button>
                </Form.Item>
            </Form>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} />}
            <Spin spinning={loading} tip="加载中...">
                <Card style={{ marginBottom: 16 }}>
                    <b>总执行次数：</b> {total_runs} <b>总集群数：</b> {total_clusters}
                </Card>
                {latest_run && (
                    <Card title={<span>最后一次执行 <b>记录ID：</b>{latest_run.record_id}
                        <b>执行时间：</b>{latest_run.run_time}</span>} style={{ marginBottom: 16 }}>
                        <Table
                            columns={latestRunColumns}
                            dataSource={latest_run.kinds || []}
                            rowKey={(r: any) => r.kind}
                            pagination={false}
                            size="small"
                        />
                    </Card>
                )}
                <div style={{ marginBottom: 8 }}><b>各集群明细：</b></div>
                {clusters.map((cluster: any, idx: number) => (
                    <Card key={idx} title={<span>集群：{cluster.cluster} (执行{cluster.run_count}次)</span>}
                        style={{ marginBottom: 16 }}>
                        <Table
                            columns={clusterColumns}
                            dataSource={cluster.kinds || []}
                            rowKey={(r: any) => r.kind}
                            pagination={false}
                            size="small"
                        />
                    </Card>
                ))}
            </Spin>
        </div>
    );
});

export default InspectionSummaryComponent;
