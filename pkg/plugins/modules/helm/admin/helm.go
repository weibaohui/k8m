package admin

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	helm "github.com/weibaohui/k8m/pkg/plugins/modules/helm/service"
	"github.com/weibaohui/k8m/pkg/service"
)

func getHelm(c *gin.Context) (helm.Helm, error) {

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return nil, err
	}
	cluster := service.ClusterService().GetClusterByID(selectedCluster)

	// return h, err
	cmd := helm.NewHelmCmd("helm", selectedCluster, cluster)
	return cmd, nil
}
func getHelmWithNoCluster() (helm.Helm, error) {
	// return h, err
	cmd := helm.NewHelmCmdWithNoCluster("helm")
	return cmd, nil
}

func handleCommonLogic(c *gin.Context, action string, releaseName, namespace, repoName string) error {
	cluster, _ := amis.GetSelectedCluster(c)

	username := amis.GetLoginUser(c)
	roles, err := service.UserService().GetRolesByUserName(username)
	if err != nil {
		return err
	}

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         "Helm",
		Name:         releaseName,
		Namespace:    namespace,
		UserName:     username,
		Group:        repoName,
		Role:         strings.Join(roles, ","),
		ActionResult: "success",
	}

	err = check(c, cluster, namespace, releaseName, action)
	if err != nil {
		log.ActionResult = err.Error()
	}
	go service.OperationLogService().Add(&log)
	return err
}
func check(c *gin.Context, cluster, ns, name, action string) error {
	ctx := amis.GetContextWithUser(c)
	var nsList []string
	if ns != "" {
		nsList = append(nsList, ns)
	}
	err := comm.CheckPermissionLogic(ctx, cluster, nsList, ns, name, action)
	return err
}
