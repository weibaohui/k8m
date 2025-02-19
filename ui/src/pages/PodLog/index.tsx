import React, { useEffect, useState } from 'react';
import { Select, Card } from 'antd';
import { useSearchParams } from 'react-router-dom';
import { fetcher } from '../../components/Amis/fetcher';
import SSELogDisplayComponent from '../../components/Amis/custom/LogView/SSELogDisplay';
import SSELogDownloadComponent from '../../components/Amis/custom/LogView/SSELogDownload';
import LogOptionsComponent from '../../components/Amis/custom/LogView/LogOptions';
import PodLogViewer from '../../components/Amis/custom/LogView/PodLogViewer';

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

    if (!namespace || !name) {
        return <div>请在URL中提供namespace和name参数</div>;
    }

    return (
        <div style={{ padding: '6px' }}>
            <PodLogViewer namespace={namespace} name={name} data={{}} showTitle={true} />
        </div>
    );
};

export default PodLog;