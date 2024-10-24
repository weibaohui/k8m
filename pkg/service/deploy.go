package service

import (
	"context"
	"strings"
	"time"

	"github.com/weibaohui/kom/kom"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployService struct {
}

func (d *DeployService) RestartDeploy(ctx context.Context, ns string, name string) (*v1.Deployment, error) {
	var deploy v1.Deployment
	err := kom.Init().WithContext(ctx).Resource(&deploy).Namespace(ns).Name(name).Get(&deploy).Error

	if err != nil {
		return nil, err
	}
	// 更新 Annotations，触发重启
	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新 Deployment
	err = kom.Init().WithContext(ctx).Resource(&deploy).Namespace(ns).Name(name).Update(&deploy).Error
	if err != nil {
		return nil, err
	}
	return &deploy, nil
}
func (d *DeployService) UpdateDeployImageTag(ctx context.Context, ns string, name string, containerName string, tag string) (*v1.Deployment, error) {
	var deploy v1.Deployment
	err := kom.Init().WithContext(ctx).Resource(&deploy).Namespace(ns).Name(name).Get(&deploy).Error

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
	err = kom.Init().WithContext(ctx).Resource(&deploy).Namespace(ns).Name(name).Update(&deploy).Error
	return &deploy, err
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
func (d *DeployService) CreateImagePullSecret(ctx context.Context, ns string, serviceAccount string, pullSecret string) error {

	secretName := "pull-secret"

	// 先查查Secrets 有没有
	secret := corev1.Secret{}
	err := kom.Init().WithContext(ctx).Resource(&secret).Namespace(ns).Name(secretName).Get(&secret).Error
	if err != nil && strings.Contains(err.Error(), "not found") {
		// 创建 secret
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(pullSecret),
			},
		}
		err = kom.Init().WithContext(ctx).Resource(&secret).Namespace(ns).Name(secretName).Create(&secret).Error
		if err != nil {
			return err
		}
	}

	var sa corev1.ServiceAccount
	// 将 secret 绑定到 ServiceAccount
	err = kom.Init().WithContext(ctx).Resource(&sa).Namespace(ns).Name(serviceAccount).Get(&sa).Error
	if err != nil {
		return err
	}

	// 检查是否已经绑定过该 secret
	for _, existingSecret := range sa.ImagePullSecrets {
		if existingSecret.Name == secretName {
			return nil // secret 已绑定，直接返回
		}
	}

	// 绑定 imagePullSecret
	sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName})
	err = kom.Init().WithContext(ctx).Resource(&sa).Namespace(ns).Name(serviceAccount).Update(&sa).Error

	return err
}
