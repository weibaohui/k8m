package helm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/models"
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

func handleCommonLogic(c *gin.Context, action string, releaseName, namespace, repoName string) (string, string, error) {
	cluster, _ := amis.GetSelectedCluster(c)
	ctx := amis.GetContextWithUser(c)
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         "Helm",
		Name:         releaseName,
		Namespace:    namespace,
		UserName:     username,
		Group:        repoName,
		Role:         role,
		ActionResult: "success",
	}

	err := check(c, cluster, namespace, releaseName, action)
	if err != nil {
		log.ActionResult = err.Error()
	}
	go service.OperationLogService().Add(&log)
	return username, role, err
}
func check(c *gin.Context, cluster, ns, name, action string) error {
	ctx := amis.GetContextWithUser(c)
	nsList := []string{ns}
	_, _, err := comm.CheckPermissionLogic(ctx, cluster, nsList, ns, name, action)
	return err
}
