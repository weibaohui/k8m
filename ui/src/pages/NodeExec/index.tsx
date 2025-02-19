import React, {useEffect, useState} from 'react';
import {Card} from 'antd';
import {useSearchParams} from 'react-router-dom';
import XTermComponent from '@/components/Amis/custom/XTerm';
import {fetcher} from '@/components/Amis/fetcher.ts';

interface PodData {
    podName: string;
    ns: string;
    containerName: string;
}

const NodeExec: React.FC = () => {
    const [searchParams] = useSearchParams();
    const nodeName = searchParams.get('name') || '';
    const [podShell, setPodSell] = useState<PodData>();

    useEffect(() => {
        if (!nodeName) return;

        // 获取Pod详情以获取容器列表
        fetcher({
            url: `/k8s/node/name/${nodeName}/create_node_shell`,
            method: 'post'
        })
            .then(response => {
                const data = response.data?.data as unknown as PodData;
                console.log(response.data?.data)
                console.log(data)
                setPodSell(data)
            })
            .catch(error => console.error('Error fetching pod details:', error));
    }, [nodeName]);

    if (!nodeName) {
        return <div>请在URL中提供节点名称参数</div>;
    }

    return (
        <div style={{padding: '6px'}}>
            <Card
                title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span>节点终端</span>
                        <span style={{ fontSize: '14px', color: 'rgba(0, 0, 0, 0.65)' }}>
                            {nodeName}
                        </span>
                    </div>
                }
                variant="outlined"
                style={{width: '100%', height: 'calc(100vh - 12px)'}}
            >
                {podShell && (
                    <div style={{background: '#f5f5f5', padding: '4px', borderRadius: '4px'}}>
                        <XTermComponent
                            url={`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/k8s/node/xterm/name/${nodeName}/pod/${podShell.podName}`}
                            params={{}}
                            data={{}}
                            height="calc(100vh - 120px)"
                        />
                    </div>
                )}
            </Card>
        </div>
    );
};

export default NodeExec;