package helm

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// RESTClientGetterImpl impl genericclioptions.RESTClientGetter
type RESTClientGetterImpl struct {
	rc *rest.Config
}

// NewRESTClientGetterImpl new RESTClientGetterImpl
func NewRESTClientGetterImpl(rc *rest.Config) *RESTClientGetterImpl {
	return &RESTClientGetterImpl{
		rc: rc,
	}
}

func (r *RESTClientGetterImpl) ToRESTConfig() (*rest.Config, error) {
	r.rc.ContentType = "application/json"
	r.rc.AcceptContentTypes = "application/json"
	return r.rc, nil
}

func (r *RESTClientGetterImpl) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	if dc, err := discovery.NewDiscoveryClientForConfig(r.rc); err != nil {
		return nil, err
	} else {
		return memory.NewMemCacheClient(dc), nil
	}
}

func (r *RESTClientGetterImpl) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(dc)

	return restmapper.NewShortcutExpander(mapper, dc, nil), nil
}

func (r *RESTClientGetterImpl) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	// e.g. helm.sh/helm/v3/pkg/cli/environment.go
	overrides.Context.Namespace = "default" // todo 传参

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
