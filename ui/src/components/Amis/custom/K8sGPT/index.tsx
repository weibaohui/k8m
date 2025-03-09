import { replacePlaceholders } from '@/utils/utils';
import React, { useEffect, useState } from 'react';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Alert, Card, List, message, Space, Tag, Typography } from 'antd';
import { ExclamationCircleOutlined, CheckCircleOutlined } from '@ant-design/icons';

interface K8sGPTProps {
    data: Record<string, any>; // 泛型数据类型
    name: string;
    api: string;
}

interface K8sGPTResult {
    errors: any;
    status: string;
    problems: number;
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
// 定义 API 响应的接口
interface ApiResponse {
    status: number;
    msg?: string;
    data?: {
        status: number;
        msg?: string;
        data: {
            result?: string | K8sGPTResult;
        }
    };
}
// 用 forwardRef 包装组件
const K8sGPTComponent = React.forwardRef<HTMLDivElement, K8sGPTProps>((props, _) => {
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<K8sGPTResult | null>(null);

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
                // 将JSON字符串转换为结构体
                const result = typeof response.data.data?.result === 'string'
                    ? JSON.parse(response.data.data.result)
                    : response.data.data?.result;
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

    if (loading) {
        return <Card loading={true} />;
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
                        icon={result.status === 'ProblemDetected' ? <ExclamationCircleOutlined /> : <CheckCircleOutlined />}>
                        {result.status === 'ProblemDetected' ? '发现问题' : '正常'}
                    </Tag>
                    {result.problems > 0 && (
                        <Tag color="warning">发现 {result.problems} 个问题</Tag>
                    )}
                </Space>
            }
        >
            <List
                dataSource={result.results}
                renderItem={item => (
                    <List.Item>
                        <div style={{ width: '100%' }}>
                            <Typography.Text strong style={{ marginBottom: '8px', display: 'block' }}>
                                {item.kind}: {item.name}
                            </Typography.Text>
                            <Space direction="vertical" style={{ width: '100%' }}>
                                {item.error.map((error, index) => (
                                    <Alert
                                        key={index}
                                        message={error.Text}
                                        description={error.KubernetesDoc}
                                        type="error"
                                        showIcon
                                    />
                                ))}
                            </Space>
                        </div>
                    </List.Item>
                )}
            />
        </Card>
    );
});

export default K8sGPTComponent;
