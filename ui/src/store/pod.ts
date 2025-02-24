export interface Pod {
    metadata: Metadata;
    spec: Spec;
    status?: {
        phase?: string;
        conditions?: Array<{
            type: string;
            status: string;
            message?: string;
            reason?: string;
        }>;
        containerStatuses?: {
            name: string;
            ready: boolean;
            state?: {
                waiting?: {
                    reason?: string;
                    message?: string;
                };
                terminated?: {
                    reason?: string;
                    message?: string;
                };
            };
        }[];
    };
}

export interface Metadata {
    name: string;
    namespace: string;
}

export interface Spec {
    containers: Container[];
    initContainers: Container[];
    ephemeralContainers: Container[];
}

export interface Container {
    name: string;
    image: string;
    command: string[];
    args: string[];
}
