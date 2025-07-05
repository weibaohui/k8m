import React, { useEffect } from 'react';
import { fetcher } from '@/components/Amis/fetcher';
import { message } from "antd";
import { Node } from "@/store/node.ts";

interface ClusterSummaryViewProps {
    data: Record<string, any>
}


const ClusterSummaryView = React.forwardRef<HTMLSpanElement, ClusterSummaryViewProps>(({ data }, _) => {

    useEffect(() => {
        const fetchValues = async () => {
            try {

                const response = await fetcher({
                    url: `/k8s/Node/group//version/v1/list`,
                    method: 'post',
                    data: {
                        page: 1,
                        perPage: 100000
                    }
                });
                //@ts-ignore
                const nodes = response.data?.data?.rows as Array<Node>;
                console.log(nodes[0].metadata.annotations)
            } catch (error) {
                message.error('获取参数值失败');
            }
        };
        fetchValues();
    }, [data]);
    return (
        <></>
    );
});

export default ClusterSummaryView;
