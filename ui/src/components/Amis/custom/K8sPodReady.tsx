import React from 'react';
import {Tooltip} from 'antd';
import {Pod} from '@/store/pod';

interface K8sPodReadyProps {
    data: Pod;
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

    // 获取未就绪容器的信息
    const containerReadyCondition = data.status?.conditions?.find(
        condition => condition.type === 'ContainersReady'
    );

    // 获取未就绪的容器列表
    const unreadyContainers = containerStatuses
        .filter(status => !status.ready)
        .map(status => status.name);

    // 根据容器就绪状态设置样式
    const isAllReady = readyCount === totalCount;

    const tooltipContent = isAllReady ? null : (
        <div>
            <div>未就绪容器：</div>
            {unreadyContainers.map((name, index) => (
                <div key={index}>• {name}</div>
            ))}
            {containerReadyCondition?.message && (
                <div style={{marginTop: '8px'}}>{containerReadyCondition.message}</div>
            )}
        </div>
    );

    const statusElement = (
        <span className={`text font-medium ${isAllReady ? 'text-black' : 'text-danger'}`}>
            {readyStatus}
        </span>
    );

    return (
        <span ref={ref}>
            {isAllReady ? (
                statusElement
            ) : (unreadyContainers.length > 0 ? (
                <Tooltip title={tooltipContent}>
                    {statusElement}
                </Tooltip>
            ) : (
                statusElement
            ))}
        </span>
    );
});

export default K8sPodReadyComponent;
