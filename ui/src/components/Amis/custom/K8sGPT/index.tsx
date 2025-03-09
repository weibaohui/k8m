import { replacePlaceholders } from '@/utils/utils';
import React, { useEffect, useState } from 'react';
import { fetcher } from "@/components/Amis/fetcher.ts";
import { message } from 'antd';

interface K8sGPTProps {
    data: Record<string, any>; // 泛型数据类型
    name: string;
    api: string;
}

// 用 forwardRef 包装组件
const K8sGPTComponent = React.forwardRef<HTMLDivElement, K8sGPTProps>((props, _) => {
    const [loading, setLoading] = useState(false);

    console.log(props);
    console.log(props.api);
    let finalUrl = replacePlaceholders(props.api, props.data);
    console.log(finalUrl);
    const handleGet = async () => {
        if (!finalUrl) return;
        setLoading(true);


        const response = await fetcher({
            url: finalUrl,
            method: 'get',
        });

        if (response.data?.status !== 0) {
            message.error(`获取巡检结果失败:请尝试刷新后重试。 ${response.data?.msg}`);
        } else {
            //@ts-ignore
            console.log(response.data.data?.result);
        }
        setLoading(false);
    };

    useEffect(() => {
        handleGet();
    }, []);
    return <div >x</div>;
});

export default K8sGPTComponent;
