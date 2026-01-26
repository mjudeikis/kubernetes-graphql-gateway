package listener

import (
	"fmt"

	"github.com/kcp-dev/multicluster-provider/apiexport"
	gatewayv1alpha1 "github.com/platform-mesh/kubernetes-graphql-gateway/common/apis/v1alpha1"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/options"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"

	kcpapisv1alpha1 "github.com/kcp-dev/sdk/apis/apis/v1alpha1"
	kcpapis "github.com/kcp-dev/sdk/apis/apis/v1alpha2"
	kcpcore "github.com/kcp-dev/sdk/apis/core/v1alpha1"
	kcptenancy "github.com/kcp-dev/sdk/apis/tenancy/v1alpha1"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"

	kcpprovider "github.com/platform-mesh/kubernetes-graphql-gateway/providers/kcp"
)

type Config struct {
	Options *options.CompletedOptions

	Provider multicluster.Provider

	Manager mcmanager.Manager
	Scheme  *runtime.Scheme

	ClientConfig *rest.Config
}

func NewConfig(options *options.CompletedOptions) (*Config, error) {
	config := &Config{
		Options: options,
	}

	// create clients
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = options.KubeConfig
	var err error
	config.ClientConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	config.ClientConfig = rest.CopyConfig(config.ClientConfig)
	config.ClientConfig = rest.AddUserAgent(config.ClientConfig, "kube-bind-backend")

	// Set up controller-runtime manager
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding client-go scheme: %w", err)
	}
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding apiextensions scheme: %w", err)
	}
	if err := gatewayv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error adding kubebind scheme: %w", err)
	}

	config.Scheme = scheme

	switch options.Provider {
	case "kcp":
		if options.ProviderKcp == nil {
			return nil, fmt.Errorf("kcp provider options must be provided when provider is kcp")
		}
		if err := kcpapisv1alpha1.AddToScheme(scheme); err != nil {
			return nil, fmt.Errorf("error adding apis scheme: %w", err)
		}
		if err := kcpapis.AddToScheme(scheme); err != nil {
			return nil, fmt.Errorf("error adding apis scheme: %w", err)
		}
		if err := kcpcore.AddToScheme(scheme); err != nil {
			return nil, fmt.Errorf("error adding core scheme: %w", err)
		}
		if err := kcptenancy.AddToScheme(scheme); err != nil {
			return nil, fmt.Errorf("error adding tenancy scheme: %w", err)
		}

		provider, err := kcpprovider.New(config.ClientConfig, options.ProviderKcp.APIExportEndpointSliceName, apiexport.Options{
			Scheme: scheme,
		})
		if err != nil {
			return nil, fmt.Errorf("error setting up kcp provider: %w", err)
		}

		config.Provider = provider
	default:
		config.Provider = nil
	}

	opts := ctrl.Options{
		Controller: ctrlconfig.Controller{},
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		Scheme: scheme,
	}

	manager, err := mcmanager.New(config.ClientConfig, config.Provider, opts)
	if err != nil {
		return nil, fmt.Errorf("error setting up controller manager: %w", err)
	}

	config.Manager = manager

	return config, nil
}
