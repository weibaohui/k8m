import {useNavigate} from "react-router-dom";
import type {MenuProps} from 'antd';
import {useEffect, useState} from 'react';
import {fetcher} from '../Amis/fetcher';

// 定义用户角色接口
interface UserRoleResponse {
    role: string;  // 根据实际数据结构调整类型
}

type MenuItem = Required<MenuProps>['items'][number];

const items: () => MenuItem[] = () => {
    const navigate = useNavigate()
    const [userRole, setUserRole] = useState<string>('');

    useEffect(() => {
        const fetchUserRole = async () => {
            try {
                const response = await fetcher({
                    url: '/mgm/user/role',
                    method: 'get'
                });
                // 检查 response.data 是否存在，并确保其类型正确
                if (response.data && typeof response.data === 'object') {
                    const role = response.data.data as UserRoleResponse;
                    setUserRole(role.role);
                    // console.log('User Role:', role.role);
                }
            } catch (error) {
                console.error('Failed to fetch user role:', error);
            }
        };
        fetchUserRole();
    }, []);

    const onMenuClick = (path: string) => {
        navigate(path)
    }
    return [
        {
            label: "多集群",
            title: "多集群",
            icon: <i className="fa-solid fa-server"></i>,
            key: "cluster_all",
            onClick: () => onMenuClick('/cluster/cluster_all')
        },
        {
            label: "命名空间",
            title: "命名空间",
            icon: <i className="fa-solid fa-border-style"></i>,
            key: "cluster_ns",
            onClick: () => onMenuClick('/cluster/ns')
        },
        {
            label: "节点",
            title: "节点",
            icon: <i className="fa-solid fa-computer"></i>,
            key: "cluster_node",
            onClick: () => onMenuClick('/cluster/node')
        },
        {
            label: "事件",
            title: "事件",
            icon: <i className="fa-solid fa-bell"></i>,
            key: "event",
            onClick: () => onMenuClick('/ns/event')
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
                    onClick: () => onMenuClick('/ns/deploy')
                },
                {
                    label: "有状态集",
                    title: "有状态集",
                    icon: <i className="fa-solid fa-database"></i>,
                    key: "statefulset",
                    onClick: () => onMenuClick('/ns/statefulset')
                },
                {
                    label: "守护进程集",
                    title: "守护进程集",
                    icon: <i className="fa-solid fa-shield-halved"></i>,
                    key: "daemonset",
                    onClick: () => onMenuClick('/ns/daemonset')
                },
                {
                    label: "任务",
                    title: "任务",
                    icon: <i className="fa-solid fa-list-check"></i>,
                    key: "job",
                    onClick: () => onMenuClick('/ns/job')
                },
                {
                    label: "定时任务",
                    title: "定时任务",
                    icon: <i className="fa-solid fa-clock"></i>,
                    key: "cronjob",
                    onClick: () => onMenuClick('/ns/cronjob')
                },
                {
                    label: "容器组",
                    title: "容器组",
                    icon: <i className="fa-solid fa-cubes"></i>,
                    key: "pod",
                    onClick: () => onMenuClick('/ns/pod')
                },
                {
                    label: "副本集",
                    title: "副本集",
                    icon: <i className="fa-solid fa-clone"></i>,
                    key: "replicaset",
                    onClick: () => onMenuClick('/ns/replicaset')
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
                    onClick: () => onMenuClick('/crd/crd')
                }
            ],
        },
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
                    onClick: () => onMenuClick('/helm/repo')
                },
                {
                    label: "应用包",
                    title: "应用包",
                    icon: <i className="fa-solid fa-cubes"></i>,
                    key: "helm_chart",
                    onClick: () => onMenuClick('/helm/chart')
                },
                {
                    label: "应用实例",
                    title: "应用实例",
                    icon: <i className="fas fa-layer-group"></i>,
                    key: "helm_release",
                    onClick: () => onMenuClick('/helm/release')
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
                    onClick: () => onMenuClick('/ns/configmap')
                },
                {
                    label: "密钥",
                    icon: <i className="fa-solid fa-key"></i>,
                    key: "secret",
                    onClick: () => onMenuClick('/ns/secret')
                },
                {
                    label: "验证钩子",
                    icon: <i className="fa-solid fa-check"></i>,
                    key: "validation_webhook",
                    onClick: () => onMenuClick('/cluster/validation_webhook')
                },
                {
                    label: "变更钩子",
                    icon: <i className="fa-solid fa-exchange"></i>,
                    key: "mutating_webhook",
                    onClick: () => onMenuClick('/cluster/mutating_webhook')
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
                    onClick: () => onMenuClick('/ns/svc')
                },
                {
                    label: "Ingress入口",
                    icon: <i className="fa-solid fa-wifi"></i>,
                    key: "ingress",
                    onClick: () => onMenuClick('/ns/ing')
                },
                {
                    label: "Ingress入口类",
                    icon: <i className="fa-solid fa-sitemap"></i>,
                    key: "ingress_class",
                    onClick: () => onMenuClick('/cluster/ingress_class')
                },
            ],
        },
        {
            label: "存储",
            icon: <i className="fa-solid fa-memory"></i>,
            key: "storage",
            children: [
                {
                    label: "持久卷声明",
                    icon: <i className="fa-solid fa-folder"></i>,
                    key: "pvc",
                    onClick: () => onMenuClick('/ns/pvc')
                },
                {
                    label: "持久卷",
                    icon: <i className="fa-solid fa-hdd"></i>,
                    key: "pv",
                    onClick: () => onMenuClick('/cluster/pv')
                },
                {
                    label: "存储类",
                    icon: <i className="fa-solid fa-coins"></i>,
                    key: "storage_class",
                    onClick: () => onMenuClick('/cluster/storage_class')
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
                    onClick: () => onMenuClick('/ns/service_account')
                },
                {
                    label: "角色",
                    icon: <i className="fa-solid fa-user-tag"></i>,
                    key: "role",
                    onClick: () => onMenuClick('/ns/role')
                },
                {
                    label: "角色绑定",
                    icon: <i className="fa-solid fa-link"></i>,
                    key: "role_binding",
                    onClick: () => onMenuClick('/ns/role_binding')
                },
                {
                    label: "集群角色",
                    icon: <i className="fa-solid fa-users"></i>,
                    key: "cluster_role",
                    onClick: () => onMenuClick('/cluster/cluster_role')
                },
                {
                    label: "集群角色绑定",
                    icon: <i className="fa-solid fa-user-lock"></i>,
                    key: "cluster_role_binding",
                    onClick: () => onMenuClick('/cluster/cluster_role_binding')
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
                    onClick: () => onMenuClick('/ns/endpoint')
                },
                {
                    label: "端点切片",
                    icon: <i className="fa-solid fa-newspaper"></i>,
                    key: "endpointslice",
                    onClick: () => onMenuClick('/ns/endpointslice')
                },
                {
                    label: "水平自动扩缩",
                    icon: <i className="fa-solid fa-arrows-left-right"></i>,
                    key: "hpa",
                    onClick: () => onMenuClick('/ns/hpa')
                },
                {
                    label: "网络策略",
                    icon: <i className="fa-solid fa-project-diagram"></i>,
                    key: "network_policy",
                    onClick: () => onMenuClick('/ns/network_policy')
                },
                {
                    label: "资源配额",
                    icon: <i className="fa-solid fa-chart-pie"></i>,
                    key: "resource_quota",
                    onClick: () => onMenuClick('/ns/resource_quota')
                },
                {
                    label: "限制范围",
                    icon: <i className="fa-solid fa-compress"></i>,
                    key: "limit_range",
                    onClick: () => onMenuClick('/ns/limit_range')
                },
                {
                    label: "Pod中断配置",
                    icon: <i className="fa-solid fa-receipt"></i>,
                    key: "pdb",
                    onClick: () => onMenuClick('/ns/pdb')
                },
                {
                    label: "租约",
                    icon: <i className="fa-solid fa-file-contract"></i>,
                    key: "lease",
                    onClick: () => onMenuClick('/ns/lease')
                },
                {
                    label: "优先级类",
                    icon: <i className="fa-solid fa-sort"></i>,
                    key: "priority_class",
                    onClick: () => onMenuClick('/cluster/priority_class')
                },
                {
                    label: "运行时类",
                    icon: <i className="fa-solid fa-play"></i>,
                    key: "runtime_class",
                    onClick: () => onMenuClick('/cluster/runtime_class')
                },
                {
                    label: "CSI节点",
                    icon: <i className="fa-solid fa-server"></i>,
                    key: "csi_node",
                    onClick: () => onMenuClick('/cluster/csi_node')
                },
                {
                    label: "API 服务",
                    icon: <i className="fa-solid fa-code"></i>,
                    key: "api_service",
                    onClick: () => onMenuClick('/cluster/api_service')
                },
                {
                    label: "流量规则",
                    icon: <i className="fa-solid fa-random"></i>,
                    key: "flow_schema",
                    onClick: () => onMenuClick('/cluster/flow_schema')
                },
                {
                    label: "优先级配置",
                    icon: <i className="fa-solid fa-sliders"></i>,
                    key: "priority_level_config",
                    onClick: () => onMenuClick('/cluster/priority_level_config')
                },
                {
                    label: "组件状态",
                    icon: <i className="fa-solid fa-info-circle"></i>,
                    key: "component_status",
                    onClick: () => onMenuClick('/cluster/component_status')
                },
            ],
        },
        {
            label: "平台设置",
            icon: <i className="fa-solid fa-wrench"></i>,
            key: "platform_settings",
            children: [
                {
                    label: "用户管理",
                    icon: <i className="fa-solid fa-users"></i>,
                    key: "user_management",
                    onClick: () => onMenuClick('/user/user')
                },
                ...(userRole === 'platform_admin' ? [
                    {
                        label: "用户组管理",
                        icon: <i className="fa-solid fa-user-group"></i>,
                        key: "user_group_management",
                        onClick: () => onMenuClick('/user/user_group')
                    },
                    {
                        label: "MCP管理",
                        icon: <i className="fa-solid fa-server"></i>,
                        key: "mcp_management",
                        onClick: () => onMenuClick('/mcp/mcp')
                    }
                ] : []),
            ],
        },
    ];
}

export default items;
