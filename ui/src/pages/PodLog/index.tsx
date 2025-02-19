import React from 'react';
import {useSearchParams} from 'react-router-dom';
import PodLogViewer from '../../components/Amis/custom/LogView/PodLogViewer';


const PodLog: React.FC = () => {

    const [searchParams] = useSearchParams();
    const namespace = searchParams.get('namespace') || '';
    const name = searchParams.get('name') || '';

    if (!namespace || !name) {
        return <div>请在URL中提供namespace和name参数</div>;
    }

    return (
        <div style={{padding: '6px'}}>
            <PodLogViewer namespace={namespace} name={name} data={{}} showTitle={true}/>
        </div>
    );
};

export default PodLog;