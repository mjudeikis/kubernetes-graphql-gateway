package listener

import (
	"context"
	"fmt"

	"github.com/platform-mesh/kubernetes-graphql-gateway/listener/controllers/graphql"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type Server struct {
	Config *Config

	Controllers
}

type Controllers struct {
	GraphQL *graphql.GraphQLReconciler
}

func NewServer(ctx context.Context, c *Config) (*Server, error) {
	logger := klog.FromContext(ctx)
	logger.Info("Setting up Listener Server controllers")

	s := &Server{
		Config: c,
	}

	opts := controller.TypedOptions[mcreconcile.Request]{}
	var err error
	s.GraphQL, err = graphql.NewGraphQLReconciler(
		ctx,
		s.Config.Manager,
		opts,
		s.Config.IOHandler,
		s.Config.SchemaResolver,
		s.Config.ClientConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up GraphQL Controller: %w", err)
	}
	if err := s.GraphQL.SetupWithManager(s.Config.Manager); err != nil {
		return nil, fmt.Errorf("error setting up GraphQL controller with manager: %w", err)
	}

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	logger.Info("Starting Listener")

	// start controller-runtime manager after bootstrap completes
	go func() {
		if err := s.Config.Manager.Start(ctx); err != nil {
			logger.Error(err, "Failed to start controller manager")
		}
	}()
	<-ctx.Done()
	logger.Info("Shutting down Listener Server")
	return nil
}
