export interface HPA {
    apiVersion: 'autoscaling/v2';
    kind: 'HorizontalPodAutoscaler';
    metadata: {
        name: string;
        namespace?: string;
        labels?: { [key: string]: string };
        annotations?: { [key: string]: string };
    };
    spec: {
        scaleTargetRef: {
            apiVersion: string;
            kind: string;
            name: string;
        };
        minReplicas?: number;
        maxReplicas: number;
        metrics: Array<{
            type: 'Resource' | 'Pods' | 'Object' | 'External' | 'ContainerResource';
            resource?: {
                name: string;
                target: {
                    type: 'Utilization' | 'AverageValue' | 'Value';
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            containerResource?: {
                container: string;
                name: string;
                target: {
                    type: 'Utilization' | 'AverageValue' | 'Value';
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            pods?: {
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                target: {
                    type: 'Utilization' | 'AverageValue' | 'Value';
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            external?: {
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                target: {
                    type: 'Utilization' | 'AverageValue' | 'Value';
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            object?: {
                describeObject: {
                    apiVersion: string;
                    kind: string;
                    name: string;
                };
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                target: {
                    type: 'Utilization' | 'AverageValue' | 'Value';
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
        }>;
        behavior?: {
            scaleUp?: {
                policies?: Array<{
                    type: 'Pods' | 'Percent';
                    value: number;
                    periodSeconds: number;
                }>;
                selectPolicy?: 'Max' | 'Min' | 'Disabled';
            };
            scaleDown?: {
                policies?: Array<{
                    type: 'Pods' | 'Percent';
                    value: number;
                    periodSeconds: number;
                }>;
                selectPolicy?: 'Max' | 'Min' | 'Disabled';
            };
        };
    };
    status?: {
        observedGeneration?: number;
        lastScaleTime?: string;
        currentReplicas: number;
        desiredReplicas: number;
        currentMetrics: Array<{
            type: 'Resource' | 'Pods' | 'Object' | 'External' | 'ContainerResource';
            resource?: {
                name: string;
                current: {
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            containerResource?: {
                container: string;
                name: string;
                current: {
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            pods?: {
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                current: {
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            external?: {
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                current: {
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
            object?: {
                describeObject: {
                    apiVersion: string;
                    kind: string;
                    name: string;
                };
                metric: {
                    name: string;
                    selector?: {
                        matchLabels?: { [key: string]: string };
                        matchExpressions?: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    };
                };
                current: {
                    averageUtilization?: number;
                    averageValue?: string;
                    value?: string;
                };
            };
        }>;
        conditions: Array<{
            type: string;
            status: 'True' | 'False' | 'Unknown';
            lastTransitionTime: string;
            reason: string;
            message: string;
        }>;
    };
}
