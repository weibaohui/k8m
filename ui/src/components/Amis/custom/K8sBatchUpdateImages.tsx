import React, { useCallback, useState } from 'react';
import * as yaml from 'js-yaml';
import { Button, message } from 'antd';
import { fetcher } from '@/components/Amis/fetcher';
import { dataMapping } from 'amis-core';
import { Deployment } from '@/store/deployment';

interface ImageBatchUpdateProps {
    data: Record<string, any>;
}

const ImageBatchUpdateComponent = React.forwardRef<HTMLDivElement, ImageBatchUpdateProps>(
    ({ data }, _) => {
        // selectedItems 是一个 Deployment 类型的数组
        const selectedItems: Deployment[] = data.selectedItems || [];
        console.log('Selected Deployments:', selectedItems);

        // 示例：访问 selectedItems 中的 Deployment 信息
        const handleBatchUpdate = useCallback(() => {
            selectedItems.forEach((deployment: Deployment) => {
                console.log('Deployment Name:', deployment.metadata.name);
                console.log('Namespace:', deployment.metadata.namespace);
                console.log('Current Replicas:', deployment.spec.replicas);
                console.log('Containers:', deployment.spec.template.spec.containers);
                
                // 访问容器镜像信息
                deployment.spec.template.spec.containers.forEach(container => {
                    console.log(`Container ${container.name} image: ${container.image}`);
                });
            });
        }, [selectedItems]);

        return (
            <div>
                <p>选中的 Deployment 数量: {selectedItems.length}</p>
                {selectedItems.length > 0 && (
                    <div>
                        <h4>选中的 Deployments:</h4>
                        <ul>
                            {selectedItems.map((deployment, index) => (
                                <li key={`${deployment.metadata.namespace}-${deployment.metadata.name}-${index}`}>
                                    {deployment.metadata.namespace}/{deployment.metadata.name} 
                                    (副本数: {deployment.spec.replicas})
                                </li>
                            ))}
                        </ul>
                        <Button onClick={handleBatchUpdate}>
                            处理选中的 Deployments
                        </Button>
                    </div>
                )}
            </div>
        );
    }
);

export default ImageBatchUpdateComponent;
