package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
)

type ServerRegistrar grpc.ServiceRegistrar

type ServerRegister func(server ServerRegistrar)

type Module struct{}

func (*Module) Configure(injector *dingo.Injector) {
	injector.Bind(new([]CallIdentifier)).ToProvider(buildIdentifier)
	injector.Bind(new(IdentityService)).In(dingo.ChildSingleton)
	injector.BindMap(new(CallIdentifierFactory), "oidc").ToInstance(oidcFactory)
	injector.BindMap(new(CallIdentifierFactory), "mock").ToInstance(mockFactory)
}

func buildIdentifier(
	provider map[string]CallIdentifierFactory,
	cfg *struct {
		Config config.Slice `inject:"config:grpc.identifier"`
	},
) []CallIdentifier {
	var identifiers []config.Map
	_ = cfg.Config.MapInto(&identifiers)

	res := make([]CallIdentifier, len(identifiers))

	var err error
	for i, identifier := range identifiers {
		identityProvider, ok := identifier["provider"].(string)
		if !ok {
			panic("no provider set")
		}
		factory, hasIt := provider[identityProvider]
		if !hasIt {
			panic("unknown identity provider " + identityProvider)
		}

		res[i], err = factory(identifier)
		if err != nil {
			panic(err)
		}

		if res[i] == nil {
			panic("can not build identity with provider " + identityProvider)
		}
	}

	return res
}

type ServerModule struct{}

func (*ServerModule) Configure(injector *dingo.Injector) {
	flamingo.BindEventSubscriber(injector).To(new(grpcServer))
}

func (*ServerModule) Depends() []dingo.Module {
	return []dingo.Module{
		new(Module),
	}
}

type grpcServer struct {
	register   []ServerRegister
	grpcServer *grpc.Server
}

func (s *grpcServer) Inject(register []ServerRegister) {
	s.register = register
}

func (s *grpcServer) Notify(ctx context.Context, event flamingo.Event) {
	switch event.(type) {
	case *flamingo.ServerStartEvent:
		go func() {
			if err := s.ServeTcpAddr(context.Background(), ":11101"); err != nil {
				log.Fatal(err)
			}
		}()
	case *flamingo.ShutdownEvent:
		s.grpcServer.GracefulStop()
	case *flamingo.ServerShutdownEvent:
		s.grpcServer.Stop()
	}
}

func (s *grpcServer) ServeTcpAddr(ctx context.Context, addr string) error {
	s.grpcServer = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{
		IsPublicEndpoint: false,
		StartOptions: trace.StartOptions{
			SpanKind: trace.SpanKindServer,
			Sampler:  trace.AlwaysSample(),
		},
	}))

	for _, rf := range s.register {
		rf(s.grpcServer)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("unable to listen: %w", err)
	}

	log.Println("ready to listen", addr)

	return s.grpcServer.Serve(listener)
}
