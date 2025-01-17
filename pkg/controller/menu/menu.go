package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
)

// Page 定义结构体
type Page struct {
	Label     string `json:"label,omitempty"`
	URL       string `json:"url,omitempty"`
	Redirect  string `json:"redirect,omitempty"`
	Icon      string `json:"icon,omitempty"`
	SchemaAPI string `json:"schemaApi,omitempty"`
	VisibleOn string `json:"visibleOn,omitempty"`
	Children  []Page `json:"children,omitempty"`
}

func List(c *gin.Context) {

	workload := Page{
		Label: "工作负载",
		Icon:  "fa fa-cube",
		Children: []Page{
			{
				Label:     "部署",
				URL:       "/deploy/list",
				Icon:      "fas fa-layer-group",
				SchemaAPI: "get:/pages/ns/deploy.json",
			},
			{
				Label:     "有状态集",
				URL:       "/statefulset/list",
				Icon:      "fas fa-poll-h",
				SchemaAPI: "get:/pages/ns/statefulset.json",
			},
			{
				Label:     "守护进程集",
				URL:       "/daemonset/list",
				Icon:      "fas fa-table",
				SchemaAPI: "get:/pages/ns/daemonset.json",
			},
			{
				Label:     "任务",
				URL:       "/job/list",
				Icon:      "fas fa-calculator",
				SchemaAPI: "get:/pages/ns/job.json",
			},
			{
				Label:     "定时任务",
				URL:       "/cronjob/list",
				Icon:      "fas fa-clock",
				SchemaAPI: "get:/pages/ns/cronjob.json",
			},
			{
				Label:     "容器组",
				URL:       "/pod/list",
				Icon:      "fas fa-cubes",
				SchemaAPI: "get:/pages/ns/pod.json",
			},
			{
				Label:     "副本集",
				URL:       "/replicaset/list",
				Icon:      "fas fa-clone",
				SchemaAPI: "get:/pages/ns/replicaset.json",
			},
			{
				Label:     "副本控制器",
				URL:       "/rc/list",
				Icon:      "fas fa-bacteria",
				SchemaAPI: "get:/pages/ns/rc.json",
				VisibleOn: "1==2",
			},
		},
	}
	config := Page{
		Label: "配置",
		Icon:  "fas fa-sliders-h",
		Children: []Page{
			{
				Label:     "配置映射",
				URL:       "/configmap/list",
				Icon:      "fas fa-file-code",
				SchemaAPI: "get:/pages/ns/configmap.json",
			},
			{
				Label:     "密钥",
				URL:       "/secret/list",
				Icon:      "fas fa-file-signature",
				SchemaAPI: "get:/pages/ns/secret.json",
			},
		},
	}
	network := Page{
		Label: "网络",
		Icon:  "fas fa-network-wired",
		Children: []Page{
			{
				Label:     "SVC服务",
				URL:       "/svc/list",
				Icon:      "fas fa-project-diagram",
				SchemaAPI: "get:/pages/ns/svc.json",
			},
			{
				Label:     "Ingress入口",
				URL:       "/ingress/list",
				Icon:      "fas fa-wifi",
				SchemaAPI: "get:/pages/ns/ing.json",
			},
		},
	}
	storage := Page{
		Label: "存储",
		Icon:  "fas fa-memory",
		Children: []Page{
			{
				Label:     "持久卷声明",
				URL:       "/pvc/list",
				Icon:      "fas fa-microchip",
				SchemaAPI: "get:/pages/ns/pvc.json",
			},
			{
				Label:     "持久卷",
				URL:       "/cluster/pv/list",
				Icon:      "fas fa-hdd",
				SchemaAPI: "get:/pages/cluster/pv.json",
			},
			{
				Label:     "存储类",
				URL:       "/cluster/storage_class/list",
				Icon:      "fas fa-coins",
				SchemaAPI: "get:/pages/cluster/storage_class.json",
			},
		},
	}
	rbac := Page{
		Label: "访问控制",
		Icon:  "fas fa-diagnoses",
		Children: []Page{
			{
				Label:     "服务账户",
				URL:       "/service_account/list",
				Icon:      "fas fa-id-card",
				SchemaAPI: "get:/pages/ns/service_account.json",
			},
			{
				Label:     "角色",
				URL:       "/role/list",
				Icon:      "fas fa-people-arrows",
				SchemaAPI: "get:/pages/ns/role.json",
			},
			{
				Label:     "角色绑定",
				URL:       "/role_binding/list",
				Icon:      "fas fa-wave-square",
				SchemaAPI: "get:/pages/ns/role_binding.json",
			},
			{
				Label:     "集群角色",
				URL:       "/cluster/cluster_role/list",
				Icon:      "fas fa-dice",
				SchemaAPI: "get:/pages/cluster/cluster_role.json",
			},
			{
				Label:     "集群角色绑定",
				URL:       "/cluster/cluster_role_binding/list",
				Icon:      "fas fa-vector-square",
				SchemaAPI: "get:/pages/cluster/cluster_role_binding.json",
			},
		},
	}
	clusterConfig := Page{
		Label: "集群配置",
		Icon:  "fas fa-info-circle",
		Children: []Page{
			{
				Label:     "API 服务",
				URL:       "/cluster/api_service/list",
				Icon:      "fas fa-screwdriver",
				SchemaAPI: "get:/pages/cluster/api_service.json",
			},
			{
				Label:     "流量规则",
				URL:       "/cluster/flow_schema/list",
				Icon:      "fas fa-cog",
				SchemaAPI: "get:/pages/cluster/flow_schema.json",
			},
			{
				Label:     "优先级配置",
				URL:       "/cluster/priority_level_config/list",
				Icon:      "fas fa-cog",
				SchemaAPI: "get:/pages/cluster/priority_level_config.json",
			},
			{
				Label:     "组件状态",
				URL:       "/cluster/component_status/list",
				Icon:      "fas fa-tools",
				SchemaAPI: "get:/pages/cluster/component_status.json",
			},
			{
				Label:     "Ingress入口类",
				URL:       "/cluster/ingress_class/list",
				Icon:      "fas fa-sitemap",
				SchemaAPI: "get:/pages/cluster/ingress_class.json",
			},
			{
				Label:     "网络策略",
				URL:       "/network_policy/list",
				Icon:      "fas fa-boxes",
				SchemaAPI: "get:/pages/ns/network_policy.json",
			},
			{
				Label:     "端点",
				URL:       "/endpoint/list",
				Icon:      "fas fa-ethernet",
				SchemaAPI: "get:/pages/ns/endpoint.json",
			},
			{
				Label:     "端点切片",
				URL:       "/endpointslice/list",
				Icon:      "fas fa-newspaper",
				SchemaAPI: "get:/pages/ns/endpointslice.json",
			},
			{
				Label:     "资源配额",
				URL:       "/resource_quota/list",
				Icon:      "fas fa-dungeon",
				SchemaAPI: "get:/pages/ns/resource_quota.json",
			},
			{
				Label:     "限制范围",
				URL:       "/limit_range/list",
				Icon:      "fas fa-compress",
				SchemaAPI: "get:/pages/ns/limit_range.json",
			},
			{
				Label:     "水平自动扩缩",
				URL:       "/hpa/list",
				Icon:      "fas fa-cogs",
				SchemaAPI: "get:/pages/ns/hpa.json",
			},
			{
				Label:     "Pod中断配置",
				URL:       "/pdb/list",
				Icon:      "fas fa-receipt",
				SchemaAPI: "get:/pages/ns/pdb.json",
			},
			{
				Label:     "租约",
				URL:       "/lease/list",
				Icon:      "fas fa-traffic-light",
				SchemaAPI: "get:/pages/ns/lease.json",
			},
			{
				Label:     "优先级类",
				URL:       "/cluster/priority_class/list",
				Icon:      "fas fa-user-shield",
				SchemaAPI: "get:/pages/cluster/priority_class.json",
			},
			{
				Label:     "运行时类",
				URL:       "/cluster/runtime_class/list",
				Icon:      "fas fa-ruler",
				SchemaAPI: "get:/pages/cluster/runtime_class.json",
			},
			{
				Label:     "验证钩子",
				URL:       "/cluster/validation_webhook/list",
				Icon:      "fas fa-cog",
				SchemaAPI: "get:/pages/cluster/validation_webhook.json",
			},
			{
				Label:     "变更钩子",
				URL:       "/cluster/mutating_webhook/list",
				Icon:      "fas fa-cog",
				SchemaAPI: "get:/pages/cluster/mutating_webhook.json",
			},
			{
				Label:     "CSI节点",
				URL:       "/cluster/csi_node/list",
				Icon:      "fas fa-cog",
				SchemaAPI: "get:/pages/cluster/csi_node.json",
			},
		},
	}

	crd := Page{
		Label: "CRD",
		Icon:  "fas fa-tape",
		Children: []Page{
			{
				Label:     "自定义资源",
				URL:       "/crd/list",
				Icon:      "fas fa-registered",
				SchemaAPI: "get:/pages/crd/crd.json",
			},

			{
				Label:     "CronTab",
				URL:       "/crd/crontab/list",
				Icon:      "fas fa-business-time",
				VisibleOn: "1==2",
				SchemaAPI: "get:/pages/crd/crontab.json",
			},
			{
				Label:     "NamespacedCR",
				URL:       "/crd/namespaced_cr/list",
				Icon:      "fas fa-business-time",
				VisibleOn: "1==2",
				SchemaAPI: "get:/pages/crd/namespaced_cr.json",
			},
			{
				Label:     "ClusterCR",
				URL:       "/crd/cluster_cr/list",
				Icon:      "fas fa-business-time",
				VisibleOn: "1==2",
				SchemaAPI: "get:/pages/crd/cluster_cr.json",
			},
		},
	}

	pages := []Page{
		{
			Label:    "Home",
			URL:      "/",
			Redirect: "/cluster/all",
		}, {
			Children: []Page{
				{
					Label:     "多集群",
					URL:       "/cluster/all",
					Icon:      "fas fa-server",
					SchemaAPI: "get:/pages/cluster/cluster_all.json",
				},
				{
					Label:     "创建",
					URL:       "/apply",
					Icon:      "fas fa-dharmachakra",
					SchemaAPI: "get:/pages/apply/apply.json",
				},
				{
					Label:     "命名空间",
					URL:       "/cluster/ns/list",
					Icon:      "fas fa-border-style",
					SchemaAPI: "get:/pages/cluster/ns.json",
				},
				{
					Label:     "节点",
					URL:       "/cluster/node/list",
					Icon:      "fas fa-server",
					SchemaAPI: "get:/pages/cluster/node.json",
				},
				{
					Label:     "事件",
					URL:       "/event/list",
					Icon:      "fas fa-calendar-alt",
					SchemaAPI: "get:/pages/ns/event.json",
				},
				workload,
				crd,
				config,
				network,
				storage,
				rbac,
				clusterConfig,
			},
		},
	}
	amis.WriteJsonData(c, gin.H{
		"pages": pages,
	})
}
