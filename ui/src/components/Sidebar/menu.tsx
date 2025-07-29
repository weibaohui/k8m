import {useNavigate} from "react-router-dom";
import type {MenuProps} from 'antd';
import {useEffect, useState} from 'react';
import {fetcher} from '../Amis/fetcher';

// 定义用户角色接口
interface UserRoleResponse {
    role: string;  // 根据实际数据结构调整类型
    cluster: string;
}

interface CRDSupportedStatus {
    IsGatewayAPISupported: boolean;
    IsOpenKruiseSupported: boolean;
    IsIstioSupported: boolean;
}

type MenuItem = Required<MenuProps>['items'][number];

const items: () => MenuItem[] = () => {
    const navigate = useNavigate()
    const [userRole, setUserRole] = useState<string>('');
    const [isGatewayAPISupported, setIsGatewayAPISupported] = useState<boolean>(false);
    const [isOpenKruiseSupported, setIsOpenKruiseSupported] = useState<boolean>(false);
    const [isIstioSupported, setIsIstioSupported] = useState<boolean>(false);

    useEffect(() => {
        const fetchUserRole = async () => {
            try {
                const response = await fetcher({
                    url: '/params/user/role',
                    method: 'get'
                });
                // 检查 response.data 是否存在，并确保其类型正确
                if (response.data && typeof response.data === 'object') {
                    const role = response.data.data as UserRoleResponse;
                    setUserRole(role.role);

                    const originCluster = localStorage.getItem('cluster') || '';
                    if (originCluster == "" && role.cluster != "") {
                        localStorage.setItem('cluster', role.cluster);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch user role:', error);
            }
        };

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
                console.error('Failed to fetch Gateway API status:', error);
            }
        };


        fetchUserRole();
        fetchCRDSupportedStatus();
    }, []);

    const loadJsonPage = (path: string) => {
        navigate(path)
    }
    return [
        {
            label: "多集群",
            title: "多集群",
            icon: <i className="fa-solid fa-server"></i>,
            key: "cluster_user",
            onClick: () => loadJsonPage('/user/cluster/cluster_user')
        },
        {
            label: "集群总览",
            title: "集群总览",
            icon: <i className="fa-solid fa-chart-pie"></i>,
            key: "cluster_summary",
            onClick: () => loadJsonPage('/cluster/summary_view')
        },
        {
            label: "命名空间",
            title: "命名空间",
            icon: <i className="fa-solid fa-border-style"></i>,
            key: "cluster_ns",
            onClick: () => loadJsonPage('/cluster/ns')
        },
        {
            label: "节点",
            title: "节点",
            icon: <i className="fa-solid fa-computer"></i>,
            key: "cluster_node",
            onClick: () => loadJsonPage('/cluster/node')
        },
        {
            label: "事件",
            title: "事件",
            icon: <i className="fa-solid fa-bell"></i>,
            key: "event",
            onClick: () => loadJsonPage('/ns/event')
        },

        {
            label: "工作负载",
            title: "工作负载",
            icon: <i className="fa-solid fa-cube"></i>,
            key: "workload",
            children: [
                {
                    label: "部署",
                    title: "部署",
                    icon: <i className="fa-solid fa-layer-group"></i>,
                    key: "deploy",
                    onClick: () => loadJsonPage('/ns/deploy')
                },
                {
                    label: "有状态集",
                    title: "有状态集",
                    icon: <i className="fa-solid fa-database"></i>,
                    key: "statefulset",
                    onClick: () => loadJsonPage('/ns/statefulset')
                },
                {
                    label: "守护进程集",
                    title: "守护进程集",
                    icon: <i className="fa-solid fa-shield-halved"></i>,
                    key: "daemonset",
                    onClick: () => loadJsonPage('/ns/daemonset')
                },
                {
                    label: "任务",
                    title: "任务",
                    icon: <i className="fa-solid fa-list-check"></i>,
                    key: "job",
                    onClick: () => loadJsonPage('/ns/job')
                },
                {
                    label: "定时任务",
                    title: "定时任务",
                    icon: <i className="fa-solid fa-clock"></i>,
                    key: "cronjob",
                    onClick: () => loadJsonPage('/ns/cronjob')
                },
                {
                    label: "容器组",
                    title: "容器组",
                    icon: <i className="fa-solid fa-cubes"></i>,
                    key: "pod",
                    onClick: () => loadJsonPage('/ns/pod')
                },
                {
                    label: "副本集",
                    title: "副本集",
                    icon: <i className="fa-solid fa-clone"></i>,
                    key: "replicaset",
                    onClick: () => loadJsonPage('/ns/replicaset')
                },
            ],
        },
        {
            label: "CRD",
            icon: <i className="fa-solid fa-file-code"></i>,
            key: "crd",
            children: [
                {
                    label: "自定义资源",
                    icon: <i className="fa-solid fa-gears"></i>,
                    key: "custom_resource",
                    onClick: () => loadJsonPage('/crd/crd')
                }
            ],
        },
        ...(isOpenKruiseSupported ? [
            {
                label: "OpenKruise",
                title: "OpenKruise",
                icon: <i className="fa-solid fa-cube"></i>,
                key: "OpenKruise-workload",
                children: [
                    {
                        label: "克隆集",
                        title: "克隆集",
                        icon: <i className="fa-solid fa-clone"></i>,
                        key: "advanced-cloneset",
                        onClick: () => loadJsonPage('/openkruise/cloneset')
                    }, {
                        label: "高级有状态集",
                        title: "高级有状态集",
                        icon: <i className="fa-solid fa-layer-group"></i>,
                        key: "advanced-statefulset",
                        onClick: () => loadJsonPage('/openkruise/statefulset')
                    },
                    {
                        label: "高级守护进程集",
                        title: "高级守护进程集",
                        icon: <i className="fa-solid fa-shield-halved"></i>,
                        key: "advanced-daemonSet",
                        onClick: () => loadJsonPage('/openkruise/daemonset')
                    },
                    {
                        label: "高级定时任务",
                        title: "高级定时任务",
                        icon: <i className="fa-solid fa-clock"></i>,
                        key: "advanced-cronJob",
                        onClick: () => loadJsonPage('/openkruise/cronjob')
                    },
                    {
                        label: "广播作业任务",
                        title: "广播作业任务",
                        icon: <i className="fa-solid fa-broadcast-tower"></i>,
                        key: "broadcast-job",
                        onClick: () => loadJsonPage('/openkruise/broadcastjob')
                    },
                    {
                        label: "边车集",
                        title: "边车集",
                        icon: <i className="fa-solid fa-car-side"></i>,
                        key: "sidecarset",
                        onClick: () => loadJsonPage('/openkruise/sidecarset')
                    },
                    {
                        label: "工作负载分布",
                        title: "工作负载分布",
                        icon: <i className="fa-solid fa-arrows-split-up-and-left"></i>,
                        key: "workload-spread",
                        onClick: () => loadJsonPage('/openkruise/workloadspread')
                    },
                    {
                        label: "联合部署",
                        title: "联合部署",
                        icon: <i className="fa-solid fa-object-group"></i>,
                        key: "united-deployment",
                        onClick: () => loadJsonPage('/openkruise/uniteddeployment')
                    },
                    {
                        label: "容器重建请求",
                        title: "容器重建请求",
                        icon: <i className="fa-solid fa-recycle"></i>,
                        key: "container_recreate_request",
                        onClick: () => loadJsonPage('/openkruise/container_recreate_request')
                    },
                    {
                        label: "镜像拉取作业",
                        title: "镜像拉取作业",
                        icon: <i className="fa-solid fa-cloud-arrow-down"></i>,
                        key: "imagepulljob",
                        onClick: () => loadJsonPage('/openkruise/imagepulljob')
                    },
                    {
                        label: "持久化状态",
                        title: "持久化状态",
                        icon: <i className="fa-solid fa-database"></i>,
                        key: "persistentpodstate",
                        onClick: () => loadJsonPage('/openkruise/persistentpodstate')
                    }, {
                        label: "Pod探测标记",
                        title: "Pod探测标记",
                        icon: <i className="fa-solid fa-magnifying-glass"></i>,
                        key: "podprobemarker",
                        onClick: () => loadJsonPage('/openkruise/podprobemarker')
                    },
                    {
                        label: "Pod不可用预算",
                        title: "Pod不可用预算",
                        icon: <i className="fa-solid fa-circle-exclamation"></i>,
                        key: "PodUnavailableBudget",
                        onClick: () => loadJsonPage('/openkruise/PodUnavailableBudget')
                    },
                    {
                        label: "资源分发",
                        title: "资源分发",
                        icon: <i className="fa-solid fa-share-nodes"></i>,
                        key: "ResourceDistribution",
                        onClick: () => loadJsonPage('/openkruise/ResourceDistribution')
                    },

                ],
            },
        ] : []),
        ...(isIstioSupported ? [
            {
                label: "Istio",
                title: "Istio",
                icon: <i className="fa-solid fa-cube"></i>,
                key: "istio",
                children: [
                    {
                        label: "虚拟服务",
                        title: "VirtualService",
                        icon: <i className="fa-solid fa-route"></i>,
                        key: "isito-VirtualService",
                        onClick: () => loadJsonPage('/istio/VirtualService')
                    },
                    {
                        label: "目标规则",
                        title: "DestinationRule",
                        icon: <i className="fa-solid fa-location-dot"></i>,
                        key: "istio-DestinationRule",
                        onClick: () => loadJsonPage('/istio/DestinationRule')
                    },
                    {
                        label: "Envoy过滤器",
                        title: "EnvoyFilter",
                        icon: <i className="fa-solid fa-filter"></i>,
                        key: "istio-EnvoyFilter",
                        onClick: () => loadJsonPage('/istio/EnvoyFilter')
                    },
                    {
                        label: "网关",
                        title: "Gateway",
                        icon: <i className="fa-solid fa-network-wired"></i>,
                        key: "istio-Gateway",
                        onClick: () => loadJsonPage('/istio/Gateway')
                    },
                    {
                        label: "对等认证",
                        title: "PeerAuthentication",
                        icon: <i className="fa-solid fa-user-shield"></i>,
                        key: "istio-PeerAuthentication",
                        onClick: () => loadJsonPage('/istio/PeerAuthentication')
                    },
                    {
                        label: "代理配置",
                        title: "ProxyConfig",
                        icon: <i className="fa-solid fa-gears"></i>,
                        key: "istio-ProxyConfig",
                        onClick: () => loadJsonPage('/istio/ProxyConfig')
                    },
                    {
                        label: "请求认证",
                        title: "RequestAuthentication",
                        icon: <i className="fa-solid fa-key"></i>,
                        key: "istio-RequestAuthentication",
                        onClick: () => loadJsonPage('/istio/RequestAuthentication')
                    },
                    {
                        label: "服务入口",
                        title: "ServiceEntry",
                        icon: <i className="fa-solid fa-door-open"></i>,
                        key: "istio-ServiceEntry",
                        onClick: () => loadJsonPage('/istio/ServiceEntry')
                    },
                    {
                        label: "边车",
                        title: "Sidecar",
                        icon: <i className="fa-solid fa-car-side"></i>,
                        key: "istio-Sidecar",
                        onClick: () => loadJsonPage('/istio/Sidecar')
                    },
                    {
                        label: "遥测",
                        title: "Telemetry",
                        icon: <i className="fa-solid fa-chart-line"></i>,
                        key: "istio-Telemetry",
                        onClick: () => loadJsonPage('/istio/Telemetry')
                    },
                    {
                        label: "授权策略",
                        title: "AuthorizationPolicy",
                        icon: <i className="fa-solid fa-user-lock"></i>,
                        key: "istio-AuthorizationPolicy",
                        onClick: () => loadJsonPage('/istio/AuthorizationPolicy')
                    },
                    {
                        label: "Wasm插件",
                        title: "WasmPlugin",
                        icon: <i className="fa-solid fa-puzzle-piece"></i>,
                        key: "istio-WasmPlugin",
                        onClick: () => loadJsonPage('/istio/WasmPlugin')
                    },
                    {
                        label: "工作负载条目",
                        title: "WorkloadEntry",
                        icon: <i className="fa-solid fa-server"></i>,
                        key: "istio-WorkloadEntry",
                        onClick: () => loadJsonPage('/istio/WorkloadEntry')
                    },
                    {
                        label: "工作负载组",
                        title: "WorkloadGroup",
                        icon: <i className="fa-solid fa-people-group"></i>,
                        key: "istio-WorkloadGroup",
                        onClick: () => loadJsonPage('/istio/WorkloadGroup')
                    }
                ],
            },
        ] : []),

        {
            label: "Helm应用",
            title: "Helm应用",
            icon: <i className="fab fa-app-store"></i>,
            key: "Helm",
            children: [
                {
                    label: "仓库",
                    title: "仓库",
                    icon: <i className="fas fa-database"></i>,
                    key: "helm_repo",
                    onClick: () => loadJsonPage('/helm/repo')
                },
                {
                    label: "应用包",
                    title: "应用包",
                    icon: <i className="fa-solid fa-cubes"></i>,
                    key: "helm_chart",
                    onClick: () => loadJsonPage('/helm/chart')
                },
                {
                    label: "应用实例",
                    title: "应用实例",
                    icon: <i className="fas fa-layer-group"></i>,
                    key: "helm_release",
                    onClick: () => loadJsonPage('/helm/release')
                }
            ]
        },
        {
            label: "配置",
            icon: <i className="fa-solid fa-sliders-h"></i>,
            key: "config",
            children: [
                {
                    label: "配置映射",
                    icon: <i className="fa-solid fa-map"></i>,
                    key: "configmap",
                    onClick: () => loadJsonPage('/ns/configmap')
                },
                {
                    label: "密钥",
                    icon: <i className="fa-solid fa-key"></i>,
                    key: "secret",
                    onClick: () => loadJsonPage('/ns/secret')
                },
                {
                    label: "验证钩子",
                    icon: <i className="fa-solid fa-check"></i>,
                    key: "validation_webhook",
                    onClick: () => loadJsonPage('/cluster/validation_webhook')
                },
                {
                    label: "变更钩子",
                    icon: <i className="fa-solid fa-exchange"></i>,
                    key: "mutating_webhook",
                    onClick: () => loadJsonPage('/cluster/mutating_webhook')
                },
            ],
        },
        {
            label: "网络",
            icon: <i className="fa-solid fa-network-wired"></i>,
            key: "network",
            children: [
                {
                    label: "SVC服务",
                    icon: <i className="fa-solid fa-project-diagram"></i>,
                    key: "svc",
                    onClick: () => loadJsonPage('/ns/svc')
                },
                {
                    label: "Ingress入口",
                    icon: <i className="fa-solid fa-wifi"></i>,
                    key: "ingress",
                    onClick: () => loadJsonPage('/ns/ing')
                },
                {
                    label: "Ingress入口类",
                    icon: <i className="fa-solid fa-sitemap"></i>,
                    key: "ingress_class",
                    onClick: () => loadJsonPage('/cluster/ingress_class')
                },
            ],
        },
        ...(isGatewayAPISupported ? [
            {
                label: "网关API",
                icon: <i className="fa-solid fa-door-closed"></i>,
                key: "GatewayAPI",
                children: [
                    {
                        label: "网关类",
                        icon: <i className="fa-solid fa-door-open"></i>,
                        key: "gatewayapi_gateway_class",
                        onClick: () => loadJsonPage('/gatewayapi/gateway_class')
                    },
                    {
                        label: "网关",
                        icon: <i className="fa-solid fa-archway"></i>,
                        key: "gatewayapi_gateway",
                        onClick: () => loadJsonPage('/gatewayapi/gateway')
                    },
                    {
                        label: "HTTP路由",
                        icon: <i className="fa-solid fa-route"></i>,
                        key: "gatewayapi_http_route",
                        onClick: () => loadJsonPage('/gatewayapi/http_route')
                    },
                    {
                        label: "GRPC路由",
                        icon: <i className="fa-solid fa-code-branch"></i>,
                        key: "gatewayapi_grpc_route",
                        onClick: () => loadJsonPage('/gatewayapi/grpc_route')
                    },
                    {
                        label: "TCP路由",
                        icon: <i className="fa-solid fa-plug"></i>,
                        key: "gatewayapi_tcp_route",
                        onClick: () => loadJsonPage('/gatewayapi/tcp_route')
                    },
                    {
                        label: "UDP路由",
                        icon: <i className="fa-solid fa-broadcast-tower"></i>,
                        key: "gatewayapi_udp_route",
                        onClick: () => loadJsonPage('/gatewayapi/udp_route')
                    },
                    {
                        label: "TLS路由",
                        icon: <i className="fa-solid fa-shield-alt"></i>,
                        key: "gatewayapi_tls_route",
                        onClick: () => loadJsonPage('/gatewayapi/tls_route')
                    },
                ],
            },
        ] : []),
        {
            label: "存储",
            icon: <i className="fa-solid fa-memory"></i>,
            key: "storage",
            children: [
                {
                    label: "持久卷声明",
                    icon: <i className="fa-solid fa-folder"></i>,
                    key: "pvc",
                    onClick: () => loadJsonPage('/ns/pvc')
                },
                {
                    label: "持久卷",
                    icon: <i className="fa-solid fa-hdd"></i>,
                    key: "pv",
                    onClick: () => loadJsonPage('/cluster/pv')
                },
                {
                    label: "存储类",
                    icon: <i className="fa-solid fa-coins"></i>,
                    key: "storage_class",
                    onClick: () => loadJsonPage('/cluster/storage_class')
                },
            ],
        },
        {
            label: "访问控制",
            icon: <i className="fa-solid fa-lock"></i>,
            key: "access_control",
            children: [
                {
                    label: "服务账户",
                    icon: <i className="fa-solid fa-user-shield"></i>,
                    key: "service_account",
                    onClick: () => loadJsonPage('/ns/service_account')
                },
                {
                    label: "角色",
                    icon: <i className="fa-solid fa-user-tag"></i>,
                    key: "role",
                    onClick: () => loadJsonPage('/ns/role')
                },
                {
                    label: "角色绑定",
                    icon: <i className="fa-solid fa-link"></i>,
                    key: "role_binding",
                    onClick: () => loadJsonPage('/ns/role_binding')
                },
                {
                    label: "集群角色",
                    icon: <i className="fa-solid fa-users"></i>,
                    key: "cluster_role",
                    onClick: () => loadJsonPage('/cluster/cluster_role')
                },
                {
                    label: "集群角色绑定",
                    icon: <i className="fa-solid fa-user-lock"></i>,
                    key: "cluster_role_binding",
                    onClick: () => loadJsonPage('/cluster/cluster_role_binding')
                },
            ],
        },
        {
            label: "集群配置",
            icon: <i className="fa-solid fa-cog"></i>,
            key: "cluster_config",
            children: [


                {
                    label: "端点",
                    icon: <i className="fa-solid fa-ethernet"></i>,
                    key: "endpoint",
                    onClick: () => loadJsonPage('/ns/endpoint')
                },
                {
                    label: "端点切片",
                    icon: <i className="fa-solid fa-newspaper"></i>,
                    key: "endpointslice",
                    onClick: () => loadJsonPage('/ns/endpointslice')
                },
                {
                    label: "水平自动扩缩",
                    icon: <i className="fa-solid fa-arrows-left-right"></i>,
                    key: "hpa",
                    onClick: () => loadJsonPage('/ns/hpa')
                },
                {
                    label: "网络策略",
                    icon: <i className="fa-solid fa-project-diagram"></i>,
                    key: "network_policy",
                    onClick: () => loadJsonPage('/ns/network_policy')
                },
                {
                    label: "资源配额",
                    icon: <i className="fa-solid fa-chart-pie"></i>,
                    key: "resource_quota",
                    onClick: () => loadJsonPage('/ns/resource_quota')
                },
                {
                    label: "限制范围",
                    icon: <i className="fa-solid fa-compress"></i>,
                    key: "limit_range",
                    onClick: () => loadJsonPage('/ns/limit_range')
                },
                {
                    label: "Pod中断配置",
                    icon: <i className="fa-solid fa-receipt"></i>,
                    key: "pdb",
                    onClick: () => loadJsonPage('/ns/pdb')
                },
                {
                    label: "租约",
                    icon: <i className="fa-solid fa-file-contract"></i>,
                    key: "lease",
                    onClick: () => loadJsonPage('/ns/lease')
                },
                {
                    label: "优先级类",
                    icon: <i className="fa-solid fa-sort"></i>,
                    key: "priority_class",
                    onClick: () => loadJsonPage('/cluster/priority_class')
                },
                {
                    label: "运行时类",
                    icon: <i className="fa-solid fa-play"></i>,
                    key: "runtime_class",
                    onClick: () => loadJsonPage('/cluster/runtime_class')
                },
                {
                    label: "CSI节点",
                    icon: <i className="fa-solid fa-server"></i>,
                    key: "csi_node",
                    onClick: () => loadJsonPage('/cluster/csi_node')
                },
                {
                    label: "API 服务",
                    icon: <i className="fa-solid fa-code"></i>,
                    key: "api_service",
                    onClick: () => loadJsonPage('/cluster/api_service')
                },
                {
                    label: "流量规则",
                    icon: <i className="fa-solid fa-random"></i>,
                    key: "flow_schema",
                    onClick: () => loadJsonPage('/cluster/flow_schema')
                },
                {
                    label: "优先级配置",
                    icon: <i className="fa-solid fa-sliders"></i>,
                    key: "priority_level_config",
                    onClick: () => loadJsonPage('/cluster/priority_level_config')
                },
                {
                    label: "组件状态",
                    icon: <i className="fa-solid fa-info-circle"></i>,
                    key: "component_status",
                    onClick: () => loadJsonPage('/cluster/component_status')
                },
            ],
        },
        ...(userRole === 'platform_admin' ? [
            {
                label: "平台设置",
                icon: <i className="fa-solid fa-wrench"></i>,
                key: "platform_settings",
                children: [
                    {
                        label: "多集群管理",
                        title: "多集群管理",
                        icon: <i className="fa-solid fa-server"></i>,
                        key: "cluster_all",
                        onClick: () => loadJsonPage('/admin/cluster/cluster_all')
                    },
                    {
                        label: "参数设置",
                        icon: <i className="fa-solid fa-sliders"></i>,
                        key: "system_config",
                        onClick: () => loadJsonPage('/admin/config/config')
                    },
                    {
                        label: "集群巡检设置",
                        icon: <i className="fa-solid fa-stethoscope"></i>,
                        key: "inspection_settings",
                        children: [
                            {
                                label: "巡检汇总",
                                icon: <i className="fa-solid fa-clipboard-list"></i>,
                                key: "inspection_summary",
                                onClick: () => loadJsonPage('/admin/inspection/summary'),

                            },
                            {
                                label: "巡检计划",
                                icon: <i className="fa-solid fa-stethoscope"></i>,
                                key: "inspection_schedule",
                                onClick: () => loadJsonPage('/admin/inspection/schedule'),

                            },
                            {
                                label: "巡检记录",
                                icon: <i className="fa-solid fa-clipboard-list"></i>,
                                key: "inspection_record",
                                onClick: () => loadJsonPage('/admin/inspection/record')
                            },
                            {
                                label: "巡检规则",
                                icon: <i className="fa-solid fa-clipboard-check"></i>,
                                key: "script_management",
                                onClick: () => loadJsonPage('/admin/inspection/script')
                            },
                            {
                                label: "webhook管理",
                                icon: <i className="fa-solid fa-bell-concierge"></i>,
                                key: "webhook_management",
                                onClick: () => loadJsonPage('/admin/inspection/webhook')
                            },
                        ]
                    },

                    {
                        label: "AI模型配置",
                        icon: <i className="fa-solid fa-sliders"></i>,
                        key: "ai_model_config",
                        onClick: () => loadJsonPage('/admin/config/ai_model_config')
                    },
                    {
                        label: "用户管理",
                        icon: <i className="fa-solid fa-user-gear"></i>,
                        key: "user_management",
                        onClick: () => loadJsonPage('/admin/user/user')
                    },
                    {
                        label: "用户组管理",
                        icon: <i className="fa-solid fa-users-gear"></i>,
                        key: "user_group_management",
                        onClick: () => loadJsonPage('/admin/user/user_group')
                    },
                    {
                        label: "MCP管理",
                        icon: <i className="fa-solid fa-server"></i>,
                        key: "mcp_management",
                        onClick: () => loadJsonPage('/admin/mcp/mcp')
                    },
                    {
                        label: "MCP执行记录",
                        icon: <i className="fa-solid fa-history"></i>,
                        key: "mcp_tool_log",
                        onClick: () => loadJsonPage('/admin/mcp/mcp_log')
                    },
                    {
                        label: "指标显示翻转",
                        icon: <i className="fa-solid fa-arrows-rotate"></i>,
                        key: "condition_reverse",
                        onClick: () => loadJsonPage('/admin/config/condition')
                    },
                    {
                        label: "单点登录",
                        icon: <i className="fa-solid fa-right-to-bracket"></i>,
                        key: "sso_config",
                        onClick: () => loadJsonPage('/admin/config/sso_config')
                    }
                ],
            },
        ] : []),

        {
            label: "个人中心",
            icon: <i className="fa-solid fa-user"></i>,
            key: "user_profile",
            children: [
                {
                    label: "登录设置",
                    icon: <i className="fa-solid fa-key"></i>,
                    key: "user_profile_login_settings",
                    onClick: () => loadJsonPage('/user/profile/login_settings')
                },
                {
                    label: "我的集群",
                    icon: <i className="fa-solid fa-server"></i>,
                    key: "user_profile_clusters",
                    onClick: () => loadJsonPage('/user/profile/my_clusters')
                },
                {
                    label: "API密钥",
                    icon: <i className="fa-solid fa-key"></i>,
                    key: "user_profile_api_keys",
                    onClick: () => loadJsonPage('/user/profile/api_keys')
                },
                {
                    label: "开放MCP",
                    icon: <i className="fa-solid fa-share-nodes"></i>,
                    key: "user_profile_mcp_keys",
                    onClick: () => loadJsonPage('/user/profile/mcp_keys')
                },

            ],
        },
        {
            label: "开放API",
            title: "开放API",
            icon: <i className="fa-solid fa-code-branch"></i>,
            key: "open_api",
            onClick: () => open('/swagger/index.html')
        },
        {
            label: "关于",
            title: "关于",
            icon: <i className="fa-solid fa-circle-info"></i>,
            key: "about",
            onClick: () => loadJsonPage('/about/about')
        },
    ];
}

export default items;
