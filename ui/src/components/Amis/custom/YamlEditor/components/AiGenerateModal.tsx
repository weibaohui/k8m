import React, { useState } from 'react';
import { Modal, Input, Button, Space, Typography, Spin, Alert, List, Tag } from 'antd';
import { RobotOutlined, SendOutlined, BulbOutlined, CopyOutlined } from '@ant-design/icons';
import { fetcher } from '@/components/Amis/fetcher';

const { TextArea } = Input;
const { Title, Text, Paragraph } = Typography;

interface AiGenerateModalProps {
    visible: boolean;
    onCancel: () => void;
    onGenerateSuccess: (yaml: string) => void;
}

const EXAMPLES = [
    { label: '部署 Nginx', description: '创建一个 Nginx 部署，3 个副本，使用 latest 镜像' },
    { label: '部署 Redis', description: '创建一个 Redis 单实例部署，设置密码为 redis123' },
    { label: '部署 MySQL', description: '创建一个 MySQL 部署，1 个副本，设置 root 密码' },
    { label: '部署 Node.js', description: '创建一个 Node.js 应用，暴露 3000 端口' },
];

const AiGenerateModal: React.FC<AiGenerateModalProps> = ({
    visible,
    onCancel,
    onGenerateSuccess
}) => {
    const [prompt, setPrompt] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);
    const [generatedYaml, setGeneratedYaml] = useState<string>('');
    const [error, setError] = useState<string>('');

    const handleGenerate = async () => {
        if (!prompt.trim()) {
            setError('请输入描述内容');
            return;
        }

        setLoading(true);
        setError('');
        setGeneratedYaml('');

        try {

            // 调用后端API
            const result = await fetcher({
                url: '/mgm/plugins/yaml_editor/ai/generate',
                method: 'post',
                data: JSON.stringify({ prompt })
            });

            console.log('API响应:', result);
            //@ts-ignore
            if (result.status === 200 && result.data && result.data.data && result.data.data.yaml) {
                //@ts-ignore
                setGeneratedYaml(result.data.data.yaml);
            } else {
                //@ts-ignore
                throw new Error(result.msg || '生成失败');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : '生成失败，请重试');
        } finally {
            setLoading(false);
        }
    };

    const handleConfirm = () => {
        if (generatedYaml) {
            onGenerateSuccess(generatedYaml);
            handleCancel();
        }
    };

    const handleCancel = () => {
        setPrompt('');
        setGeneratedYaml('');
        setError('');
        onCancel();
    };

    const handleExampleClick = (example: typeof EXAMPLES[0]) => {
        setPrompt(example.description);
    };

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(generatedYaml);
        } catch (err) {
            console.error('复制失败:', err);
        }
    };

    return (
        <Modal
            title={
                <Space>
                    <RobotOutlined />
                    <span>AI 辅助生成 YAML</span>
                </Space>
            }
            open={visible}
            onCancel={handleCancel}
            width={800}
            footer={generatedYaml ? (
                <Space>
                    <Button onClick={handleCancel}>取消</Button>
                    <Button icon={<CopyOutlined />} onClick={handleCopy}>复制</Button>
                    <Button type="primary" onClick={handleConfirm}>
                        确认使用
                    </Button>
                </Space>
            ) : (
                <Space>
                    <Button onClick={handleCancel}>取消</Button>
                    <Button
                        type="primary"
                        icon={<SendOutlined />}
                        onClick={handleGenerate}
                        loading={loading}
                        disabled={!prompt.trim()}
                    >
                        生成 YAML
                    </Button>
                </Space>
            )}
            destroyOnClose
        >
            <Space direction="vertical" style={{ width: '100%' }} size="large">
                {!generatedYaml && (
                    <>
                        <div>
                            <Title level={5}>
                                <BulbOutlined /> 示例描述
                            </Title>
                            <List
                                dataSource={EXAMPLES}
                                renderItem={(item) => (
                                    <List.Item
                                        style={{ cursor: 'pointer', padding: '8px' }}
                                        onClick={() => handleExampleClick(item)}
                                        onMouseEnter={(e) => {
                                            e.currentTarget.style.backgroundColor = '#f0f0f0';
                                        }}
                                        onMouseLeave={(e) => {
                                            e.currentTarget.style.backgroundColor = 'transparent';
                                        }}
                                    >
                                        <Space>
                                            <Tag color="blue">{item.label}</Tag>
                                            <Text type="secondary">{item.description}</Text>
                                        </Space>
                                    </List.Item>
                                )}
                            />
                        </div>

                        <div>
                            <Title level={5}>自然语言描述</Title>
                            <TextArea
                                value={prompt}
                                onChange={(e) => setPrompt(e.target.value)}
                                placeholder="请描述你想要创建的 Kubernetes 资源，例如：创建一个 Nginx 部署，3 个副本..."
                                rows={4}
                                onPressEnter={(e) => {
                                    if (e.shiftKey) {
                                        return;
                                    }
                                    e.preventDefault();
                                    handleGenerate();
                                }}
                            />
                            <Paragraph type="secondary" style={{ marginTop: 8, fontSize: 12 }}>
                                提示：按 Enter 生成，Shift+Enter 换行
                            </Paragraph>
                        </div>

                        {error && (
                            <Alert
                                message="生成失败"
                                description={error}
                                type="error"
                                showIcon
                            />
                        )}

                        {loading && (
                            <div style={{ textAlign: 'center', padding: '40px' }}>
                                <Spin size="large" />
                                <div style={{ marginTop: 16 }}>AI 正在生成 YAML 配置...</div>
                            </div>
                        )}
                    </>
                )}

                {generatedYaml && (
                    <div>
                        <Title level={5}>生成的 YAML</Title>
                        <pre
                            style={{
                                backgroundColor: '#f5f5f5',
                                padding: '16px',
                                borderRadius: '4px',
                                maxHeight: '400px',
                                overflow: 'auto',
                                fontSize: '12px',
                                fontFamily: 'monospace',
                            }}
                        >
                            {generatedYaml}
                        </pre>
                    </div>
                )}
            </Space>
        </Modal>
    );
};

export default AiGenerateModal;
