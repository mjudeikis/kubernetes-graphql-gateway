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

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1alpha1 "github.com/platform-mesh/kubernetes-graphql-gateway/common/apis/v1alpha1"
)

type reconciler struct {
}

func (r *reconciler) reconcile(ctx context.Context, client client.Client, _ cache.Cache, cluster *gatewayv1alpha1.GraphQL) error {
	var errs []error

	return utilerrors.NewAggregate(errs)
}
