import React, {useState, useEffect} from 'react';
import {Button, Col, Form, Row, Select, Typography, message} from 'antd';
import Editor from '@monaco-editor/react';
import {fetcher} from '@/components/Amis/fetcher';
import yaml from "js-yaml";

interface HelmUpdateReleaseProps {
    releaseName: string;
    repoName: string;
    chart: {
        metadata: {
            name: string;
            version: string;
        };
    };
    data: Record<string, any>
}

const HelmUpdateRelease = React.forwardRef<HTMLSpanElement, HelmUpdateReleaseProps>(({data}, _) => {
    const [versions, setVersions] = useState<string[]>([]);
    const [version, setVersion] = useState('');
    const [values, setValues] = useState('');
    const [loading, setLoading] = useState(false);
    const [isFetching, setIsFetching] = useState(false);
    const [clusterInfo, setClusterInfo] = useState('');

    useEffect(() => {
        const originCluster = localStorage.getItem('cluster') || '';
        setClusterInfo(originCluster ? originCluster : '未选择集群');
    }, []);
    let repoName = data.info.description
    let chartName = data.chart.metadata.name
    let releaseName = data.name
    let namespace = data.namespace
    useEffect(() => {
        const ov = yaml.dump(data.config, {
            indent: 2,
            lineWidth: -1,  // 禁用自动换行
            noRefs: true    // 避免引用标记
        });
        setValues(ov)
    }, [data.config])

    useEffect(() => {
        const fetchVersions = async () => {
            try {
                const response = await fetcher({
                    url: `/k8s/helm/repo/${repoName}/chart/${chartName}/versions`,
                    method: 'get'
                });
                // @ts-ignore
                var options = response.data?.data?.options;
                // options = options.filter((opt: string) => opt !== data.chart.metadata.version);
                setVersions(options || []);
            } catch (error) {
                message.error('获取版本列表失败');
            }
        };
        fetchVersions();
    }, [chartName]);
    const fetchValues = async () => {
        if (!version) {
            message.error('请先选择版本');
            return;
        }
        setIsFetching(true);
        try {
            const response = await fetcher({
                url: `/mgm/helm/repo/${repoName}/chart/${chartName}/version/${version}/values`,
                method: 'get'
            });
            // @ts-ignore
            setValues(response.data?.data.yaml || '');
        } catch (error) {
            message.error('获取参数值失败');
        } finally {
            setIsFetching(false);
        }
    };


    const handleSubmit = async () => {
        if (!version) {
            message.error('请选择一个版本');
            return;
        }
        setLoading(true);
        try {
            await fetcher({
                url: '/mgm/helm/release/upgrade',
                method: 'post',
                data: {
                    values,
                    release_name: releaseName,
                    repo_name: repoName,
                    version: version,
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
            <Form layout="horizontal" labelCol={{span: 4}} wrapperCol={{span: 20}}>

                <Form.Item label="更新操作">
                    <Button
                        type="primary"
                        onClick={handleSubmit}
                        loading={loading}
                        style={{marginRight: 16}}
                    >
                        提交更新
                    </Button>
                    <Button
                        type="default"
                        onClick={fetchValues}
                        loading={isFetching}
                        style={{marginRight: 16}}
                    >
                        🗳️ 加载Chart包默认参数
                    </Button>
                    <Button
                        type="default"
                        onClick={() => {
                            const ov = yaml.dump(data.config, {
                                indent: 2,
                                lineWidth: -1,
                                noRefs: true
                            });
                            setValues(ov);
                        }}
                    >
                        ⚙️ 使用用户输入参数
                    </Button>
                </Form.Item>
                <Form.Item label="基本信息">
                    <Row justify={'start'}>
                        <Col span={8}>
                            <Form.Item label="所属集群" labelCol={{span: 8}} wrapperCol={{span: 16}}>
                                <Typography.Text ellipsis={{tooltip: true}}>{clusterInfo}</Typography.Text>
                            </Form.Item>
                        </Col>
                        <Col span={6}>
                            <Form.Item label="发布名称" labelCol={{span: 8}} wrapperCol={{span: 16}}>
                                <Typography.Text ellipsis={{tooltip: true}}>{namespace}/{releaseName}</Typography.Text>
                            </Form.Item>
                        </Col>
                        <Col span={6}>
                            <Form.Item label="Chart名称" labelCol={{span: 8}} wrapperCol={{span: 16}}>
                                <Typography.Text ellipsis={{tooltip: true}}>{chartName}</Typography.Text>
                            </Form.Item>
                        </Col>
                    </Row>
                </Form.Item>
                <Form.Item label="升/降版本">
                    <Row gutter={16}>

                        <Col span={12}>
                            <Select
                                value={version}
                                onChange={setVersion}
                                options={(Array.isArray(versions) ? versions : []).map(v => ({label: v, value: v}))}
                                placeholder="请选择目标版本"
                            />
                        </Col>
                        <Col span={12}>
                            <div style={{lineHeight: '32px'}}>
                                当前版本：{data.chart.metadata.version}
                            </div>
                        </Col>
                    </Row>
                </Form.Item>

                <Form.Item label="安装参数">
                    <div style={{border: '1px solid #d9d9d9', borderRadius: '4px'}}
                    >
                        <Editor
                            height="600px"
                            language="yaml"
                            value={values}
                            onChange={(value) => setValues(value || '')}
                            options={{
                                minimap: {enabled: false},
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