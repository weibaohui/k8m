import React from 'react';

interface K8sPodReadyProps {
    data: {
        status?: {
            containerStatuses?: { ready: boolean }[];
        };
        spec?: {
            containers?: any[];
        };
    };
}

// 用 forwardRef 让组件兼容 AMIS
const K8sPodReadyComponent = React.forwardRef<HTMLSpanElement, K8sPodReadyProps>(({data}, ref) => {
    // 获取 Pod 状态中的容器状态列表
    const containerStatuses = data.status?.containerStatuses || [];

    // 获取定义的容器总数
    const containerSpecs = data.spec?.containers || [];

    // 计算处于 Ready 状态的容器数量
    const readyCount = containerStatuses.filter(status => status.ready).length;

    // 总的容器数量应该从 spec 中获取
    const totalCount = containerSpecs.length;

    // 格式化状态 "N/M"
    const readyStatus = `${readyCount}/${totalCount}`;

    return <span ref={ref}>{readyStatus}</span>;
});

export default K8sPodReadyComponent;
