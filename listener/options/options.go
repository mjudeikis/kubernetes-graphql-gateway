package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"

	providerkcp "github.com/platform-mesh/kubernetes-graphql-gateway/providers/kcp/options"
)

type Options struct {
	Logs *logs.Options

	ProviderKcp *providerkcp.Options

	ExtraOptions
}

type ExtraOptions struct {
	KubeConfig string

	Provider string

	SchemasDir string
}

type completedOptions struct {
	Logs *logs.Options

	// Provider specific options
	ProviderKcp *providerkcp.CompletedOptions

	ExtraOptions
}

type CompletedOptions struct {
	*completedOptions
}

func NewOptions() *Options {
	// Default to -v=2
	logs := logs.NewOptions()
	logs.Verbosity = logsv1.VerbosityLevel(2)

	opts := &Options{
		Logs:        logs,
		ProviderKcp: providerkcp.NewOptions(),

		ExtraOptions: ExtraOptions{
			Provider:   "kubernetes",
			SchemasDir: "/tmp/schemas",
		},
	}
	return opts
}

var providerAliases = map[string]string{
	"kcp":        "kcp",
	"kubernetes": "kubernetes",
	"":           "kubernetes",
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	logsv1.AddFlags(options.Logs, fs)
	options.ProviderKcp.AddFlags(fs)

	fs.StringVar(&options.KubeConfig, "kubeconfig", options.KubeConfig, "path to a kubeconfig. Only required if out-of-cluster")

	fs.StringVar(&options.Provider, "multicluster-runtime-provider", options.Provider,
		fmt.Sprintf("The multicluster runtime provider. Possible values are: %v", sets.List(sets.Set[string](sets.StringKeySet(providerAliases)))),
	)

	fs.StringVar(&options.SchemasDir, "schemas-dir", options.SchemasDir, "Directory to store schema files")
}

func (options *Options) Complete() (*CompletedOptions, error) {

	co := &CompletedOptions{
		completedOptions: &completedOptions{
			Logs:         options.Logs,
			ExtraOptions: options.ExtraOptions,
		},
	}

	if options.Provider == "kcp" {
		opts, err := options.ProviderKcp.Complete()
		if err != nil {
			return nil, err
		}
		co.completedOptions.ProviderKcp = opts
	}
	return co, nil
}

func (options *CompletedOptions) Validate() error {
	provider := providerAliases[options.Provider]
	if provider == "" {
		return fmt.Errorf("unknown provider %q, must be one of %v", options.Provider, sets.List(sets.Set[string](sets.StringKeySet(providerAliases))))
	}
	options.Provider = provider
	return nil
}
