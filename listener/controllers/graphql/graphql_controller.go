/*
Copyright 2025 The Kube Bind Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graphql

import (
	"context"
	"fmt"
	"reflect"

	gatewayv1alpha1 "github.com/platform-mesh/kubernetes-graphql-gateway/common/apis/v1alpha1"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/apischema"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/workspacefile"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

const (
	controllerName = "graphql-cluster-controller"
)

// GraphQLReconciler reconciles a GraphQL object.
type GraphQLReconciler struct {
	manager    mcmanager.Manager
	opts       controller.TypedOptions[mcreconcile.Request]
	reconciler *reconciler
}

// NewGraphQLReconciler returns a new GraphQLReconciler to reconcile GraphQLs
// and its resources.
func NewGraphQLReconciler(
	_ context.Context,
	mgr mcmanager.Manager,
	opts controller.TypedOptions[mcreconcile.Request],
	ioHandler *workspacefile.FileHandler,
	schemaResolver apischema.Resolver,
	hostConfig *rest.Config,
) (*GraphQLReconciler, error) {
	r := &GraphQLReconciler{
		manager:    mgr,
		opts:       opts,
		reconciler: newReconciler(ioHandler, schemaResolver, hostConfig),
	}

	return r, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GraphQLReconciler) Reconcile(ctx context.Context, req mcreconcile.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling GraphQL", "request", req)

	cl, err := r.manager.GetCluster(ctx, req.ClusterName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get client for cluster %q: %w", req.ClusterName, err)
	}

	client := cl.GetClient()
	cache := cl.GetCache()

	graphql := &gatewayv1alpha1.GraphQL{}
	if err := client.Get(ctx, req.NamespacedName, graphql); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("GraphQL not found, skipping reconciliation")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get GraphQL: %w", err)
	}

	original := graphql.DeepCopy()
	if err := r.reconciler.reconcile(ctx, req.ClusterName, client, cache, graphql); err != nil {
		logger.Error(err, "Failed to reconcile GraphQL")
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(original, graphql) {
		err := client.Update(ctx, graphql)
		if err != nil {
			logger.Error(err, "Failed to update GraphQL status")
			return ctrl.Result{}, fmt.Errorf("failed to update GraphQL status: %w", err)
		}
		logger.Info("GraphQL status updated")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GraphQLReconciler) SetupWithManager(mgr mcmanager.Manager) error {
	return mcbuilder.ControllerManagedBy(mgr).
		For(&gatewayv1alpha1.GraphQL{}).
		WithOptions(r.opts).
		Named(controllerName).
		Complete(r)
}
