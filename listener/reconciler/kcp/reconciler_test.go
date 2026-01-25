package kcp_test

import (
	"testing"

	"github.com/platform-mesh/golang-commons/logger"
	"github.com/platform-mesh/kubernetes-graphql-gateway/common/config"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/reconciler"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/reconciler/kcp"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func TestNewKCPReconciler(t *testing.T) {
	mockLogger, _ := logger.New(logger.DefaultConfig())

	tests := []struct {
		name        string
		appCfg      config.Config
		opts        reconciler.ReconcilerOpts
		wantErr     bool
		errContains string
	}{
		{
			name: "missing_endpoint_slice_name",
			appCfg: config.Config{
				OpenApiDefinitionsPath: t.TempDir(),
			},
			opts: reconciler.ReconcilerOpts{
				Config: &rest.Config{
					Host: "https://kcp.example.com",
				},
				Scheme: runtime.NewScheme(),
				ManagerOpts: ctrl.Options{
					Metrics: server.Options{BindAddress: "0"},
				},
			},
			wantErr:     true,
			errContains: "APIExportEndpointSliceName must be configured",
		},
		// Note: Tests for "invalid_openapi_definitions_path", "nil_scheme", and "successful_creation"
		// have been removed as they require a real KCP cluster connection.
		// The apiexport.New() function attempts to connect to the API server before
		// other validations occur. These scenarios are tested in integration tests.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler, err := kcp.NewKCPReconciler(tt.appCfg, tt.opts, mockLogger)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, reconciler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reconciler)
				assert.NotNil(t, reconciler.GetManager())
			}
		})
	}
}

func TestKCPReconciler_GetManager(t *testing.T) {
	// Note: GetManager() on a nil or empty reconciler will panic because
	// it calls GetLocalManager() on a nil mcManager. This is expected behavior
	// as the reconciler should only be used after proper initialization.
	// Integration tests verify this with a real cluster.
	t.Skip("Skipping: GetManager requires a properly initialized reconciler with real cluster connection")
}

func TestKCPReconciler_Reconcile(t *testing.T) {
	reconciler := &kcp.ExportedKCPReconciler{}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	}

	// The Reconcile method should be a no-op and always return empty result with no error
	result, err := reconciler.Reconcile(t.Context(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestKCPReconciler_SetupWithManager(t *testing.T) {
	reconciler := &kcp.ExportedKCPReconciler{}

	// The SetupWithManager method should be a no-op when apiBindingReconciler is nil
	// (which is the case for an empty struct)
	err := reconciler.SetupWithManager(nil)

	assert.NoError(t, err)
}
