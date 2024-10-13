package kubectl

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubectl *Kubectl
)

type Kubectl struct {
	client        *kubernetes.Clientset
	config        *rest.Config
	dynamicClient dynamic.Interface

	apiResources []metav1.APIResource
}

func Init() *Kubectl {
	return kubectl
}

func init() {
	log.Println("k8s client init")
	kubectl = &Kubectl{}

	config, err := getConfig()
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
			kubectl.apiResources = append(kubectl.apiResources, resource)
		}
	}
}

func getConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Printf("尝试读取集群内访问配置：%v\n", err)
		log.Println("尝试读取本地配置")
		// 不是在集群中,读取参数配置
		var kubeConfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeConfig file")
		} else {
			kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
		}
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			log.Println(err.Error())
		}

		log.Printf("服务器地址：%s\n", config.Host)
	}

	return config, err
}
