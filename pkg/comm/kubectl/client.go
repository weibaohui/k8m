package kubectl

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	kubectl *Kubectl
)
var apiResources []metav1.APIResource

type Kubectl struct {
	client        *kubernetes.Clientset
	config        *rest.Config
	dynamicClient dynamic.Interface

	callbacks *callbacks
	Statement *Statement
	clone     int
	Error     error
}

func Init() *Kubectl {
	return kubectl
}

// InitConnection 在主入口处进行初始化
func InitConnection(path string) {
	klog.V(2).Infof("k8s client init")
	kubectl = &Kubectl{clone: 1}

	config, err := getKubeConfig(path)
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	kubectl.client = client
	kubectl.config = config
	kubectl.dynamicClient = dynClient
	_, lists, _ := kubectl.client.Discovery().ServerGroupsAndResources()
	for _, list := range lists {

		resources := list.APIResources
		version := list.GroupVersionKind().Version
		group := list.GroupVersionKind().Group
		groupVersion := list.GroupVersion
		gvs := strings.Split(groupVersion, "/")
		if len(gvs) == 2 {
			group = gvs[0]
			version = gvs[1]
		} else {
			// 只有version的情况"v1"
			version = groupVersion
		}

		for _, resource := range resources {
			resource.Group = group
			resource.Version = version
			apiResources = append(apiResources, resource)
		}
	}

	// 注册回调参数
	kubectl.Statement = &Statement{
		Kubectl:       kubectl,
		Context:       context.Background(),
		client:        client,
		DynamicClient: dynClient,
		config:        config,
	}

	kubectl.callbacks = initializeCallbacks(kubectl)

}

func getKubeConfig(path string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		klog.V(2).Infof("尝试读取集群内访问配置：%v\n", err)
		klog.V(2).Infof("尝试读取本地配置%s", path)
		// 不是在集群中,读取参数配置
		config, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			klog.Errorf(err.Error())
		}

	}
	if config != nil {
		klog.V(2).Infof("服务器地址：%s\n", config.Host)
	}
	return config, err
}

func (k8s *Kubectl) getInstance() *Kubectl {
	if k8s.clone > 0 {
		tx := &Kubectl{Error: k8s.Error}

		if k8s.clone == 1 {
			// clone with new statement
			tx.Statement = &Statement{
				Kubectl:       tx,
				Context:       k8s.Statement.Context,
				client:        k8s.Statement.client,
				DynamicClient: k8s.Statement.DynamicClient,
				config:        k8s.Statement.config,
			}
			tx.callbacks = k8s.callbacks

		} else {
			// with clone statement
			tx.Statement = k8s.Statement.clone()
			tx.callbacks = k8s.callbacks
			tx.Statement.Kubectl = tx

		}

		return tx
	}

	return k8s
}
func (k8s *Kubectl) Callback() *callbacks {
	return k8s.callbacks
}
