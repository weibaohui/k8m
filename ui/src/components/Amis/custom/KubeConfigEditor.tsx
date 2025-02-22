import React from 'react';

interface KubeConfigProps {
    data: Record<string, any>;
}

const KubeConfigEditorComponent = React.forwardRef<HTMLDivElement, KubeConfigProps>(() => {
    return <span></span>;
});

export default KubeConfigEditorComponent;
