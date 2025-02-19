import React, { useEffect, useState } from 'react';
import { Card, Spin } from 'antd';
import { useSearchParams } from 'react-router-dom';
import { fetcher } from '@/components/Amis/fetcher.ts';
import FileExplorerComponent from '@/components/Amis/custom/FileExplorer/FileExplorer';
import { Pod } from '@/store/pod';


interface PodShell {
    podName: string;
    ns: string;
    containerName: string;
    pod: Pod
}

const NodeExec: React.FC = () => {
    const [searchParams] = useSearchParams();
    const nodeName = searchParams.get('name') || '';
    const [podShell, setPodShell] = useState<PodShell>();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string>();

    useEffect(() => {
        if (!nodeName) return;

        setLoading(true);
        setError(undefined);

        // 获取Pod详情以获取容器列表
        fetcher({
            url: `/k8s/node/name/${nodeName}/create_node_shell`,
            method: 'post'
        })
            .then(response => {
                const data = response.data?.data as unknown as PodShell;
                if (!data) {
                    throw new Error('未能获取节点终端信息');
                }
                setPodShell(data);
                setError(undefined);
            })
            .catch(error => {
                console.error('Error fetching pod details:', error);
                setError(error.message || '获取节点终端失败');
                setPodShell(undefined);
            })
            .finally(() => {
                setLoading(false);
            });
    }, [nodeName]);

    if (!nodeName) {
        return <div>请在URL中提供节点名称参数</div>;
    }

    return (
        <div style={{ padding: '6px' }}>
            <Card
                title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span>节点终端</span>
                        <span style={{ fontSize: '14px', color: 'rgba(0, 0, 0, 0.65)' }}>
                            {nodeName}
                        </span>
                        <span>（ESC 关闭）</span>
                    </div>
                }
                variant="outlined"
                style={{ width: '100%', height: 'calc(100vh - 12px)' }}
            >
                <div style={{ padding: '4px', borderRadius: '4px', minHeight: '400px' }}>
                    <Spin spinning={loading} tip="正在加载节点终端...">
                        {error && (
                            <div style={{ color: '#ff4d4f', textAlign: 'center', padding: '6px' }}>{error}</div>
                        )}
                    </Spin>
                    {podShell && (
                        <FileExplorerComponent data={podShell.pod} remove='true' />)
                    }
                </div>
            </Card>
        </div>
    );
};

export default NodeExec;