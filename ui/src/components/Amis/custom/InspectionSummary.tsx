import React, { useState, useEffect, useCallback } from 'react';
import { Alert, Button, Card, DatePicker, Form, Spin, Table, Typography, Space, MenuProps } from "antd";
import dayjs, { Dayjs } from 'dayjs';
import { fetcher } from '@/components/Amis/fetcher';
import { replacePlaceholders } from '@/utils/utils';
import { Dropdown, Menu } from "antd";
import { DownOutlined, UserOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;

interface InspectionSummaryComponentProps {
    schedule_id: string;
    data: Record<string, any>;
}

/**
 * InspectionSummaryComponent 组件
 * 优化：
 * 1. 结构更清晰，表单与数据展示分离
 * 2. 代码风格统一，类型更严格
 * 3. 错误与加载提示更友好
 * 4. 支持外部 schedule_id、data 变化自动刷新
 */
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
        const url = `/admin/inspection/schedule/id/${realScheduleId}/summary/start_time/${encodeURIComponent(sTime)}/end_time/${encodeURIComponent(eTime)}`;
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
    }, [startTime, endTime]);

    // 外部参数变化自动刷新
    useEffect(() => {
        fetchSummary();
    }, [startTime, endTime, fetchSummary]);

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

    const { total_runs, total_clusters, latest_run = {}, clusters = [], total_schedules } = summaryData || {};

    const handleMenuClick: MenuProps['onClick'] = (e) => {
        console.log('click', e);
    };
    // 时间快捷选项
    const quickRanges = [
        { label: "1月", value: { start: dayjs().subtract(1, 'month').startOf('day'), end: dayjs().endOf('day') } },
        { label: "1周", value: { start: dayjs().subtract(7, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "2天", value: { start: dayjs().subtract(2, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "1天", value: { start: dayjs().subtract(1, 'day').startOf('day'), end: dayjs().endOf('day') } },
        { label: "6小时", value: { start: dayjs().subtract(6, 'hour'), end: dayjs() } },
        { label: "1小时", value: { start: dayjs().subtract(1, 'hour'), end: dayjs() } },
    ];
    const quickMenu = (
        <Menu>
            {quickRanges.map((item, idx) => (
                <Menu.Item key={idx} onClick={() => {
                    setStartTime(item.value.start);
                    setEndTime(item.value.end);
                    form.setFieldsValue({ startTime: item.value.start, endTime: item.value.end });
                }}>{item.label}</Menu.Item>
            ))}
        </Menu>
    );


    const items: MenuProps['items'] = [
        {
            label: '1st menu item',
            key: '1',
            icon: <UserOutlined />,
        },
        {
            label: '2nd menu item',
            key: '2',
            icon: <UserOutlined />,
        },
        {
            label: '3rd menu item',
            key: '3',
            icon: <UserOutlined />,
            danger: true,
        },
        {
            label: '4rd menu item',
            key: '4',
            icon: <UserOutlined />,
            danger: true,
            disabled: true,
        },
    ];
    const menuProps = {
        items,
        onClick: handleMenuClick,
    };
    return (
        <div>
            <Card style={{ marginBottom: 16 }}>
                <Form
                    form={form}
                    layout="inline"
                    initialValues={{ startTime, endTime }}
                    onFinish={values => {
                        setStartTime(values.startTime);
                        setEndTime(values.endTime);
                        fetchSummary(values);
                    }}
                >
                    <Form.Item label="起始时间" name="startTime" rules={[{ required: true, message: '请选择起始时间' }]}>
                        <DatePicker showTime format="YYYY-MM-DD HH:mm" value={startTime} allowClear={false} />
                    </Form.Item>
                    <Form.Item label="结束时间" name="endTime" rules={[{ required: true, message: '请选择结束时间' }]}>
                        <DatePicker showTime format="YYYY-MM-DD HH:mm" value={endTime} allowClear={false} />
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit" loading={loading}>查询</Button>
                    </Form.Item>
                    <Form.Item>
                        <Dropdown overlay={quickMenu} placement="bottomLeft">
                            <Button>最近</Button>
                        </Dropdown>
                        <Dropdown menu={menuProps}>
                            <Button>
                                <Space>
                                    最近时间
                                    <DownOutlined />
                                </Space>
                            </Button>
                        </Dropdown>
                    </Form.Item>
                </Form>
            </Card>
            {error && <Alert type="error" message={error} style={{ marginBottom: 8 }} showIcon />}
            <Spin spinning={loading} tip="加载中...">
                <Card style={{ marginBottom: 16 }}>
                    <Space>
                        <Text strong>总执行次数：</Text> <Text>{total_runs ?? '-'}</Text>
                        <Text strong>总集群数：</Text> <Text>{total_clusters ?? '-'}</Text>
                        {/* 新增：运行巡检计划数 */}
                        {total_schedules !== undefined && (
                            <>
                                <Text strong>运行巡检计划数：</Text> <Text>{total_schedules}</Text>
                            </>
                        )}
                    </Space>
                </Card>
                {latest_run && latest_run.record_id && (
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
        </div>
    );
});

export default InspectionSummaryComponent;
