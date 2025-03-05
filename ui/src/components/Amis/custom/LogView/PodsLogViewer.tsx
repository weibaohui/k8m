import React, { useEffect, useState } from 'react';
import { Select, Card } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';
import SSELogDisplayComponent from '@/components/Amis/custom/LogView/SSELogDisplay';
import SSELogDownloadComponent from '@/components/Amis/custom/LogView/SSELogDownload';
import LogOptionsComponent from '@/components/Amis/custom/LogView/LogOptions';
import { replacePlaceholders } from '@/utils/utils';
import { Container, Pod } from '@/store/pod';


interface PodLogViewerProps {
    url: string;
    data: Record<string, any>;
}

const PodLogViewerComponent: React.FC<PodLogViewerProps> = ({ url, data }) => {
    url = replacePlaceholders(url, data);

    const [pods, setPods] = useState<Pod[]>([]);
    const [selectedPod, setSelectedPod] = useState<{ name: string; namespace: string }>();
    const [containers, setContainers] = useState<Container[]>([]);
    const [selectedContainer, setSelectedContainer] = useState<string>('');

    const [tailLines, setTailLines] = React.useState(100);
    const [follow, setFollow] = React.useState(true);
    const [timestamps, setTimestamps] = React.useState(false);
    const [previous, setPrevious] = React.useState(false);
    const [sinceTime, setSinceTime] = React.useState<string>();


    // 在 useEffect 中处理 fetcher 的响应
    useEffect(() => {
        fetcher({ url: url, method: 'get' })
            .then((response) => {
                //@ts-ignore
                if (response?.data?.data?.rows) {
                    //@ts-ignore
                    const podList = response.data.data?.rows;
                    setPods(podList);
                    if (podList.length > 0) {
                        const firstPod = podList[0];
                        setSelectedPod({
                            namespace: firstPod.metadata.namespace,
                            name: firstPod.metadata.name
                        });
                    }
                } else {
                    console.warn('No pod data found in response:', response);
                    setPods([]);
                }
            })
            .catch(error => {
                console.error('Error fetching pod details:', error);
                setPods([]);
            });
    }, [url]);

    useEffect(() => {
        if (selectedPod) {
            const podData = pods.find(pod =>
                pod.metadata.name === selectedPod.name &&
                pod.metadata.namespace === selectedPod.namespace
            );
            // 合并 initContainers 和 containers
            const allContainers = [
                ...(podData?.spec?.containers || []),
                ...(podData?.spec?.initContainers || []),
                ...(podData?.spec?.ephemeralContainers || []),
            ];

            if (allContainers.length > 0) {
                setContainers(allContainers);
                // 默认选择第一个容器
                setSelectedContainer(allContainers[0].name);
            } else {
                setContainers([]);
                setSelectedContainer('');
            }
        }
    }, [selectedPod, pods]);

    return (
        <Card
            title={
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <Select
                        style={{ minWidth: 200 }}
                        value={selectedPod ? selectedPod.name : undefined}
                        onChange={(value) => {
                            const namespace = pods.find(pod => pod.metadata.name === value)?.metadata.namespace || '';
                            setSelectedPod({ namespace, name: value });
                        }}
                        options={pods.map(pod => ({
                            label: pod.metadata.name,
                            value: pod.metadata.name
                        }))}
                        placeholder="选择Pod"
                    />
                    <Select
                        style={{ minWidth: 200 }}
                        value={selectedContainer}
                        onChange={setSelectedContainer}
                        options={containers.map(container => ({
                            label: container.name,
                            value: container.name
                        }))}
                        placeholder="选择容器"
                        disabled={!selectedPod}
                    />
                </div>
            }
            variant="outlined"
            style={{ width: '100%', height: 'calc(100vh - 12px)' }}
        >
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <LogOptionsComponent
                    tailLines={tailLines}
                    follow={follow}
                    timestamps={timestamps}
                    previous={previous}
                    sinceTime={sinceTime}
                    onTailLinesChange={setTailLines}
                    onFollowChange={setFollow}
                    onTimestampsChange={setTimestamps}
                    onPreviousChange={setPrevious}
                    onSinceTimeChange={setSinceTime}
                />
                {selectedContainer && selectedPod && (
                    <SSELogDownloadComponent
                        url={`/k8s/pod/logs/download/ns/${selectedPod.namespace}/pod_name/${selectedPod.name}/container/${selectedContainer}`}
                        data={{
                            tailLines: tailLines,
                            sinceTime: sinceTime,
                            previous: previous,
                            timestamps: timestamps,
                        }}
                    />
                )}
            </div>
            <div style={{ background: '#f5f5f5', padding: '4px', borderRadius: '4px', height: 'calc(100vh - 150px)', overflow: 'auto' }}>
                {selectedContainer && selectedPod && (
                    <SSELogDisplayComponent
                        url={`/k8s/pod/logs/sse/ns/${selectedPod.namespace}/pod_name/${selectedPod.name}/container/${selectedContainer}`}
                        data={{
                            tailLines: tailLines,
                            sinceTime: sinceTime,
                            follow: follow,
                            previous: previous,
                            timestamps: timestamps,
                        }}
                    />
                )}
            </div>
        </Card>
    );
};

export default PodLogViewerComponent;