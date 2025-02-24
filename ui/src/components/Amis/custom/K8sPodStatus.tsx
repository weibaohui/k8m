import React from 'react';
import {Tooltip} from 'antd';
import {Pod} from '@/store/pod';

interface K8sPodStatusProps {
    data: Pod
}

interface StatusConfig {
    status: string;
    color: string;
    label: string;
}

const statusConfigs: StatusConfig[] = [
    {status: 'Running', color: 'success', label: '运行中'},
    {status: 'Pending', color: 'warning', label: '调度中'},
    {status: 'Succeeded', color: 'primary', label: '成功完成'},
    {status: 'Failed', color: 'error', label: '失败'},
    {status: 'Unknown', color: 'warning', label: '未知'},
];

// 用 forwardRef 让组件兼容 AMIS
const K8sPodStatusComponent = React.forwardRef<HTMLDivElement, K8sPodStatusProps>(({data}, ref) => {
    // 获取 Pod 状态中的容器状态列表
    const containerStatuses = data.status?.containerStatuses || [];
    const phase = data.status?.phase || 'Unknown';

    // 获取当前状态的配置
    const currentStatus = statusConfigs.find(config => config.status === phase) || statusConfigs[4]; // 默认使用 Unknown

    // 检查容器状态中是否有错误
    const errorContainers = containerStatuses.filter(status => {
        const waiting = status.state?.waiting;
        const terminated = status.state?.terminated;
        return (waiting && waiting.reason !== 'ContainerCreating') || (terminated && terminated?.reason !== 'Completed');
    });

    return (
        <div ref={ref} style={{display: 'flex', flexDirection: 'column', gap: '4px'}}>
            <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                <span
                    className={`label label-${errorContainers.length > 0 ? 'warning' : currentStatus.color}`}>{currentStatus.label}</span>
            </div>
            {
                errorContainers.length > 0 && (
                    <div>
                        {errorContainers.map((container, index) => {
                            const waiting = container.state?.waiting;
                            const terminated = container.state?.terminated;
                            const message = waiting?.message || terminated?.message || '';
                            const reason = waiting?.reason || terminated?.reason || '';
                            return (
                                <Tooltip
                                    key={index}
                                    fresh={true}
                                    title={
                                        <div style={{whiteSpace: 'pre-line'}}>
                                            <div style={{
                                                fontWeight: 'bold',
                                                marginBottom: '4px'
                                            }}>容器: {container.name}</div>
                                            {message && <div>{message}</div>}
                                        </div>
                                    }
                                >
                                    <span className='text text-danger font-medium text-sm '
                                          style={{marginRight: '4px', cursor: 'pointer'}}>{reason}</span>
                                </Tooltip>
                            );
                        })}
                    </div>
                )
            }
        </div>
    );
});

export default K8sPodStatusComponent;
