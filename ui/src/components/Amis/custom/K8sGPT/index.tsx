import React, { useEffect, useState } from 'react';

interface K8sGPTProps {
    data: Record<string, any>; // 泛型数据类型
    name: string;
}

// 用 forwardRef 包装组件
const K8sGPTComponent = React.forwardRef<HTMLDivElement, K8sGPTProps>(({ data, name }, _) => {
    console.log(data);
    console.log(name);
    return <div >x</div>;
});

export default K8sGPTComponent;
