package cronjob

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/klog/v2"
)

func Pause(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}
func Resume(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}

func BatchResume(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
			Ctl().CronJob().Resume()
		if x != nil {
			klog.V(6).Infof("批量恢复 cronjob 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchPause(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
			Ctl().CronJob().Pause()
		if x != nil {
			klog.V(6).Infof("批量暂停 cronjob 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
