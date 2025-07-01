import React, { useState, useEffect } from 'react';
import { Col, Form, Row, Typography, message } from 'antd';
import Editor from '@monaco-editor/react';
import { fetcher } from '@/components/Amis/fetcher';

interface HelmViewReleaseProps {
    data: Record<string, any>
}

const HelmViewRelease = React.forwardRef<HTMLSpanElement, HelmViewReleaseProps>(({ data }, _) => {
    const [values, setValues] = useState('');

    let chartName = data.chart
    let releaseName = data.name
    let namespace = data.namespace
    let revision = data.revision

    useEffect(() => {
        const fetchValues = async () => {
            try {
                const response = await fetcher({
                    url: `/k8s/helm/release/ns/${namespace}/name/${releaseName}/revision/${revision}/values`,
                    method: 'post'
                });
                // @ts-ignore
                setValues(response.data?.data || '');
            } catch (error) {
                message.error('获取参数值失败');
            }
        };
        fetchValues();
    }, [namespace, releaseName, revision]);

    return (
        <div>
            <Form layout="horizontal" labelCol={{ span: 4 }} wrapperCol={{ span: 20 }}>
                <Form.Item label="基本信息">
                    <Row justify={'start'}>

                        <Col span={6}>
                            <Form.Item label="发布名称" labelCol={{ span: 8 }} wrapperCol={{ span: 16 }}>
                                <Typography.Text ellipsis={{ tooltip: true }}>{namespace}/{releaseName}</Typography.Text>
                            </Form.Item>
                        </Col>
                        <Col span={6}>
                            <Form.Item label="Chart名称" labelCol={{ span: 8 }} wrapperCol={{ span: 16 }}>
                                <Typography.Text ellipsis={{ tooltip: true }}>{chartName}</Typography.Text>
                            </Form.Item>
                        </Col>
                    </Row>
                </Form.Item>
                <Form.Item label="安装参数">
                    <div style={{ border: '1px solid #d9d9d9', borderRadius: '4px' }}>
                        <Editor
                            height="calc(100vh - 200px)"
                            language="yaml"
                            value={values}
                            options={{
                                minimap: { enabled: false },
                                scrollBeyondLastLine: false,
                                automaticLayout: true,
                                wordWrap: 'on',
                                scrollbar: {
                                    vertical: 'auto',
                                    verticalScrollbarSize: 8
                                }
                            }}
                        />
                    </div>
                </Form.Item>
            </Form>
        </div>
    );
});

export default HelmViewRelease;
