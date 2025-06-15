import React, { useState, useEffect, useCallback } from 'react';
import {
    Alert,
    Card,
    DatePicker,
    Form,
    Spin,
    Table,
    Typography,
    Space,
    Dropdown,
    MenuProps,
    Tag,
    Drawer,
    Select
} from "antd";
import dayjs, { Dayjs } from 'dayjs';
import { fetcher } from '@/components/Amis/fetcher';
import { replacePlaceholders, toUrlSafeBase64 } from '@/utils/utils';
import { DownOutlined } from '@ant-design/icons';
import InspectionEventListComponent from './InspectionEventList';
import { useClusterOptions } from './useClusterOptions';

const { Title, Text } = Typography;

interface InspectionSummaryComponentProps {
    schedule_id: string;
    data: Record<string, any>;
}

const InspectionSummaryComponent = React.forwardRef<HTMLDivElement, InspectionSummaryComponentProps>(({
    schedule_id,
    data
}, _) => {
    // 表单状态
    const [form] = Form.useForm();
    const [startTime, setStartTime] = useState<Dayjs>(() => dayjs().startOf('day'));
    const [endTime, setEndTime] = useState<Dayjs>(() => dayjs().add(1, 'day').startOf('day'));
    const [loading, setLoading] = useState(false);
    const [summaryData, setSummaryData] = useState<any>({});
    const [error, setError] = useState<string | null>(null);
    const [drawerOpen, setDrawerOpen] = useState(false);
    const [drawerRecordId, setDrawerRecordId] = useState<number | null>(null);
    const { options: clusterOptions, loading: clusterLoading } = useClusterOptions();
    const [cluster, setCluster] = useState<string | undefined>(undefined);



    let realScheduleId = ""
    // 处理 schedule_id 占位符
    if (schedule_id !== undefined) {
        // 处理 schedule_id 占位符
        realScheduleId = replacePlaceholders(schedule_id, data) || "";
    }

    // 查询API，使用fetcher
    const fetchSummary = useCallback((params?: { startTime?: Dayjs, endTime?: Dayjs }) => {
        setLoading(true);
        setError(null);
        const sTime = (params?.startTime || startTime).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
        const eTime = (params?.endTime || endTime).format('YYYY-MM-DDTHH:mm:ss') + 'Z';
        let url = '';
        let clusterBase64 = "";
        if (cluster) {
            clusterBase64 = toUrlSafeBase64(cluster)
        }
        url = `/admin/inspection/schedule/id/${realScheduleId}/summary/cluster/${clusterBase64}/start_time/${encodeURIComponent(sTime)}/end_time/${encodeURIComponent(eTime)}`;

        fetcher({ url, method: 'post' })
            .then((response: any) => {
                if (response?.data?.data) {
                    setSummaryData(response.data.data);
                } else {
                    setSummaryData({});
                    setError('未获取到数据');
                }
            })
            .catch((err: any) => {
                setError(err.message || '未知错误');
                setSummaryData({});
            })
            .finally(() => {
                setLoading(false);
            });
    }, [startTime, endTime, cluster, realScheduleId]);

    // 外部参数变化自动刷新
    useEffect(() => {
        fetchSummary();
    }, [startTime, endTime, fetchSummary]);

    // cluster 变化时刷新
    useEffect(() => {
        fetchSummary();
    }, [cluster]);

    // antd表格列定义
    const latestRunColumns = [
        { title: '资源类型', dataIndex: 'kind', key: 'kind' },
        {
            title: '异常数', dataIndex: 'error_count', key: 'error_count',
            render: (text: any, _: any) => (
                <span style={{ color: '#ff4d4f', cursor: 'pointer' }}
                    onClick={() => {
                        setDrawerRecordId(latest_run.record_id);
                        setDrawerOpen(true);
                    }}
                >{text}</span>
            )
        }
    ];
    const clusterColumns = [
        { title: '资源类型', dataIndex: 'kind', key: 'kind' },
        { title: '异常数', dataIndex: 'error_count', key: 'error_count' }
    ];

    const { total_runs, total_clusters, latest_run = {}, clusters = [], total_schedules } = summaryData || {};

    // 时间快捷选项
    const quickRanges = [
        { label: "1月", value: { start: dayjs().subtract(1, 'month').startOf('day'), end: dayjs().endOf('day') } },
        { label: "1周", value: { start: dayjs().subtract(7, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "2天", value: { start: dayjs().subtract(2, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "1天", value: { start: dayjs().subtract(1, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "6小时", value: { start: dayjs().subtract(6, 'hour'), end: dayjs() } },
        { label: "1小时", value: { start: dayjs().subtract(1, 'hour'), end: dayjs() } },
    ];
    const quickMenuProps: MenuProps = {
        items: quickRanges.map((item, idx) => ({
            key: idx.toString(),
            label: item.label,
        })),
        onClick: (info) => {
            const idx = Number(info.key);
            const range = quickRanges[idx];
            setStartTime(range.value.start);
            setEndTime(range.value.end);
            form.setFieldsValue({ startTime: range.value.start, endTime: range.value.end });
            fetchSummary({ startTime: range.value.start, endTime: range.value.end });
        },
    };


    return (
        <div>
            <Card style={{ marginBottom: 16 }}>
                <Form
                    form={form}
                    layout="inline"
                    initialValues={{ startTime, endTime, cluster }}
                    onFinish={values => {
                        setStartTime(values.startTime);
                        setEndTime(values.endTime);
                        setCluster(values.cluster || undefined);
                        fetchSummary(values);
                    }}
                >
                    <Form.Item label="起始时间" name="startTime" rules={[{ required: true, message: '请选择起始时间' }]}>
                        <DatePicker showTime format="YYYY-MM-DD HH:mm" value={startTime} allowClear={false} />
                    </Form.Item>
                    <Form.Item label="结束时间" name="endTime" rules={[{ required: true, message: '请选择结束时间' }]}>
                        <DatePicker showTime format="YYYY-MM-DD HH:mm" value={endTime} allowClear={false} />
                    </Form.Item>
                    <Form.Item label="集群" name="cluster" initialValue={cluster || ''}>
                        <Select
                            style={{ minWidth: 220 }}
                            loading={clusterLoading}
                            allowClear
                            placeholder="全部集群"
                            onChange={val => {
                                setCluster(val || undefined);
                                form.submit();
                            }}
                        >
                            <Select.Option value="">全部集群</Select.Option>
                            {clusterOptions.map(opt => (
                                <Select.Option key={opt.value} value={opt.value}>{opt.label}</Select.Option>
                            ))}
                        </Select>
                    </Form.Item>
                    <Form.Item>
                        <Dropdown.Button
                            icon={<DownOutlined />}
                            menu={quickMenuProps}
                            placement="bottomLeft"
                            onClick={() => form.submit()}
                            type="primary"
                            loading={loading}
                        >
                            查询
                        </Dropdown.Button>
                    </Form.Item>
                </Form>
            </Card>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} showIcon />}
            <Spin spinning={loading} tip="加载中...">
                <Card style={{ marginBottom: 16 }}>
                    <Space>
                        <Text strong>总执行次数：</Text> <Text>{total_runs ?? '-'}</Text>
                        <Text strong>总集群数：</Text> <Text>{total_clusters ?? '-'}</Text>
                        {total_schedules !== undefined && (
                            <>
                                <Text strong>运行巡检计划数：</Text> <Text>{total_schedules}</Text>
                            </>
                        )}
                    </Space>
                </Card>
                {latest_run && latest_run.record_id && (
                    <Card
                        title={
                            <span>
                                <span style={{ marginRight: 8 }}>
                                    <b>最后执行信息</b>
                                </span>
                                <span style={{ marginRight: 8 }}>
                                    <b>计划ID：</b>
                                    <Tag color="blue">{latest_run.schedule_id ?? '-'}</Tag>
                                </span>
                                <span style={{ marginRight: 8, cursor: 'pointer' }}
                                    onClick={() => {
                                        setDrawerRecordId(latest_run.record_id);
                                        setDrawerOpen(true);
                                    }}
                                >
                                    <b>记录ID：</b>
                                    <Tag color="green">{latest_run.record_id}</Tag>
                                </span>
                                <span>
                                    <b>执行时间：</b>
                                    <Tag
                                        color="volcano">{latest_run.run_time ? dayjs(latest_run.run_time).format('YYYY-MM-DD HH:mm:ss') : '-'}</Tag>
                                </span>
                            </span>
                        }
                        style={{ marginBottom: 16 }}
                    >
                        <Table
                            columns={latestRunColumns}
                            dataSource={latest_run.kinds || []}
                            rowKey={(r: any) => r.kind}
                            pagination={false}
                            size="small"
                        />
                    </Card>
                )}
                <Title level={5} style={{ margin: '16px 0 8px 0' }}>汇总数据：</Title>
                {clusters.length === 0 && <Text type="secondary">暂无集群数据</Text>}
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
            <Drawer
                title="事件明细"
                width={900}
                open={drawerOpen}
                onClose={() => setDrawerOpen(false)}
                destroyOnClose
            >
                {drawerRecordId &&
                    <InspectionEventListComponent record_id={`${drawerRecordId}`} />}
            </Drawer>
        </div>
    );
});

export default InspectionSummaryComponent;
