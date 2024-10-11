package kubectl

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) CreateImagePullSecret(ns string, serviceAccount string, pullSecret string) error {

	secretName := "pull-secret"

	// 先查查Secrets 有没有
	_, err := k8s.GetSecret(ns, secretName)
	if err != nil && strings.Contains(err.Error(), "not found") {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(pullSecret),
			},
		}
		// 创建 secret
		_, err := k8s.CreateSecret(secret)
		if err != nil {
			return err
		}
	}

	// 将 secret 绑定到 ServiceAccount
	sa, err := k8s.GetServiceAccount(ns, serviceAccount)
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
	_, err = k8s.UpdateServiceAccount(sa)
	return err
}
