package kubectl

import (
	"context"
	"strings"
	"time"

	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) GetDeploy(ns, name string) (*v1.Deployment, error) {
	deployment, err := k8s.client.AppsV1().Deployments(ns).Get(context.TODO(), name, metav1.GetOptions{})
	return deployment, err
}

func (k8s *Kubectl) CreateDeploy(deploy *v1.Deployment) (*v1.Deployment, error) {
	deployment, err := k8s.client.AppsV1().Deployments(deploy.Namespace).Create(context.TODO(), deploy, metav1.CreateOptions{})
	return deployment, err
}

func (k8s *Kubectl) RestartDeploy(ns string, name string) (*v1.Deployment, error) {
	deployment, err := k8s.GetDeploy(ns, name)
	if err != nil {
		return nil, err
	}
	// 更新 Annotations，触发重启
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新 Deployment
	updatedDeployment, err := k8s.client.AppsV1().Deployments(deployment.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return updatedDeployment, nil
}
func (k8s *Kubectl) UpdateDeployImageTag(ns string, name string, containerName string, tag string) (*v1.Deployment, error) {
	deploy, err := k8s.GetDeploy(ns, name)
	if err != nil {
		return nil, err
	}

	for i := range deploy.Spec.Template.Spec.Containers {
		c := &deploy.Spec.Template.Spec.Containers[i]
		if c.Name == containerName {
			// 调用 replaceImageTag 方法替换 tag
			c.Image = replaceImageTag(c.Image, tag)
		}
	}
	deployment, err := k8s.client.AppsV1().Deployments(deploy.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	return deployment, err
}

// replaceImageTag 替换镜像的 tag
func replaceImageTag(imageName, newTag string) string {
	// 检查镜像名称是否包含 tag
	if strings.Contains(imageName, ":") {
		// 按照 ":" 分割镜像名称和 tag
		parts := strings.Split(imageName, ":")
		// 使用新的 tag 替换旧的 tag
		return parts[0] + ":" + newTag
	} else {
		// 如果镜像名称中没有 tag，直接添加新的 tag
		return imageName + ":" + newTag
	}
}
