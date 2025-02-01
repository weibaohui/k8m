import React, {useMemo} from 'react';
import {render as renderAmis} from 'amis';

// 定义 Props 类型
interface K8sTextConditionsProps {
    data: {
        status?: {
            conditions?: Array<{
                type: string;
                status: string;
            }>;
        };
    };
}

const K8sTextConditionsComponent = React.forwardRef<HTMLSpanElement, K8sTextConditionsProps>(({data}, ref) => {
    const {allNormal, conditionDetails} = useMemo(() => {
        const conditions = data.status?.conditions || [];

        if (conditions.length === 0) {
            return {allNormal: true, conditionDetails: ''};
        }

        const allNormal = conditions.every(condition =>
            condition.status === 'True' ||
            (condition.status === 'False' && (condition.type.includes('Pressure') || condition.type.includes("Unavailable")))
        );

        // 生成 conditions 的 HTML 片段
        const conditionDetails = conditions.map(condition => (
            `<p>${condition.type}: <strong>${condition.status === 'True' ||
            (condition.status === 'False' && (condition.type.includes('Pressure') || condition.type.includes("Unavailable"))) ?
                '<span class="text-green-500 text-xs">正常</span>' :
                '<span class="text-red-500 text-xs">异常</span>'}</strong></p>`
        )).join('');

        return {allNormal, conditionDetails};
    }, [data.status?.conditions]);
    // 确定状态文本和按钮颜色
    const statusText = allNormal ? '就绪' : '未就绪';
    const level = allNormal ? 'primary' : 'danger';

    return (
        <span ref={ref}>
            {renderAmis({
                type: "button",
                size: "xs",
                label: statusText,
                level: level,
                actionType: "dialog",
                dialog: {
                    closeOnEsc: true,
                    closeOnOutside: true,
                    title: "条件状态 (ESC 关闭)",
                    size: "md",
                    body: <div dangerouslySetInnerHTML={{__html: conditionDetails}}/>
                }
            })}
        </span>
    );
});

export default K8sTextConditionsComponent;
