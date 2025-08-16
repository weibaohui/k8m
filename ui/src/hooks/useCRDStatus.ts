import { useState, useEffect } from 'react';
import { fetcher } from '@/components/Amis/fetcher';

interface CRDSupportedStatus {
    IsGatewayAPISupported: boolean;
    IsOpenKruiseSupported: boolean;
    IsIstioSupported: boolean;
}

export const useCRDStatus = () => {
    const [isGatewayAPISupported, setIsGatewayAPISupported] = useState<boolean>(false);
    const [isOpenKruiseSupported, setIsOpenKruiseSupported] = useState<boolean>(false);
    const [isIstioSupported, setIsIstioSupported] = useState<boolean>(false);

    useEffect(() => {
        const fetchCRDSupportedStatus = async () => {
            try {
                const response = await fetcher({
                    url: '/k8s/crd/status',
                    method: 'get'
                });
                
                if (response.data && typeof response.data === 'object') {
                    const status = response.data.data as CRDSupportedStatus;
                    setIsGatewayAPISupported(status.IsGatewayAPISupported);
                    setIsOpenKruiseSupported(status.IsOpenKruiseSupported);
                    setIsIstioSupported(status.IsIstioSupported);
                }
            } catch (error) {
                console.error('Failed to fetch CRD status:', error);
            }
        };

        fetchCRDSupportedStatus();
    }, []);

    return {
        isGatewayAPISupported,
        isOpenKruiseSupported,
        isIstioSupported
    };
};