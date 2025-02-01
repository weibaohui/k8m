const ShowAnnotationIcon = (annotations: Record<string, string>) => {
    // 如果 annotations 存在且不是空对象，则返回主要颜色图标，否则返回次要颜色图标
    return annotations && Object.keys(annotations).length > 0;
};
export default ShowAnnotationIcon;
