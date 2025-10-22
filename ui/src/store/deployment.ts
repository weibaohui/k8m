import { Metadata, Container } from './pod';

export interface Deployment {
    apiVersion: string;
    kind: string;
    metadata: Metadata;
    spec: DeploymentSpec;
    status?: DeploymentStatus;
}

export interface DeploymentSpec {
    replicas?: number;
    selector: LabelSelector;
    template: PodTemplateSpec;
    strategy?: DeploymentStrategy;
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
    paused?: boolean;
    progressDeadlineSeconds?: number;
}

export interface LabelSelector {
    matchLabels?: { [key: string]: string };
    matchExpressions?: Array<{
        key: string;
        operator: string; // In, NotIn, Exists, DoesNotExist
        values?: string[];
    }>;
}

export interface PodTemplateSpec {
    metadata?: {
        labels?: { [key: string]: string };
        annotations?: { [key: string]: string };
    };
    spec: PodSpec;
}

export interface PodSpec {
    containers: Container[];
    initContainers?: Container[];
    ephemeralContainers?: Container[];
    volumes?: Array<{
        name: string;
        persistentVolumeClaim?: {
            claimName: string;
            readOnly?: boolean;
        };
        configMap?: {
            name: string;
            items?: Array<{
                key: string;
                path: string;
            }>;
        };
        secret?: {
            secretName: string;
            items?: Array<{
                key: string;
                path: string;
            }>;
        };
        emptyDir?: {
            sizeLimit?: string;
        };
        hostPath?: {
            path: string;
            type?: string;
        };
    }>;
    nodeName?: string;
    nodeSelector?: { [key: string]: string };
    serviceAccountName?: string;
    restartPolicy?: string;
    terminationGracePeriodSeconds?: number;
    dnsPolicy?: string;
    hostNetwork?: boolean;
    hostPID?: boolean;
    hostIPC?: boolean;
    imagePullSecrets?: Array<{
        name: string;
    }>;
    affinity?: {
        nodeAffinity?: NodeAffinity;
        podAffinity?: PodAffinity;
        podAntiAffinity?: PodAntiAffinity;
    };
    tolerations?: Array<{
        key?: string;
        operator?: string;
        value?: string;
        effect?: string;
        tolerationSeconds?: number;
    }>;
    securityContext?: {
        runAsUser?: number;
        runAsGroup?: number;
        runAsNonRoot?: boolean;
        fsGroup?: number;
        seLinuxOptions?: {
            level?: string;
            role?: string;
            type?: string;
            user?: string;
        };
    };
}

export interface NodeAffinity {
    requiredDuringSchedulingIgnoredDuringExecution?: {
        nodeSelectorTerms: Array<{
            matchExpressions?: Array<{
                key: string;
                operator: string;
                values?: string[];
            }>;
            matchFields?: Array<{
                key: string;
                operator: string;
                values?: string[];
            }>;
        }>;
    };
    preferredDuringSchedulingIgnoredDuringExecution?: Array<{
        weight: number;
        preference: {
            matchExpressions?: Array<{
                key: string;
                operator: string;
                values?: string[];
            }>;
            matchFields?: Array<{
                key: string;
                operator: string;
                values?: string[];
            }>;
        };
    }>;
}

export interface PodAffinity {
    requiredDuringSchedulingIgnoredDuringExecution?: Array<{
        labelSelector?: LabelSelector;
        namespaces?: string[];
        topologyKey: string;
    }>;
    preferredDuringSchedulingIgnoredDuringExecution?: Array<{
        weight: number;
        podAffinityTerm: {
            labelSelector?: LabelSelector;
            namespaces?: string[];
            topologyKey: string;
        };
    }>;
}

export interface PodAntiAffinity {
    requiredDuringSchedulingIgnoredDuringExecution?: Array<{
        labelSelector?: LabelSelector;
        namespaces?: string[];
        topologyKey: string;
    }>;
    preferredDuringSchedulingIgnoredDuringExecution?: Array<{
        weight: number;
        podAffinityTerm: {
            labelSelector?: LabelSelector;
            namespaces?: string[];
            topologyKey: string;
        };
    }>;
}

export interface DeploymentStrategy {
    type?: string; // RollingUpdate, Recreate
    rollingUpdate?: {
        maxUnavailable?: string | number;
        maxSurge?: string | number;
    };
}

export interface DeploymentStatus {
    observedGeneration?: number;
    replicas?: number;
    updatedReplicas?: number;
    readyReplicas?: number;
    availableReplicas?: number;
    unavailableReplicas?: number;
    conditions?: Array<{
        type: string;
        status: string;
        lastUpdateTime?: string;
        lastTransitionTime?: string;
        reason?: string;
        message?: string;
    }>;
    collisionCount?: number;
}

// 批量操作相关的接口
export interface BatchUpdateImageRequest {
    deployments: Array<{
        name: string;
        namespace: string;
        containers: Array<{
            name: string;
            image: string;
        }>;
    }>;
}

export interface BatchOperationResult {
    success: boolean;
    message: string;
    results: Array<{
        name: string;
        namespace: string;
        success: boolean;
        message: string;
    }>;
}