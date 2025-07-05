import React, {useState, useEffect} from 'react';
import {fetcher} from '@/components/Amis/fetcher';
import {message} from "antd";

interface ClusterSummaryViewProps {
    data: Record<string, any>
}

const ClusterSummaryView = React.forwardRef<HTMLSpanElement, ClusterSummaryViewProps>(({data}, _) => {
    const [values, setValues] = useState('');

    useEffect(() => {
        const fetchValues = async () => {
            try {

                const response = await fetcher({
                    url: `/k8s/Node/group//version/v1/list`,
                    method: 'get'
                });
                setValues((response.data as any)?.data || '');
            } catch (error) {
                message.error('获取参数值失败');
            }
        };
        fetchValues();
    }, [data]);
    console.log(values)
    return (
        <></>
    );
});

export default ClusterSummaryView;
