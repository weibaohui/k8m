import React, {useEffect, useState} from 'react';
import {Card} from 'antd';
import {useSearchParams} from 'react-router-dom';
import {fetcher} from '@/components/Amis/fetcher.ts';
import FileExplorerComponent from "@/components/Amis/custom/FileExplorer/FileExplorer.tsx";
import {Pod} from "@/store/pod.ts";


const PodExec: React.FC = () => {
    const [searchParams] = useSearchParams();
    const namespace = searchParams.get('namespace') || '';
    const name = searchParams.get('name') || '';
    const [pod, setPod] = useState<Pod>();

    useEffect(() => {
        if (!namespace || !name) return;

        // 获取Pod详情以获取容器列表
        fetcher({
            url: `/k8s/Pod/group//version/v1/ns/${namespace}/name/${name}/json`,
            method: 'get'
        })
            .then(response => {
                const data = response.data?.data as unknown as Pod;
                setPod(data)
            })
            .catch(error => console.error('Error fetching pod details:', error));
    }, [namespace, name]);

    if (!namespace || !name) {
        return <div>请在URL中提供namespace和name参数</div>;
    }


    return (
        <div style={{padding: '6px'}}>
            <Card
                title={
                    <div style={{display: 'flex', alignItems: 'center', gap: '12px'}}>
                        <span>容器终端</span>
                        <span style={{fontSize: '14px', color: 'rgba(0, 0, 0, 0.65)'}}>
                            {namespace}/{name}
                        </span>

                    </div>
                }
                variant="outlined"
                style={{width: '100%', height: 'calc(100vh - 12px)'}}
            >
                {pod && (
                    <FileExplorerComponent data={pod} remove='false'/>)
                }
            </Card>
        </div>
    );
};

export default PodExec;