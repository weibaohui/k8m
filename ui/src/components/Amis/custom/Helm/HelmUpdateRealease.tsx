import React, { useState, useEffect } from 'react';
import { Button, Col, Form, Row, Typography, message } from 'antd';
import Editor from '@monaco-editor/react';
import { fetcher } from '@/components/Amis/fetcher';

interface HelmUpdateReleaseProps {
    data: Record<string, any>
}

const HelmUpdateRelease = React.forwardRef<HTMLSpanElement, HelmUpdateReleaseProps>(({ data }, _) => {
    const [values, setValues] = useState('');
    const [loading, setLoading] = useState(false);
    const [clusterInfo, setClusterInfo] = useState('');

    useEffect(() => {
        const originCluster = localStorage.getItem('cluster') || '';
        setClusterInfo(originCluster ? originCluster : '未选择集群');
    }, []);

    let chartName = data.chart || ''
    let releaseName = data.name || ''
    let namespace = data.namespace || 'default'
    let revision = data.revision || ''

    if (!releaseName || !namespace) {
        message.error('缺少必要的 Release 信息');
        return null;
    }



    useEffect(() => {

        const fetchValues = async () => {

            try {
                const response = await fetcher({
                    url: `/k8s/helm/release/ns/${namespace}/name/${releaseName}/revision/${revision}/values`,
                    method: 'get'
                });
                setValues((response.data as any)?.data || '');
            } catch (error) {
                message.error('获取参数值失败');
            }
        };
        fetchValues();
    }, [namespace, releaseName, revision]);


    const handleSubmit = async () => {

        setLoading(true);
        try {
            await fetcher({
                url: '/k8s/helm/release/upgrade',
                method: 'post',
                data: {
                    values,
                    name: releaseName,
                    namespace: namespace
                }
            });
            message.success('更新成功');
        } catch (error) {
            message.error('更新失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <Form layout="horizontal" labelCol={{ span: 4 }} wrapperCol={{ span: 20 }}>

                <Form.Item label="更新操作">
                    <Button
                        type="primary"
                        onClick={handleSubmit}
                        loading={loading}
                        style={{ marginRight: 16 }}
                    >
                        提交更新
                    </Button>


                </Form.Item>
                <Form.Item label="基本信息">
                    <Row justify={'start'}>
                        <Col span={8}>
                            <Form.Item label="所属集群" labelCol={{ span: 8 }} wrapperCol={{ span: 16 }}>
                                <Typography.Text ellipsis={{ tooltip: true }}>{clusterInfo}</Typography.Text>
                            </Form.Item>
                        </Col>
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
                    <div style={{ border: '1px solid #d9d9d9', borderRadius: '4px' }}
                    >
                        <Editor
                            height="calc(100vh - 200px)"
                            language="yaml"
                            value={values}
                            onChange={(value) => setValues(value || '')}
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


export default HelmUpdateRelease;