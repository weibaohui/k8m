import {useEffect, useState} from 'react';
import {fetcher} from '@/components/Amis/fetcher';

export interface ClusterOption {
    label: string;
    value: string;
}

export function useClusterOptions() {
    const [options, setOptions] = useState<ClusterOption[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        setLoading(true);
        fetcher({url: '/params/cluster/option_list', method: 'get'})
            .then((res: any) => {
                if (res?.data?.data?.options) {
                    setOptions(res.data.data?.options);
                } else {
                    setOptions([]);
                }
            })
            .catch((err: any) => {
                setError(err.message || '获取集群列表失败');
                setOptions([]);
            })
            .finally(() => setLoading(false));
    }, []);

    return {options, loading, error};
}
