import React, { useCallback, useState } from 'react';
import * as yaml from 'js-yaml';
import { Button, message } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';

interface KubeConfigProps {
    data: Record<string, any>;
}

interface ClusterInfo {
    clusterName: string;
    serverUrl: string;
    userName: string;
    namespace?: string;
    displayName: string;
}

const KubeConfigEditorComponent = React.forwardRef<HTMLDivElement, KubeConfigProps>(() => {
    const [editorContent, setEditorContent] = useState('');
    const [isValid, setIsValid] = useState(false);
    const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [displayName, setDisplayName] = useState('');
    const [displayNameError, setDisplayNameError] = useState<string | null>(null);

    const validateDisplayName = (name: string): boolean => {
        const regex = /^[a-zA-Z0-9_]+$/;
        return regex.test(name);
    };

    const validateAndParseConfig = useCallback((content: string) => {
        try {
            const config = yaml.load(content) as any;
            if (!config || typeof config !== 'object') {
                throw new Error('无效的YAML格式');
            }

            if (!config.clusters?.[0]?.name || !config.clusters?.[0]?.cluster?.server || !config.users?.[0]?.name) {
                throw new Error('缺少必要的配置信息');
            }

            setClusterInfo({
                clusterName: config.clusters[0].name,
                serverUrl: config.clusters[0].cluster.server,
                userName: config.users[0].name,
                namespace: config.contexts?.[0]?.context?.namespace,
                displayName: displayName || config.clusters[0].name
            });

            const isDisplayNameValid = validateDisplayName(displayName);
            setIsValid(isDisplayNameValid && displayName.trim() !== '');
            setError(null);
        } catch (err) {
            setIsValid(false);
            setError(err instanceof Error ? err.message : '无效的配置格式');
            setClusterInfo(null);
        }
    }, [displayName]);

    const handleDisplayNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newDisplayName = e.target.value;
        setDisplayName(newDisplayName);

        if (!validateDisplayName(newDisplayName) && newDisplayName.trim() !== '') {
            setDisplayNameError('集群名称只能包含英文字母、数字和下划线');
        } else {
            setDisplayNameError(null);
        }

        if (clusterInfo) {
            setClusterInfo({ ...clusterInfo, displayName: newDisplayName });
        }
        setIsValid(() => clusterInfo !== null && newDisplayName.trim() !== '' && validateDisplayName(newDisplayName));
    };

    const handleEditorChange = (value: string | undefined) => {
        const content = value || '';
        setEditorContent(content);
        validateAndParseConfig(content);
    };

    const handleSave = async () => {
        if (!isValid || !clusterInfo || !editorContent) return;

        setLoading(true);
        try {
            const response = await fetcher({
                url: '/admin/cluster/kubeconfig/save',
                method: 'post',
                data: {
                    content: editorContent,
                    server: clusterInfo.serverUrl,
                    user: clusterInfo.userName,
                    cluster: clusterInfo.clusterName,
                    namespace: clusterInfo.namespace,
                    display_name: clusterInfo.displayName
                }
            });

            if (response.data?.status === 0) {
                message.success('集群纳管成功');
            } else {
                throw new Error(response.data?.msg || '纳管失败');
            }
        } catch (error) {
            message.error('纳管失败：' + (error instanceof Error ? error.message : '未知错误'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {error && (
                <div style={{ color: 'red', padding: '8px', backgroundColor: '#ffebee', borderRadius: '4px' }}>
                    {error}
                </div>
            )}
            <div style={{
                padding: '12px',
                backgroundColor: '#e8f5ff',
                borderRadius: '4px',
                marginBottom: '12px',
                border: '1px solid #91caff'
            }}>
                <div style={{ color: '#1677ff' }}>请将kubeconfig文件内容粘贴到下面的编辑窗口</div>
            </div>
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '12px',
                gap: '8px'
            }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flex: 1 }}>
                    <span style={{ color: '#ff4d4f', marginRight: '4px' }}>*</span>
                    <span>名称:</span>
                    <input
                        type="text"
                        value={displayName}
                        onChange={handleDisplayNameChange}
                        style={{
                            padding: '4px 8px',
                            borderRadius: '4px',
                            flex: 1,
                            border: (!displayName.trim() || displayNameError) ? '1px solid #ff4d4f' : '1px solid #d9d9d9'
                        }}
                        placeholder="请输入集群显示名称（仅限英文字母、数字和下划线）"
                    />
                </div>
                {displayNameError && (
                    <div style={{ color: '#ff4d4f', fontSize: '12px', marginTop: '4px' }}>
                        {displayNameError}
                    </div>
                )}
                <Button
                    type="primary"
                    disabled={!isValid}
                    loading={loading}
                    onClick={handleSave}
                >
                    确认纳管
                </Button>
            </div>
            {(
                <div style={{
                    padding: '16px',
                    backgroundColor: '#f5f5f5',
                    borderRadius: '4px',
                    border: '1px solid #e0e0e0'
                }}>
                    <div style={{
                        marginBottom: '12px'
                    }}>
                        <h4 style={{ margin: 0 }}>配置信息</h4>
                    </div>
                    <div style={{ display: 'grid', gap: '8px' }}>
                        <div>集群名称: {clusterInfo?.clusterName}</div>
                        <div>服务器地址: {clusterInfo?.serverUrl}</div>
                        <div>用户名称: {clusterInfo?.userName}</div>
                        <div>命名空间: {clusterInfo?.namespace || '默认'}</div>
                    </div>
                </div>
            )}
            <div style={{ border: '1px solid #d9d9d9', borderRadius: '4px' }}>
                <textarea
                    style={{
                        width: '100%',
                        height: '300px',
                        padding: '8px',
                        fontFamily: 'monospace',
                        resize: 'none',
                        border: 'none',
                        outline: 'none'
                    }}
                    value={editorContent}
                    onChange={(e) => handleEditorChange(e.target.value)}
                    placeholder="请粘贴kubeconfig内容"
                />
            </div>
        </div>
    );
});

export default KubeConfigEditorComponent;