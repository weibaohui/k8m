import React from 'react';

interface NodeRolesProps {
    data: {
        metadata?: {
            labels?: Record<string, string>;
        };
    };
}

// 用 forwardRef 包装组件
const NodeRolesComponent = React.forwardRef<HTMLSpanElement, NodeRolesProps>(({data}, ref) => {
    const labels = data.metadata?.labels || {};
    const roles = Object.keys(labels).filter(label =>
        label.startsWith('node-role.kubernetes.io/')
    );

    const roleMap = {
        'master': '主节点',
        'control-plane': '控制平面',
        'worker': '工作节点',
        'ingress': '入口节点',
        'storage': '存储节点',
        'compute': '计算节点',
        'agent': '代理节点',
    };

    const displayedRoles = roles.map(role => {
        const roleKey = role.replace('node-role.kubernetes.io/', '');
        return roleMap[roleKey as keyof typeof roleMap] || roleKey;
    });

    return (
        <span ref={ref}>
            {displayedRoles.length > 0 ? displayedRoles.join(', ') : ''}
        </span>
    );
});

export default NodeRolesComponent;