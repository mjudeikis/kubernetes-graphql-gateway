package kcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/platform-mesh/golang-commons/logger"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/apischema"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/workspacefile"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

// Ensure APIBindingReconciler implements mcreconcile.Reconciler
var _ mcreconcile.Reconciler = &APIBindingReconciler{}

// APIBindingReconciler reconciles an APIBinding object
type APIBindingReconciler struct {
	Scheme              *runtime.Scheme
	RestConfig          *rest.Config
	IOHandler           *workspacefile.FileHandler
	DiscoveryFactory    DiscoveryFactory
	APISchemaResolver   apischema.Resolver
	ClusterPathResolver ClusterPathResolver
	Log                 *logger.Logger
	mcManager           mcmanager.Manager
}

func (r *APIBindingReconciler) Reconcile(ctx context.Context, req mcreconcile.Request) (ctrl.Result, error) {
	r.Log.Info().Str("cluster", req.ClusterName).Str("name", req.Name).Msg("Reconcile called")

	// ignore system workspaces (e.g. system:shard)
	if strings.HasPrefix(req.ClusterName, "system") {
		r.Log.Info().Str("cluster", req.ClusterName).Msg("Ignoring system workspace")
		return ctrl.Result{}, nil
	}

	logger := r.Log.With().Str("cluster", req.ClusterName).Str("name", req.Name).Logger()

	// Get cluster from the multicluster manager
	cluster, err := r.mcManager.GetCluster(ctx, req.ClusterName)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get cluster from manager")
		return ctrl.Result{}, fmt.Errorf("failed to get cluster: %w", err)
	}
	clusterClt := cluster.GetClient()

	clusterPath, err := PathForCluster(req.ClusterName, clusterClt)
	if err != nil {
		if errors.Is(err, ErrClusterIsDeleted) {
			logger.Info().Msg("cluster is deleted, triggering cleanup")
			if err = r.IOHandler.Delete(clusterPath); err != nil {
				logger.Error().Err(err).Msg("failed to delete workspace file after cluster deletion")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		logger.Error().Err(err).Msg("failed to get cluster path")
		return ctrl.Result{}, err
	}

	logger = logger.With().Str("clusterPath", clusterPath).Logger()
	logger.Info().Msg("starting reconciliation...")

	dc, err := r.DiscoveryFactory.ClientForCluster(clusterPath)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create discovery client for cluster")
		return ctrl.Result{}, err
	}

	rm, err := r.DiscoveryFactory.RestMapperForCluster(clusterPath)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create rest mapper for cluster")
		return ctrl.Result{}, err
	}

	// Generate current schema
	currentSchema, err := r.generateCurrentSchema(dc, rm, clusterPath)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Read existing schema (if it exists)
	savedSchema, err := r.IOHandler.Read(clusterPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		logger.Error().Err(err).Msg("failed to read existing schema file")
		return ctrl.Result{}, err
	}

	// Write if file doesn't exist or content has changed
	if errors.Is(err, fs.ErrNotExist) || !bytes.Equal(currentSchema, savedSchema) {
		if err := r.IOHandler.Write(currentSchema, clusterPath); err != nil {
			logger.Error().Err(err).Msg("failed to write schema to filesystem")
			return ctrl.Result{}, err
		}
		logger.Info().Msg("schema file updated")
	}

	return ctrl.Result{}, nil
}

// generateCurrentSchema is a subroutine that resolves the current API schema and injects KCP metadata
func (r *APIBindingReconciler) generateCurrentSchema(dc discovery.DiscoveryInterface, rm meta.RESTMapper, clusterPath string) ([]byte, error) {
	// Use shared schema generation logic
	return generateSchemaWithMetadata(
		SchemaGenerationParams{
			ClusterPath:     clusterPath,
			DiscoveryClient: dc,
			RESTMapper:      rm,
			// No HostOverride for regular workspaces - uses environment kubeconfig
		},
		r.APISchemaResolver,
		r.Log,
	)
}
