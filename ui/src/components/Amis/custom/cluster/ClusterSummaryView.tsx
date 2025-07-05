import React, {useEffect, useState} from 'react';
import {fetcher} from '@/components/Amis/fetcher';
import {message, Card, Progress, Row, Col} from "antd";
import {Node} from "@/store/node.ts";

interface ClusterSummaryViewProps {
    data: Record<string, any>
}

interface ResourceSummary {
    cpu: {
        request: number;
        limit: number;
        total: number;
        available: number;
    };
    memory: {
        request: number;
        limit: number;
        total: number;
        available: number;
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
    return parseFloat(str);
}

const ClusterSummaryView = React.forwardRef<HTMLSpanElement, ClusterSummaryViewProps>(({data}, _) => {
    const [summary, setSummary] = useState<ResourceSummary | null>(null);

    useEffect(() => {
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
                let cpuRequest = 0, cpuLimit = 0, cpuTotal = 0;
                let memoryRequest = 0, memoryLimit = 0, memoryTotal = 0;
                let podUsed = 0, podTotal = 0;
                let ipUsed = 0, ipTotal = 0;
                nodes.forEach(n => {
                    const a = n.metadata.annotations || {};

                    cpuRequest += parseCpu(a["cpu.request"]);
                    cpuLimit += parseCpu(a["cpu.limit"]);
                    cpuTotal += parseCpu(n.status?.capacity?.cpu || "");
                    memoryRequest += parseMemory(a["memory.request"]);
                    memoryLimit += parseMemory(a["memory.limit"]);
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
                        total: cpuTotal || cpuLimit, // fallback
                        available: (cpuTotal || cpuLimit) - cpuLimit
                    },
                    memory: {
                        request: memoryRequest,
                        limit: memoryLimit,
                        total: memoryTotal || memoryLimit, // fallback
                        available: (memoryTotal || memoryLimit) - memoryLimit
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
        fetchValues();
    }, [data]);

    if (!summary) return null;

    return (
        <Row gutter={[16, 16]}>
            <Col span={12}>
                <Card title="CPU Resources">
                    <div>Requests: {summary.cpu.request.toFixed(2)} cores / Limits: {summary.cpu.limit.toFixed(2)} /
                        Total: {summary.cpu.total.toFixed(2)} cores
                    </div>
                    <div style={{margin: '8px 0'}}>
                        <span style={{color: '#1677ff'}}>Requests</span>
                        <Progress percent={summary.cpu.request / summary.cpu.total * 100} showInfo={false}
                                  strokeColor="#1677ff"/>
                        <span style={{color: '#fa8c16'}}>Limits</span>
                        <Progress percent={summary.cpu.limit / summary.cpu.total * 100} showInfo={false}
                                  strokeColor="#fa8c16"/>
                    </div>
                    <div>Available: {summary.cpu.available.toFixed(2)} cores</div>
                </Card>
            </Col>
            <Col span={12}>
                <Card title="Memory Resources">
                    <div>Requests: {summary.memory.request.toFixed(2)} GiB / Limits: {summary.memory.limit.toFixed(2)} /
                        Total: {summary.memory.total.toFixed(2)} GiB
                    </div>
                    <div style={{margin: '8px 0'}}>
                        <span style={{color: '#1677ff'}}>Requests</span>
                        <Progress percent={summary.memory.request / summary.memory.total * 100} showInfo={false}
                                  strokeColor="#1677ff"/>
                        <span style={{color: '#fa8c16'}}>Limits</span>
                        <Progress percent={summary.memory.limit / summary.memory.total * 100} showInfo={false}
                                  strokeColor="#fa8c16"/>
                    </div>
                    <div>Available: {summary.memory.available.toFixed(2)} GiB</div>
                </Card>
            </Col>
            <Col span={12}>
                <Card title="Pod Resources">
                    <div>Used: {summary.pod.used} / Total: {summary.pod.total}</div>
                    <Progress percent={summary.pod.used / summary.pod.total * 100} showInfo={false}
                              strokeColor="#fa8c16"/>
                    <div>Available: {summary.pod.available}</div>
                </Card>
            </Col>
            <Col span={12}>
                <Card title="IP Resources">
                    <div>Used: {summary.ip.used} / Total: {summary.ip.total}</div>
                    <Progress percent={summary.ip.used / summary.ip.total * 100} showInfo={false}
                              strokeColor="#fa8c16"/>
                    <div>Available: {summary.ip.available}</div>
                </Card>
            </Col>
        </Row>
    );
});

export default ClusterSummaryView;
