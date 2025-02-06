package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/weibaohui/kom/kom"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

var linkCacheTTL = 1 * time.Minute

func (p *podService) LinksServices(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.Service, error) {

	services, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedService()
	return services, err
}

func (p *podService) LinksEndpoints(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.Endpoints, error) {
	return kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedEndpoints()
}

func (p *podService) LinksPVC(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.PersistentVolumeClaim, error) {
	return kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedPVC()
}

func (p *podService) LinksPV(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.PersistentVolume, error) {
	return kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedPV()
}

func (p *podService) LinksIngress(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*networkingv1.Ingress, error) {
	return kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedIngress()
}

func (p *podService) LinksEnv(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*kom.Env, error) {
	env, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedEnv()
	if err != nil {
		// error executing command: Internal error occurred: Internal error occurred: error executing command in container: failed to exec in container: failed to start exec \"915a4933acbb460d0b1859831d8f392dc96ca1f91447a94dbc41962900b91281\": OCI runtime exec failed: exec failed: unable to start container process: exec: \"env\": executable file not found in $PATH: unknown
		// 提取executable file not found in $PATH
		// 展示为简短通用的错误
		if strings.Contains(err.Error(), "executable file not found in $PATH") {
			return nil, fmt.Errorf("容器中无env命令")
		} else {
			return nil, err
		}
	}
	return env, nil
}

func (p *podService) LinksEnvFromPod(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*kom.Env, error) {
	env, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedEnvFromPod()
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (p *podService) LinksConfigMap(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.ConfigMap, error) {
	configMap, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedConfigMap()
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func (p *podService) LinksSecret(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*v1.Secret, error) {
	secret, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedSecret()
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (p *podService) LinksNode(ctx context.Context, selectedCluster string, item *v1.Pod) ([]*kom.SelectedNode, error) {
	return kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(item.Namespace).
		Name(item.Name).
		WithCache(linkCacheTTL).Ctl().Pod().LinkedNode()
}
