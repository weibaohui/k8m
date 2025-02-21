import React, { useEffect, useState } from 'react';
import DiffEditorComponent from '.';
import { fetcher } from "@/components/Amis/fetcher";
import { replacePlaceholders } from '@/utils/utils';

interface DeploymentRevisionDiffEditorProps {
    data: Record<string, any>;
    url: string;
    revision: string;
}

// 用于展示Deployment版本差异的组件
const DeploymentRevisionDiffEditor: React.FC<DeploymentRevisionDiffEditorProps> = ({
    url, data, revision
}) => {
    const [original, setOriginal] = useState<string>();
    const [modified, setModified] = useState<string>();
    url = replacePlaceholders(url, data);
    revision = replacePlaceholders(revision, data);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetcher({
                    url: url,
                    method: 'get',
                });
                //@ts-ignore
                setOriginal(response.data?.data.latest);
                //@ts-ignore
                setModified(response.data?.data.current);
            } catch (error) {
                console.error('Failed to fetch deployment revision data:', error);
            }
        };
        fetchData();
    }, [url]);
    return (
        <DiffEditorComponent
            originalValue={original}  // 左侧显示指定版本
            modifiedValue={modified}   // 右侧显示最新版本
            originalLabel="当前运行版本"
            modifiedLabel={`#${revision}版本`}
            height="calc(100vh - 200px)"  // 设置合适的高度
            width="100%"
            readOnly={true}               // 设置为只读模式
        />
    );
};

export default DeploymentRevisionDiffEditor;