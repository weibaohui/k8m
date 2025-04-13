package constants

//	用户角色两种，平台管理员、普通用户
//
// 平台管理员拥有所有权限
// 普通用户需要赋予集群角色，
// 集群角色三种，集群管理员、集群只读、集群Pod内执行命令
const (
	RolePlatformAdmin = "platform_admin" // 平台管理员
	RoleGuest         = "guest"          // 普通用户，只能登录，约等于游客,无任何集群权限，也看不到集群列表

	RoleClusterAdmin    = "cluster_admin"    // 集群管理员
	RoleClusterReadonly = "cluster_readonly" // 集群只读权限
	RoleClusterPodExec  = "cluster_pod_exec" // 集群Pod内执行命令权限
)

// ClusterAuthorizationType 集群授权类型
type ClusterAuthorizationType string

const (
	ClusterAuthorizationTypeUser      ClusterAuthorizationType = "user"
	ClusterAuthorizationTypeUserGroup ClusterAuthorizationType = "user_group"
)
