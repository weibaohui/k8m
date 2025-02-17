import React from 'react';
import { Select } from 'antd';

interface ContainerSelectorProps {
    selectedContainer: string;
    containers: Array<{
        name: string;
    }>;
    onContainerChange: (value: string) => void;
}

const ContainerSelector: React.FC<ContainerSelectorProps> = ({
    selectedContainer,
    containers,
    onContainerChange
}) => {
    const containerOptions = containers.map(container => ({
        label: container.name,
        value: container.name
    }));

    return (
        <Select
            prefix='容器：'
            value={selectedContainer}
            onChange={onContainerChange}
            options={containerOptions}
        />
    );
};

export default ContainerSelector;