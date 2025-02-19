import React, { useEffect, useState } from 'react';
import { Select, Card } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';
import SSELogDisplayComponent from '@/components/Amis/custom/LogView/SSELogDisplay';
import SSELogDownloadComponent from '@/components/Amis/custom/LogView/SSELogDownload';
import LogOptionsComponent from '@/components/Amis/custom/LogView/LogOptions';
import { replacePlaceholders } from '@/utils/utils';

interface PodSpec {
    containers: Container[];
}

interface PodData {
    spec: PodSpec;
}

interface Container {
    name: string;
}

interface PodLogViewerProps {
    namespace: string;
    name: string;
    data: Record<string, any>;

}

const PodLogViewerComponent: React.FC<PodLogViewerProps> = ({ namespace, name, data }) => {

    namespace = replacePlaceholders(namespace, data);
    name = replacePlaceholders(name, data);
    const url = `/k8s/Pod/group//version/v1/ns/${namespace}/name/${name}/json`;

    const [containers, setContainers] = useState<Container[]>([]);
    const [selectedContainer, setSelectedContainer] = useState<string>('');

    const [tailLines, setTailLines] = React.useState(100);
    const [follow, setFollow] = React.useState(true);
    const [timestamps, setTimestamps] = React.useState(false);
    const [previous, setPrevious] = React.useState(false);
    const [sinceTime, setSinceTime] = React.useState<string>();

    useEffect(() => {
        if (!namespace || !name) return;

        fetcher({ url: url, method: 'get' })
            .then(response => {
                const data = response.data?.data as unknown as PodData;

                if (data.spec?.containers) {
                    setContainers(data.spec.containers);
                    if (data.spec.containers.length > 0) {
                        setSelectedContainer(data.spec.containers[0].name);
                    }
                }
            })
            .catch(error => console.error('Error fetching pod details:', error));
    }, [namespace, name]);

    return (
        <Card
            title={
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <span>容器日志</span>
                    <div>
                        {namespace}/{name}
                    </div>
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
                {selectedContainer && (
                    <SSELogDownloadComponent
                        url={`/k8s/pod/logs/download/ns/${namespace}/pod_name/${name}/container/${selectedContainer}`}
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
                {selectedContainer && (
                    <SSELogDisplayComponent
                        url={`/k8s/pod/logs/sse/ns/${namespace}/pod_name/${name}/container/${selectedContainer}`}
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