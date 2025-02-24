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
            lastTransitionTime?: string;
        }>;
        containerStatuses?: Array<{
            name: string;
            ready: boolean;
            restartCount?: number;
            image?: string;
            imageID?: string;
            state?: {
                waiting?: {
                    reason?: string;
                    message?: string;
                };
                running?: {
                    startedAt?: string;
                };
                terminated?: {
                    reason?: string;
                    message?: string;
                    exitCode?: number;
                    startedAt?: string;
                    finishedAt?: string;
                };
            };
        }>;
        hostIP?: string;
        podIP?: string;
        startTime?: string;
        qosClass?: string;
    };
}

export interface Metadata {
    name: string;
    namespace: string;
    labels?: { [key: string]: string };
    annotations?: { [key: string]: string };
    uid?: string;
    resourceVersion?: string;
    creationTimestamp?: string;
    deletionTimestamp?: string;
    generateName?: string;
    ownerReferences?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        uid: string;
        controller?: boolean;
        blockOwnerDeletion?: boolean;
    }>;
}

export interface Spec {
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
        // 其他卷类型可以根据需要添加
    }>;
    nodeName?: string;
    serviceAccountName?: string;
    restartPolicy?: string;
    terminationGracePeriodSeconds?: number;
    dnsPolicy?: string;
    hostNetwork?: boolean;
    hostPID?: boolean;
    hostIPC?: boolean;
}

export interface Container {
    name: string;
    image: string;
    command?: string[];
    args?: string[];
    env?: Array<{
        name: string;
        value?: string;
        valueFrom?: {
            configMapKeyRef?: {
                name: string;
                key: string;
            };
            secretKeyRef?: {
                name: string;
                key: string;
            };
            fieldRef?: {
                fieldPath: string;
            };
        };
    }>;
    resources?: {
        limits?: {
            cpu?: string;
            memory?: string;
        };
        requests?: {
            cpu?: string;
            memory?: string;
        };
    };
    volumeMounts?: Array<{
        name: string;
        mountPath: string;
        readOnly?: boolean;
        subPath?: string;
    }>;
    ports?: Array<{
        containerPort: number;
        protocol?: string;
        hostPort?: number;
    }>;
    livenessProbe?: {
        httpGet?: {
            path: string;
            port: number;
            scheme?: string;
        };
        tcpSocket?: {
            port: number;
        };
        exec?: {
            command: string[];
        };
        initialDelaySeconds?: number;
        periodSeconds?: number;
        timeoutSeconds?: number;
        successThreshold?: number;
        failureThreshold?: number;
    };
    readinessProbe?: {
        httpGet?: {
            path: string;
            port: number;
            scheme?: string;
        };
        tcpSocket?: {
            port: number;
        };
        exec?: {
            command: string[];
        };
        initialDelaySeconds?: number;
        periodSeconds?: number;
        timeoutSeconds?: number;
        successThreshold?: number;
        failureThreshold?: number;
    };
    imagePullPolicy?: string;
    securityContext?: {
        privileged?: boolean;
        runAsUser?: number;
        runAsGroup?: number;
        readOnlyRootFilesystem?: boolean;
        allowPrivilegeEscalation?: boolean;
    };
}
