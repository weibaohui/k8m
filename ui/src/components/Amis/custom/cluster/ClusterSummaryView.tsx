import React, { useEffect, useState } from 'react';
import { fetcher } from '@/components/Amis/fetcher';
import { message, Card, Progress, Row, Col, Avatar, Statistic, Select, Space, Spin } from "antd";
import { Node } from "@/store/node.ts";
import CountUp from 'react-countup';
import type { StatisticProps } from 'antd';

const formatter: StatisticProps['formatter'] = (value) => (
    <CountUp end={value as number} separator="," />
);
interface ClusterSummaryViewProps {
    data: Record<string, any>
}

interface ResourceCount {
    Count: number,
    Group: string,
    Version: string,
    Resource: string
}

interface ResourceSummary {
    cpu: {
        request: number;
        limit: number;
        realtime: number;
        total: number;
        available: number;
        requestFraction: string; // 百分比字符串
        limitFraction: string; // 上限百分比字符串
        realtimeFraction: string; // 实时百分比字符串
    };
    memory: {
        request: number;
        limit: number;
        realtime: number;
        total: number;
        available: number;
        requestFraction: string; // 请求百分比字符串
        limitFraction: string; // 上限百分比字符串
        realtimeFraction: string; // 实时百分比字符串
    };
    pod: {
        used: number;
        total: number;
        available: number;
    };
    ip: {
        used: number;
        total: number;
        available: number;
    };
}

function parseCpu(str: string) {
    if (!str) return 0;
    if (str.endsWith('core')) return parseFloat(str);
    if (str.endsWith('m')) return parseFloat(str) / 1000;
    return parseFloat(str);
}

function parseMemory(str: string) {
    if (!str) return 0;
    if (str.endsWith('Gi')) return parseFloat(str);
    if (str.endsWith('Mi')) return parseFloat(str) / 1024;
    if (str.endsWith('Ki')) return parseFloat(str) / 1024 / 1024;
    return parseFloat(str);
}

const refreshOptions = [
    { label: '不自动刷新', value: 0 },
    { label: '10秒', value: 10000 },
    { label: '30秒', value: 30000 },
    { label: '60秒', value: 60000 },
];

const ClusterSummaryView = React.forwardRef<HTMLSpanElement, ClusterSummaryViewProps>(({ data }, _) => {
    const [summary, setSummary] = useState<ResourceSummary | null>(null);
    const [resourceGroups, setResourceGroups] = useState<Record<string, ResourceCount[]>>({});
    const [refreshInterval, setRefreshInterval] = useState<number>(0);
    const [loading, setLoading] = useState<boolean>(false);
    const intervalRef = React.useRef<NodeJS.Timeout | null>(null);

    // 数据获取函数
    const fetchAll = async () => {
        setLoading(true);
        try {
            await Promise.all([fetchValues(), fetchResource()]);
        } finally {
            setLoading(false);
        }
    };
    const fetchValues = async () => {
        try {
            const response = await fetcher({
                url: `/k8s/Node/group//version/v1/list`,
                method: 'post',
                data: {
                    page: 1,
                    perPage: 100000
                }
            });
            //@ts-ignore
            const nodes = response.data?.data?.rows as Array<Node>;
            // 汇总
            let cpuRequest = 0, cpuLimit = 0, cpuRealtime = 0, cpuTotal = 0;
            let memoryRequest = 0, memoryLimit = 0, memoryRealtime = 0, memoryTotal = 0;
            let podUsed = 0, podTotal = 0;
            let ipUsed = 0, ipTotal = 0;
            nodes.forEach(n => {
                const a = n.metadata.annotations || {};

                cpuRequest += parseCpu(a["cpu.request"]);
                cpuLimit += parseCpu(a["cpu.limit"]);
                cpuRealtime += parseCpu(a["cpu.realtime"]);
                cpuTotal += parseCpu(n.status?.capacity?.cpu || "");
                memoryRequest += parseMemory(a["memory.request"]);
                memoryLimit += parseMemory(a["memory.limit"]);
                memoryRealtime += parseMemory(a["memory.realtime"]);
                memoryTotal += parseMemory(n.status?.capacity?.memory || "");
                podUsed += parseInt(a["pod.count.used"] || "0");
                podTotal += parseInt(a["pod.count.total"] || "0");
                ipUsed += parseInt(a["ip.usage.used"] || "0");
                ipTotal += parseInt(a["ip.usage.total"] || "0");
            });
            setSummary({
                cpu: {
                    request: cpuRequest,
                    limit: cpuLimit,
                    realtime: cpuRealtime,
                    total: cpuTotal,
                    available: (cpuTotal || cpuLimit) - cpuRealtime,
                    requestFraction: cpuTotal > 0 ? ((cpuRequest / cpuTotal * 100).toFixed(2)) : '0.00',
                    limitFraction: cpuTotal > 0 ? ((cpuLimit / cpuTotal * 100).toFixed(2)) : '0.00',
                    realtimeFraction: cpuTotal > 0 ? ((cpuRealtime / cpuTotal * 100).toFixed(2)) : '0.00'
                },
                memory: {
                    request: memoryRequest,
                    limit: memoryLimit,
                    realtime: memoryRealtime,
                    total: memoryTotal,
                    available: (memoryTotal || memoryLimit) - memoryRealtime,
                    requestFraction: memoryTotal > 0 ? ((memoryRequest / memoryTotal * 100).toFixed(2)) : '0.00',
                    limitFraction: memoryTotal > 0 ? ((memoryLimit / memoryTotal * 100).toFixed(2)) : '0.00',
                    realtimeFraction: memoryTotal > 0 ? ((memoryRealtime / memoryTotal * 100).toFixed(2)) : '0.00'
                },
                pod: {
                    used: podUsed,
                    total: podTotal,
                    available: podTotal - podUsed
                },
                ip: {
                    used: ipUsed,
                    total: ipTotal,
                    available: ipTotal - ipUsed
                }
            });
        } catch (error) {
            message.error('获取参数值失败');
        }
    };
    const fetchResource = async () => {
        try {
            const response = await fetcher({
                url: `/k8s/status/resource_count/cache_seconds/60`,
                method: 'get',
            });
            let counts = response.data?.data as Array<ResourceCount>;
            // 按 group 分组
            const groups: Record<string, ResourceCount[]> = {};
            counts?.forEach(item => {
                if (!groups[item.Group]) groups[item.Group] = [];
                groups[item.Group].push(item);
            });
            setResourceGroups(groups);
        } catch (error) {
            message.error('获取参数值失败');
        }
    };
    // 首次和依赖变化时获取
    useEffect(() => {
        fetchAll();
    }, [data]);

    // 自动刷新逻辑
    useEffect(() => {
        if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
        }
        if (refreshInterval > 0) {
            intervalRef.current = setInterval(() => {
                fetchAll();
            }, refreshInterval);
        }
        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
                intervalRef.current = null;
            }
        };
    }, [refreshInterval]);

    if (!summary) return null;

    return (
        <Spin spinning={loading} tip="数据加载中...">
            <Row justify="end" style={{ marginBottom: 16 }}>
                <Col>
                    <Space>
                        <span>自动刷新：</span>
                        <Select
                            style={{ width: 120 }}
                            value={refreshInterval}
                            options={refreshOptions}
                            onChange={setRefreshInterval}
                        />
                    </Space>
                </Col>
            </Row>
            <Row gutter={[16, 16]}>
                <Col span={12}>
                    <Card title="CPU （cores）">
                        <div>请求: {summary.cpu.request.toFixed(2)} / 上限: {summary.cpu.limit.toFixed(2)} /
                            共计: {summary.cpu.total.toFixed(2)} / 实时: {summary.cpu.realtime.toFixed(2)} /
                            可用: {summary.cpu.available.toFixed(2)} </div>
                        <div style={{ margin: '8px 0' }}>
                            <span style={{ color: '#1677ff' }}>请求 {summary.cpu.requestFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.cpu.requestFraction)}
                                strokeColor="#1677ff" showInfo={false} />
                            <span style={{ color: '#fa8c16' }}>上限 {summary.cpu.limitFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.cpu.limitFraction)}
                                strokeColor="#fa8c16" showInfo={false} />
                            <span style={{ color: '#52c41a' }}>实时 {summary.cpu.realtimeFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.cpu.realtimeFraction)}
                                strokeColor="#52c41a" showInfo={false} />
                        </div>
                    </Card>
                </Col>
                <Col span={12}>
                    <Card title="内存 （GiB）">
                        <div>请求: {summary.memory.request.toFixed(2)} / 上限: {summary.memory.limit.toFixed(2)} /
                            共计: {summary.memory.total.toFixed(2)} / 实时: {summary.memory.realtime.toFixed(2)} /
                            可用: {summary.memory.available.toFixed(2)} </div>
                        <div style={{ margin: '8px 0' }}>
                            <span style={{ color: '#1677ff' }}>请求 {summary.memory.requestFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.memory.requestFraction)}
                                strokeColor="#1677ff" showInfo={false} />
                            <span style={{ color: '#fa8c16' }}>上限 {summary.memory.limitFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.memory.limitFraction)}
                                strokeColor="#fa8c16" showInfo={false} />
                            <span style={{ color: '#52c41a' }}>实时 {summary.memory.realtimeFraction}%</span>
                            <Progress size="small" percent={parseFloat(summary.memory.realtimeFraction)} showInfo={false}
                                strokeColor="#52c41a" />
                        </div>
                    </Card>
                </Col>
            </Row>
            {/* ResourceCount 分组展示 */}
            <div style={{ marginTop: 24 }}>
                {Object.entries(resourceGroups)
                    .sort(([a], [b]) => a.localeCompare(b))
                    .map(([group, items]: [string, ResourceCount[]], groupIdx) => (
                        <Card key={group} title={group} style={{ marginBottom: 16 }}>
                            <Row gutter={[16, 16]}>
                                {items.sort((a, b) => a.Resource.localeCompare(b.Resource)).map((item: ResourceCount) => {
                                    const colors = ['#1677ff', '#fa8c16', '#52c41a', '#eb2f96', '#13c2c2', '#722ed1'];
                                    const color = colors[groupIdx % colors.length];
                                    return (
                                        <Col key={item.Resource} span={6}>
                                            <div style={{ display: 'flex', alignItems: 'center', background: '#f6f8fa', borderRadius: 8, padding: 12 }}>
                                                <Avatar style={{ backgroundColor: color, verticalAlign: 'middle', marginRight: 12 }} size="large">
                                                    {item.Resource?.[0]?.toUpperCase() || '?'}
                                                </Avatar>
                                                <div>
                                                    <div style={{ fontSize: 18, fontWeight: 600 }}>
                                                        <Statistic value={item.Count} formatter={formatter} />
                                                    </div>

                                                    <div style={{ fontSize: 14, color: '#888' }}>{item.Resource}({item.Version})</div>
                                                </div>
                                            </div>
                                        </Col>
                                    );
                                })}
                            </Row>
                        </Card>
                    ))}
            </div>
        </Spin>
    );
});

export default ClusterSummaryView;
