import { HPA } from '@/store/hpa';
import React from 'react';
import { Tooltip } from 'antd';

interface HPAMetricsProps {
    data: HPA;
}

const HPAMetricsComponent = React.forwardRef<HTMLSpanElement, HPAMetricsProps>(({ data }, ref) => {
    const getMetricsData = () => {
        console.log('data', data);
        console.log('data', data.spec);
        console.log('data', data.spec.metrics);

        if (!data?.spec?.metrics) return [];

        return data.spec.metrics.map((metric, index) => {
            let name = '';
            let target = null;
            let current = null;

            if (metric.type === 'Resource' && metric.resource) {
                name = metric.resource.name;
                target = metric.resource.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Resource' && m.resource?.name === name)?.resource?.current;
            } else if (metric.type === 'ContainerResource' && metric.containerResource) {
                name = `${metric.containerResource.container}/${metric.containerResource.name}`;
                target = metric.containerResource.target;
                current = data?.status?.currentMetrics?.find(m =>
                    m.type === 'ContainerResource' &&
                    m.containerResource?.container === metric.containerResource?.container &&
                    m.containerResource?.name === metric.containerResource?.name
                )?.containerResource?.current;
            } else if (metric.type === 'Pods' && metric.pods) {
                name = metric.pods.metric.name;
                target = metric.pods.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Pods' && m.pods?.metric.name === name)?.pods?.current;
            } else if (metric.type === 'External' && metric.external) {
                name = metric.external.metric.name;
                target = metric.external.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'External' && m.external?.metric.name === name)?.external?.current;
            } else if (metric.type === 'Object' && metric.object) {
                name = metric.object.metric.name;
                target = metric.object.target;
                current = data?.status?.currentMetrics?.find(m => m.type === 'Object' && m.object?.metric.name === name)?.object?.current;
            }

            return {
                key: index,
                type: metric.type,
                name,
                target,
                current,
            };
        });
    };

    const formatValue = (value: string | undefined, name: string) => {
        if (!value) return '-';
        if (name === 'cpu') {
            const cores = parseInt(value) / 1000;
            return `${cores}核`;
        } else if (name === 'memory') {
            if (value.endsWith('i')) return value;
            const mi = Math.round(parseInt(value) / (1024 * 1024));
            return `${mi}Mi`;
        }
        return value;
    };

    return (
        <span ref={ref}>
            <div >
                {getMetricsData().map((item, index) => (
                    <div key={index} style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                    }}>
                        <Tooltip title={item.name}>
                            <span style={{ color: '#1890ff' }}>{item.name}</span>
                        </Tooltip>
                        <span>/</span>
                        <span style={{ color: '#1890ff' }}>
                            {item?.target?.type === 'Utilization' && item.target.averageUtilization ?
                                `${item.target.averageUtilization}%` :
                                ((item?.target?.type === 'AverageValue' && item.target.averageValue) || (item?.target?.type === 'Value' && item.target.value)) ?
                                    formatValue(item.target.averageValue || item.target.value, item.name) :
                                    '-'
                            }
                        </span>
                        <span>/</span>
                        <span style={{ color: '#52c41a' }}>
                            {!item.current ? '-' :
                                item.current.averageUtilization ?
                                    `${item.current.averageUtilization}%` :
                                    (item.current.averageValue || item.current.value) ?
                                        formatValue(item.current.averageValue || item.current.value, item.name) :
                                        '-'
                            }
                        </span>
                    </div>
                ))}
            </div>
        </span>
    );
});

export default HPAMetricsComponent;
