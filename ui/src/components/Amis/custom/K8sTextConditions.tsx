import React, { useEffect, useMemo } from 'react';
import { render as renderAmis } from 'amis';

import useConditionsStore from '@/store/conditions';
import { GetValueByPath } from '@/utils/utils';

interface Condition {
    reason: any;
    message: any;
    type: string;
    status: string;
}

interface K8sTextConditionsProps {
    data: any;
}

const K8sTextConditionsComponent = React.forwardRef<HTMLSpanElement, K8sTextConditionsProps>((props, ref) => {
    //@ts-ignore
    const { name, data } = props;

    const { reverseConditions, initialized, initReverseConditions } = useConditionsStore();
    useEffect(() => {
        if (!initialized) {
            initReverseConditions();
        }
    }, [initialized, initReverseConditions]);

    const { allNormal, conditionDetails } = useMemo(() => {

        const conditions: Array<Condition> = GetValueByPath(data, name) || [];


        if (conditions.length === 0) {
            return { allNormal: true, conditionDetails: '' };
        }

        const allNormal = conditions.every(condition => {
            const shouldReverse = reverseConditions.some(rc => condition.type.includes(rc));
            return shouldReverse ?
                condition.status === 'False' :
                condition.status === 'True';
        });

        // 生成 conditions 的 HTML 片段
        const conditionDetails = conditions.map(condition => {
            const shouldReverse = reverseConditions.some(rc => condition.type.includes(rc));
            const isNormal = shouldReverse ?
                condition.status === 'False' :
                condition.status === 'True';

            return `<p>${condition.type}: <strong>${isNormal ?
                '<span class="text-green-500 text-xs">正常</span>' :
                `<span class="text-red-500 text-xs">异常</span>`}</strong>
                ${condition.message ? `<span class='ml-4 text-gray-500 text-xs'>${condition.message}</span>` : ''}
                ${condition.reason ? `<span class='ml-4 text-gray-500 text-xs'>${condition.reason}</span>` : ''}
                </p>`;
        }).join('');

        return { allNormal, conditionDetails };
    }, [data.status?.conditions, reverseConditions]);
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
                    body: <div dangerouslySetInnerHTML={{ __html: conditionDetails }} />
                }
            })}
        </span>
    );
});

export default K8sTextConditionsComponent;
