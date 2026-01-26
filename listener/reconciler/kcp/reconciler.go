package kcp

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/platform-mesh/golang-commons/logger"
	"github.com/platform-mesh/kubernetes-graphql-gateway/common/config"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/apischema"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/workspacefile"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/reconciler"

	ctrl "sigs.k8s.io/controller-runtime"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"

	"github.com/kcp-dev/multicluster-provider/apiexport"
	kcpapis "github.com/kcp-dev/sdk/apis/apis/v1alpha2"
)

type KCPReconciler struct {
	mgr                        mcmanager.Manager
	provider                   *apiexport.Provider
	apiBindingReconciler       *APIBindingReconciler
	virtualWorkspaceReconciler *VirtualWorkspaceReconciler
	configWatcher              *ConfigWatcher
	log                        *logger.Logger
}

func NewKCPReconciler(
	appCfg config.Config,
	opts reconciler.ReconcilerOpts,
	log *logger.Logger,
) (*KCPReconciler, error) {
	log.Info().Msg("Setting up KCP reconciler with multicluster-provider")

	// Validate that endpoint slice name is configured
	endpointSliceName := appCfg.Listener.APIExportEndpointSliceName
	if endpointSliceName == "" {
		return nil, fmt.Errorf("APIExportEndpointSliceName must be configured for KCP mode")
	}

	// Create the apiexport provider with logging
	providerLogger := log.ComponentLogger("apiexport-provider").Logr()
	spew.Dump("Creating apiexport provider with endpoint slice name:", endpointSliceName)
	provider, err := apiexport.New(opts.Config, endpointSliceName, apiexport.Options{
		Scheme: opts.Scheme,
		Log:    &providerLogger,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create apiexport provider")
		return nil, err
	}

	// Create multicluster manager
	mgr, err := mcmanager.New(opts.Config, provider, opts.ManagerOpts)
	if err != nil {
		log.Error().Err(err).Msg("failed to create multicluster manager")
		return nil, err
	}

	// Create IO handler for schema files
	ioHandler, err := workspacefile.NewIOHandler(appCfg.OpenApiDefinitionsPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to create IO handler")
		return nil, err
	}

	// Create schema resolver
	schemaResolver := apischema.NewResolver(log)

	// Create cluster path resolver
	clusterPathResolver, err := NewClusterPathResolver(opts.Config, opts.Scheme)
	if err != nil {
		log.Error().Err(err).Msg("failed to create cluster path resolver")
		return nil, err
	}

	// Create discovery factory
	discoveryFactory, err := NewDiscoveryFactory(opts.Config)
	if err != nil {
		log.Error().Err(err).Msg("failed to create discovery factory")
		return nil, err
	}

	// Create APIBinding reconciler (but don't set up controller yet)
	apiBindingReconciler := &APIBindingReconciler{
		Scheme:              opts.Scheme,
		RestConfig:          opts.Config,
		IOHandler:           ioHandler,
		DiscoveryFactory:    discoveryFactory,
		APISchemaResolver:   schemaResolver,
		ClusterPathResolver: clusterPathResolver,
		Log:                 log,
		mcManager:           mgr,
	}

	// Setup virtual workspace components
	virtualWSManager := NewVirtualWorkspaceManager(appCfg)
	virtualWorkspaceReconciler := NewVirtualWorkspaceReconciler(
		virtualWSManager,
		ioHandler,
		schemaResolver,
		log,
	)

	configWatcher, err := NewConfigWatcher(virtualWSManager, log)
	if err != nil {
		log.Error().Err(err).Msg("failed to create config watcher")
		return nil, err
	}

	reconcilerInstance := &KCPReconciler{
		mgr:                        mgr,
		provider:                   provider,
		apiBindingReconciler:       apiBindingReconciler,
		virtualWorkspaceReconciler: virtualWorkspaceReconciler,
		configWatcher:              configWatcher,
		log:                        log,
	}

	log.Info().Str("endpointSlice", endpointSliceName).Msg("Successfully configured KCP reconciler with multicluster-provider")
	return reconcilerInstance, nil
}

func (r *KCPReconciler) GetManager() ctrl.Manager {
	return r.mgr.GetLocalManager()
}

func (r *KCPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// This method is required by the reconciler.CustomReconciler interface but is not used directly.
	// Actual reconciliation is handled by the APIBinding controller set up in SetupWithManager().
	// KCPReconciler acts as a coordinator/manager rather than a direct reconciler.
	return ctrl.Result{}, nil
}

func (r *KCPReconciler) SetupWithManager(_ ctrl.Manager) error {
	// Handle cases where the reconciler wasn't properly initialized (e.g., in tests)
	if r.apiBindingReconciler == nil {
		return nil
	}

	// Setup the APIBinding controller using multicluster builder
	// This watches APIBindings across all logical clusters via the APIExport virtual workspace
	if err := mcbuilder.ControllerManagedBy(r.mgr).
		Named("apibinding-controller").
		For(&kcpapis.APIBinding{}).
		Complete(r.apiBindingReconciler); err != nil {
		r.log.Error().Err(err).Msg("failed to setup APIBinding controller")
		return err
	}

	r.log.Info().Msg("Successfully set up APIBinding controller with multicluster-provider")
	return nil
}

// StartVirtualWorkspaceWatching starts watching virtual workspace configuration
func (r *KCPReconciler) StartVirtualWorkspaceWatching(ctx context.Context, configPath string) error {
	if configPath == "" {
		r.log.Info().Msg("no virtual workspace config path provided, skipping virtual workspace watching")
		return nil
	}

	r.log.Info().Str("configPath", configPath).Msg("starting virtual workspace configuration watching")

	// Start config watcher with a wrapper function
	changeHandler := func(config *VirtualWorkspacesConfig) error {
		if err := r.virtualWorkspaceReconciler.ReconcileConfig(ctx, config); err != nil {
			r.log.Error().Err(err).Msg("failed to reconcile virtual workspaces config")
			return err
		}
		return nil
	}
	return r.configWatcher.Watch(ctx, configPath, changeHandler)
}
