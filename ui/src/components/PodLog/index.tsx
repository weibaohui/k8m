import React, { useEffect, useState } from 'react';
import { Select, InputNumber, Switch, DatePicker, Card } from 'antd';
import { render as amisRender } from "amis";
import { useSearchParams } from 'react-router-dom';
import { fetcher } from '../Amis/fetcher';
import SSELogDisplayComponent from '../Amis/custom/SSELogDisplay';
import SSELogDownloadComponent from '../Amis/custom/SSELogDownload';

interface PodSpec {
    containers: Container[];
}

interface PodData {
    spec: PodSpec;
}
interface Container {
    name: string;
}
const PodLog: React.FC = () => {

    const [searchParams] = useSearchParams();
    const namespace = searchParams.get('namespace') || '';
    const name = searchParams.get('name') || '';
    const [containers, setContainers] = useState<Container[]>([]);
    const [selectedContainer, setSelectedContainer] = useState<string>('');


    const [tailLines, setTailLines] = React.useState(100);
    const [follow, setFollow] = React.useState(true);
    const [timestamps, setTimestamps] = React.useState(false);
    const [previous, setPrevious] = React.useState(false);
    const [sinceTime, setSinceTime] = React.useState<string>();


    useEffect(() => {
        if (!namespace || !name) return;

        // 获取Pod详情以获取容器列表
        fetcher({
            url: `/k8s/Pod/group//version/v1/ns/${namespace}/name/${name}/json`,
            method: 'get'
        })
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

    if (!namespace || !name) {
        return <div>请在URL中提供namespace和name参数</div>;
    }





    const logDownloadSchema = {
        type: 'log-download',
        url: `/k8s/pod/logs/download/ns/${namespace}/pod_name/${name}/container/${selectedContainer}`,
        data: {
            selectedContainer,
            tailLines,
            sinceTime,
            previous,
            timestamps,

        }
    };

    const logDisplaySchema = {
        type: 'log-display',
        url: `/k8s/pod/logs/sse/ns/${namespace}/pod_name/${name}/container/${selectedContainer}`,
        data: {
            selectedContainer,
            tailLines,
            sinceTime,
            follow,
            previous,
            timestamps,
        }
    };

    return (
        <div style={{ padding: '6px' }}>
            <Card
                title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span>容器日志</span>
                        <div >
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
                    <InputNumber
                        value={tailLines}
                        onChange={(value) => setTailLines(value ?? 100)} // 使用默认值100
                        min={0}
                        prefix="行数"
                        placeholder="显示行数"
                        style={{ width: "120px" }}
                    />
                    <Switch
                        checked={follow}
                        onChange={setFollow}
                        checkedChildren="实时"
                        unCheckedChildren="实时"
                    />
                    <Switch
                        checked={timestamps}
                        onChange={setTimestamps}
                        checkedChildren="时间戳"
                        unCheckedChildren="时间戳"
                    />
                    <Switch
                        checked={previous}
                        onChange={setPrevious}
                        checkedChildren="上一个"
                        unCheckedChildren="上一个"
                    />
                    <DatePicker
                        showTime
                        format="YYYY-MM-DD HH:mm:ss"
                        onChange={(date) => setSinceTime(date?.format('YYYY-MM-DD HH:mm:ss'))}
                        placeholder="选择开始时间"
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
                        ></SSELogDownloadComponent>
                    )}
                </div>
                <div style={{ background: '#f5f5f5', padding: '4px', borderRadius: '4px', height: 'calc(100vh - 150px)', overflow: 'auto' }}>

                    {selectedContainer && (
                        <>
                            <SSELogDisplayComponent
                                url={`/k8s/pod/logs/sse/ns/${namespace}/pod_name/${name}/container/${selectedContainer}`}
                                data={{
                                    tailLines: tailLines,
                                    sinceTime: sinceTime,
                                    follow: follow,
                                    previous: previous,
                                    timestamps: timestamps,
                                }}
                            ></SSELogDisplayComponent>
                        </>
                    )}
                </div>
            </Card>
        </div>
    );
};

export default PodLog;