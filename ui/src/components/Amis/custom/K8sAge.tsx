import React, {useEffect, useState} from 'react';

interface K8sAgeDisplayProps {
    data: Record<string, any>; // 泛型数据类型
    name: string; // 指定获取 `creationTimestamp` 的路径
}

// 用 forwardRef 包装组件
const K8sAgeDisplayComponent = React.forwardRef<HTMLSpanElement, K8sAgeDisplayProps>(({data, name}, ref) => {
    const [currentTime, setCurrentTime] = useState(Date.now());

    // 获取嵌套对象值的辅助函数
    const getValueByPath = (obj: Record<string, any>, path: string): any => {
        return path.split('.').reduce((acc, part) => acc && acc[part], obj);
    };

    // 动态获取 creationTimestamp
    const creationTimestamp = getValueByPath(data, name);

    // 每秒更新当前时间
    useEffect(() => {
        const interval = setInterval(() => {
            setCurrentTime(Date.now());
        }, 1000); // 每秒更新一次

        return () => clearInterval(interval); // 组件卸载时清除定时器
    }, []);

    // 当 props 变化时，手动更新当前时间
    useEffect(() => {
        setCurrentTime(Date.now());
    }, [data]);

    // 格式化时间的函数
    const formatHumanDuration = (durationInMs: number): string => {
        const seconds = Math.floor(durationInMs / 1000);
        if (seconds < 0) return "0s";
        if (seconds < 120) return `${seconds}s`;

        const minutes = Math.floor(seconds / 60);
        if (minutes < 10) {
            const s = seconds % 60;
            return s === 0 ? `${minutes}m` : `${minutes}m${s}s`;
        }
        if (minutes < 180) return `${minutes}m`;

        const hours = Math.floor(minutes / 60);
        if (hours < 8) {
            const m = minutes % 60;
            return m === 0 ? `${hours}h` : `${hours}h${m}m`;
        }
        if (hours < 48) return `${hours}h`;

        const days = Math.floor(hours / 24);
        if (days < 8) {
            const h = hours % 24;
            return h === 0 ? `${days}d` : `${days}d${h}h`;
        }
        if (days < 730) return `${days}d`; // 2 年以内

        const years = Math.floor(days / 365);
        const dy = days % 365;
        return dy === 0 ? `${years}y` : `${years}y${dy}d`;
    };

    // 如果 creationTimestamp 无效，返回 "N/A"
    if (!creationTimestamp) {
        return <span ref={ref}>N/A</span>;
    }

    // 计算时间差
    const durationInMs = currentTime - new Date(creationTimestamp).getTime();
    const formattedTime = formatHumanDuration(durationInMs);

    // 显示格式化的时间
    return <span ref={ref}>{formattedTime}</span>;
});

export default K8sAgeDisplayComponent;
