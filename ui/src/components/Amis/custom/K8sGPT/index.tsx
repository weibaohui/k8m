import {replacePlaceholders} from '@/utils/utils';
import React, {useEffect, useState} from 'react';
import {fetcher} from "@/components/Amis/fetcher.ts";
import k8sDate from "../K8sDate";
import {Alert, Button, Card, List, message, Space, Tag, Typography} from 'antd';
import {ExclamationCircleOutlined, CheckCircleOutlined, QuestionCircleOutlined} from '@ant-design/icons';
import WebSocketMarkdownViewerComponent from '../WebSocketMarkdownViewer';

interface K8sGPTProps {
    data: Record<string, any>; // 泛型数据类型
    name: string;
    api: string;
}

interface K8sGPTResult {
    errors: any;
    status: string;
    problems: number;
    lastRunTime?: string;
    results: Array<{
        kind: string;
        name: string;
        error: Array<{
            Text: string;
            KubernetesDoc: string;
            Sensitive: Array<{
                Unmasked: string;
                Masked: string;
            }>;
        }>;
        details: string;
        parentObject: string;
    }>;
}

interface ApiResponse {
    status: number;
    msg?: string;
    data?: {
        status: number;
        msg?: string;
        data: K8sGPTResult;
    };
}

const K8sGPTComponent = React.forwardRef<HTMLDivElement, K8sGPTProps>((props, _) => {
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<K8sGPTResult | null>(null);
    const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});

    let finalUrl = replacePlaceholders(props.api, props.data);

    const handleGet = async () => {
        if (!finalUrl) return;
        setLoading(true);

        try {
            const response = await fetcher({
                url: finalUrl,
                method: 'get',
            }) as ApiResponse;

            if (response.data?.status !== 0) {
                message.error(`获取巡检结果失败:请尝试刷新后重试。 ${response.data?.msg}`);
            } else {
                const result = response.data.data;
                setResult(result);
            }
        } catch (error) {
            message.error('获取巡检结果失败，请稍后重试');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        handleGet();
    }, []);

    const toggleExplanation = (itemKey: string) => {
        setExpandedItems(prev => ({
            ...prev,
            [itemKey]: !prev[itemKey]
        }));
    };

    if (loading) {
        return <Card loading={true}/>;
    }

    if (!result) {
        return null;
    }

    return (
        <Card
            title={
                <Space>
                    <Typography.Text strong>K8s资源巡检结果</Typography.Text>
                    <Tag color={result.status === 'ProblemDetected' ? 'error' : 'success'}
                         icon={result.status === 'ProblemDetected' ? <ExclamationCircleOutlined/> :
                             <CheckCircleOutlined/>}>
                        {result.status === 'ProblemDetected' ? '发现问题' : '正常'}
                    </Tag>
                    {result.problems > 0 && (
                        <Tag color="warning">发现 {result.problems} 个问题</Tag>
                    )}
                    {result.lastRunTime && (
                        <Tag color="processing">最后运行: {k8sDate(result.lastRunTime)}</Tag>
                    )}
                </Space>
            }
        >
            <List
                dataSource={result.results}
                renderItem={item => {
                    const itemKey = `${item.kind}-${item.name}`;
                    return (
                        <List.Item>
                            <div style={{width: '100%'}}>
                                <Typography.Text strong style={{marginBottom: '8px', display: 'block'}}>
                                    {item.kind}: {item.name}
                                </Typography.Text>
                                <Space direction="vertical" style={{width: '100%'}}>
                                    {item.error.map((error, index) => (
                                        <div key={index} style={{width: '100%'}}>
                                            <Alert
                                                message={error.Text}
                                                description={error.KubernetesDoc}
                                                type="error"
                                                showIcon
                                                action={
                                                    <Button
                                                        icon={<QuestionCircleOutlined/>}
                                                        onClick={() => toggleExplanation(`${itemKey}-${index}`)}
                                                        type="link"
                                                    >
                                                        AI解释
                                                    </Button>
                                                }
                                            />
                                            {expandedItems[`${itemKey}-${index}`] && (
                                                <div style={{marginTop: '8px', marginBottom: '16px'}}>
                                                    <WebSocketMarkdownViewerComponent
                                                        url="/ai/chat/k8s_gpt/resource"
                                                        params={{
                                                            kind: item.kind,
                                                            name: item.name,
                                                            data: error.Text,
                                                            field: error.KubernetesDoc
                                                        }}
                                                        data={{}}
                                                    />
                                                </div>
                                            )}
                                        </div>
                                    ))}
                                </Space>
                            </div>
                        </List.Item>
                    );
                }}
            />
        </Card>
    );
});

export default K8sGPTComponent;
