package service

import (
	"context"
	"strings"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployService struct {
}

func (d *deployService) CreateImagePullSecret(ctx context.Context, selectedCluster string, ns string, serviceAccount string, pullSecret string) error {

	secretName := "pull-secret"

	// 先查查Secrets 有没有
	secret := corev1.Secret{}
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&secret).Namespace(ns).Name(secretName).Get(&secret).Error
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
		err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&secret).Namespace(ns).Name(secretName).Create(&secret).Error
		if err != nil {
			return err
		}
	}

	var sa corev1.ServiceAccount
	// 将 secret 绑定到 ServiceAccount
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&sa).Namespace(ns).Name(serviceAccount).Get(&sa).Error
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
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&sa).Namespace(ns).Name(serviceAccount).Update(&sa).Error

	return err
}
