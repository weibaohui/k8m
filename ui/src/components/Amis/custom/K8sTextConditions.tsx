import React, { useEffect, useMemo, useState } from 'react';
import { render as renderAmis } from 'amis';
import { fetcher } from '../fetcher';

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

const K8sTextConditionsComponent = React.forwardRef<HTMLSpanElement, K8sTextConditionsProps>(({ data }, ref) => {
    const [reverseConditions, setReverseConditions] = useState<string[]>([]);

    useEffect(() => {
        fetcher({
            url: '/params/condition/reverse/list',
            method: 'get'
        })
            .then(response => {
                if (Array.isArray(response.data?.data)) {
                    setReverseConditions(response.data.data);
                }
            })
            .catch(error => console.error('获取反转条件列表失败:', error));
    }, []);

    const { allNormal, conditionDetails } = useMemo(() => {
        const conditions = data.status?.conditions || [];

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
                '<span class="text-red-500 text-xs">异常</span>'}</strong></p>`;
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
