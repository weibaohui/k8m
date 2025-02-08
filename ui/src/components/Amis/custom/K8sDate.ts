const K8sDate = (str: string) => {
    // 转换为本地时间
    const formatLocalTime = (utcTime?: string) => {
        if (!utcTime) return '未知时间';
        const date = new Date(utcTime);
        return date.toLocaleString(); // 本地格式
    };
    return formatLocalTime(str);

};
export default K8sDate;
