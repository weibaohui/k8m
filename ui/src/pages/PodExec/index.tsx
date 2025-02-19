import React, { useEffect, useState } from 'react';
import { Select, Card } from 'antd';
import { useSearchParams } from 'react-router-dom';
import XTermComponent from '../../components/Amis/custom/XTerm';
import { fetcher } from '../../components/Amis/fetcher';
import { Container, Pod } from '@/store/pod';


const PodExec: React.FC = () => {
    const [searchParams] = useSearchParams();
    const namespace = searchParams.get('namespace') || '';
    const name = searchParams.get('name') || '';
    const [containers, setContainers] = useState<Container[]>([]);
    const [selectedContainer, setSelectedContainer] = useState<string>('');

    useEffect(() => {
        if (!namespace || !name) return;

        // 获取Pod详情以获取容器列表
        fetcher({
            url: `/k8s/Pod/group//version/v1/ns/${namespace}/name/${name}/json`,
            method: 'get'
        })
            .then(response => {
                const data = response.data?.data as unknown as Pod;

                if (data.spec?.containers) {
                    setContainers(data.spec.containers);
                    if (data.spec.containers.length > 0) {
                        setSelectedContainer(data.spec.containers[0].name);
                    }
                }
            })
            .catch(error => console.error('Error fetching pod details:', error));
    }, [namespace, name]);

    if (!namespace || !name) {
        return <div>请在URL中提供namespace和name参数</div>;
    }

    return (
        <div style={{ padding: '6px' }}>
            <Card
                title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span>容器终端</span>
                        <span style={{ fontSize: '14px', color: 'rgba(0, 0, 0, 0.65)' }}>
                            {namespace}/{name}
                        </span>
                        <Select
                            style={{ width: 200 }}
                            value={selectedContainer}
                            onChange={setSelectedContainer}
                            options={containers.map(container => ({
                                label: container.name,
                                value: container.name
                            }))}
                            placeholder="选择容器"
                        />
                    </div>
                }
                variant="outlined"
                style={{ width: '100%', height: 'calc(100vh - 12px)' }}
            >
                {selectedContainer && (
                    <div style={{ background: '#f5f5f5', padding: '4px', borderRadius: '4px' }}>
                        <XTermComponent
                            url={`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/k8s/pod/xterm/ns/${namespace}/pod_name/${name}`}
                            params={{ container_name: selectedContainer }}
                            data={{}}
                            height="calc(100vh - 120px)"
                        />
                    </div>
                )}
            </Card>
        </div>
    );
};

export default PodExec;