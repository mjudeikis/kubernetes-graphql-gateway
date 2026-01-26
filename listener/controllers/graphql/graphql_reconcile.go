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
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	gatewayv1alpha1 "github.com/platform-mesh/kubernetes-graphql-gateway/common/apis/v1alpha1"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/apischema"
	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/pkg/workspacefile"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// defaultTokenExpirationSeconds is the default token expiration time (1 hour)
	defaultTokenExpirationSeconds = 3600
)

var (
	ErrCreateHTTPClient       = errors.New("failed to create HTTP client")
	ErrCreateRESTMapper       = errors.New("failed to create REST mapper")
	ErrNoAuthConfigured       = errors.New("either kubeconfigSecretRef or serviceAccountRef must be specified")
	ErrKubeconfigKeyNotFound  = errors.New("kubeconfig key not found in secret")
	ErrServiceAccountNotFound = errors.New("service account not found")
	ErrTokenRequestFailed     = errors.New("failed to create token request")
)

type reconciler struct {
	ioHandler      *workspacefile.FileHandler
	schemaResolver apischema.Resolver
	// hostConfig is the rest.Config for the host cluster (where the controller runs)
	// Used to determine the API server URL when using ServiceAccountRef
	hostConfig *rest.Config
}

func newReconciler(ioHandler *workspacefile.FileHandler, schemaResolver apischema.Resolver, hostConfig *rest.Config) *reconciler {
	return &reconciler{
		ioHandler:      ioHandler,
		schemaResolver: schemaResolver,
		hostConfig:     hostConfig,
	}
}

func (r *reconciler) reconcile(ctx context.Context, cluster string, k8sClient client.Client, _ cache.Cache, graphql *gatewayv1alpha1.GraphQL) error {
	var errs []error
	logger := log.FromContext(ctx)

	if cluster == "" {
		cluster = "default"
	}

	graphqlName := graphql.GetName()
	graphqlNamespace := graphql.GetNamespace()
	schemaPath := fmt.Sprintf("%s", cluster)

	logger.Info("Processing GraphQL resource", "name", graphqlName, "namespace", graphqlNamespace)
	spew.Dump("GraphQL spec:", graphql.Spec)
	// Build rest.Config based on the auth method configured
	targetConfig, err := r.buildTargetConfig(ctx, k8sClient, graphql)
	if err != nil {
		logger.Error(err, "Failed to build target config")
		errs = append(errs, fmt.Errorf("failed to build target config: %w", err))
		return utilerrors.NewAggregate(errs)
	}

	logger.Info("Successfully loaded target config", "host", targetConfig.Host)

	// Create discovery client for target cluster
	targetDiscovery, err := discovery.NewDiscoveryClientForConfig(targetConfig)
	if err != nil {
		logger.Error(err, "Failed to create discovery client", "host", targetConfig.Host)
		errs = append(errs, fmt.Errorf("failed to create discovery client: %w", err))
		return utilerrors.NewAggregate(errs)
	}

	// Create REST mapper for target cluster
	targetRM, err := r.restMapperFromConfig(targetConfig)
	if err != nil {
		logger.Error(err, "Failed to create REST mapper", "host", targetConfig.Host)
		errs = append(errs, fmt.Errorf("failed to create REST mapper: %w", err))
		return utilerrors.NewAggregate(errs)
	}

	// Generate schema for target cluster using the schema resolver
	schemaJSON, err := r.schemaResolver.Resolve(targetDiscovery, targetRM)
	if err != nil {
		logger.Error(err, "Failed to resolve schema", "host", targetConfig.Host)
		errs = append(errs, fmt.Errorf("failed to resolve schema: %w", err))
		return utilerrors.NewAggregate(errs)
	}

	// If path changed, delete the old schema file referenced in the status
	prevPath := graphql.Status.ObservedPath
	if prevPath != "" && prevPath != schemaPath {
		if err := r.ioHandler.Delete(prevPath); err != nil {
			logger.Info("Failed to delete previous schema file", "previousPath", prevPath, "error", err)
		}
	}

	// Write schema to file
	if err := r.ioHandler.Write(schemaJSON, schemaPath); err != nil {
		logger.Error(err, "Failed to write schema", "path", schemaPath)
		errs = append(errs, fmt.Errorf("failed to write schema: %w", err))
		return utilerrors.NewAggregate(errs)
	}

	// Update status.ObservedPath if changed
	if prevPath != schemaPath {
		graphql.Status.ObservedPath = schemaPath
	}

	logger.Info("Successfully processed GraphQL resource", "name", graphqlName, "path", schemaPath)
	return utilerrors.NewAggregate(errs)
}

// buildTargetConfig builds a rest.Config based on the auth method configured in the GraphQL spec
func (r *reconciler) buildTargetConfig(ctx context.Context, k8sClient client.Client, graphql *gatewayv1alpha1.GraphQL) (*rest.Config, error) {
	// Priority: KubeconfigSecretRef > ServiceAccountRef
	if graphql.Spec.KubeconfigSecretRef != nil {
		return r.buildConfigFromKubeconfigSecret(ctx, k8sClient, graphql)
	}

	if graphql.Spec.ServiceAccountRef != nil {
		return r.buildConfigFromServiceAccount(ctx, k8sClient, graphql)
	}

	return nil, ErrNoAuthConfigured
}

// buildConfigFromKubeconfigSecret builds a rest.Config from a kubeconfig stored in a secret
func (r *reconciler) buildConfigFromKubeconfigSecret(ctx context.Context, k8sClient client.Client, graphql *gatewayv1alpha1.GraphQL) (*rest.Config, error) {
	secretRef := graphql.Spec.KubeconfigSecretRef

	// Determine namespace - use secret's namespace if specified, otherwise use GraphQL resource's namespace
	namespace := secretRef.Namespace
	if namespace == "" {
		namespace = graphql.GetNamespace()
	}

	// Get the secret containing the kubeconfig
	secret := &corev1.Secret{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      secretRef.Name,
		Namespace: namespace,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig secret %s/%s: %w", namespace, secretRef.Name, err)
	}

	// Extract kubeconfig data from secret using the specified key
	key := secretRef.Key
	if key == "" {
		key = "kubeconfig" // Default key if not specified
	}

	kubeconfigData, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("%w: key %q in secret %s/%s", ErrKubeconfigKeyNotFound, key, namespace, secretRef.Name)
	}

	// Parse kubeconfig and build rest.Config
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config from kubeconfig: %w", err)
	}

	return restConfig, nil
}

// buildConfigFromServiceAccount builds a rest.Config using a service account token
func (r *reconciler) buildConfigFromServiceAccount(ctx context.Context, k8sClient client.Client, graphql *gatewayv1alpha1.GraphQL) (*rest.Config, error) {
	saRef := graphql.Spec.ServiceAccountRef

	// Determine namespace - use SA's namespace if specified, otherwise use GraphQL resource's namespace
	namespace := saRef.Namespace
	if namespace == "" {
		namespace = graphql.GetNamespace()
	}

	// Get the service account
	sa := &corev1.ServiceAccount{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      saRef.Name,
		Namespace: namespace,
	}, sa)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s: %v", ErrServiceAccountNotFound, namespace, saRef.Name, err)
	}

	// Create a token request for the service account
	expirationSeconds := int64(defaultTokenExpirationSeconds)
	tokenRequest := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: &expirationSeconds,
		},
	}

	// Request a token for the service account
	err = k8sClient.SubResource("token").Create(ctx, sa, tokenRequest)
	if err != nil {
		return nil, fmt.Errorf("%w for service account %s/%s: %v", ErrTokenRequestFailed, namespace, saRef.Name, err)
	}

	if tokenRequest.Status.Token == "" {
		return nil, fmt.Errorf("received empty token from TokenRequest API for service account %s/%s", namespace, saRef.Name)
	}

	// Build rest.Config using the token
	// Use the host cluster's API server since the service account is in the same cluster
	config := &rest.Config{
		Host:        r.hostConfig.Host,
		BearerToken: tokenRequest.Status.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   r.hostConfig.CAData,
			CAFile:   r.hostConfig.CAFile,
			Insecure: r.hostConfig.Insecure,
		},
	}

	return config, nil
}

// restMapperFromConfig creates a REST mapper from a config
func (r *reconciler) restMapperFromConfig(cfg *rest.Config) (meta.RESTMapper, error) {
	httpClt, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, errors.Join(ErrCreateHTTPClient, err)
	}
	rm, err := apiutil.NewDynamicRESTMapper(cfg, httpClt)
	if err != nil {
		return nil, errors.Join(ErrCreateRESTMapper, err)
	}

	return rm, nil
}
