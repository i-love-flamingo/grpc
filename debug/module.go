package debug

import (
	"context"

	"flamingo.me/dingo"
	"flamingo.me/grpc"
)

type Module struct{}

func (*Module) Configure(injector *dingo.Injector) {
	injector.BindMulti(new(grpc.ServerRegister)).ToProvider(registerProvider)
}

func registerProvider(impl *impl) grpc.ServerRegister {
	return func(server grpc.ServerRegistrar) {
		RegisterFlamingoGrpcDebugServer(server, impl)
	}
}

type impl struct {
	UnimplementedFlamingoGrpcDebugServer
	identifier *grpc.IdentityService
}

func (impl *impl) Inject(identifier *grpc.IdentityService) {
	impl.identifier = identifier
}

func (impl *impl) Identify(ctx context.Context, _ *IdentifyRequest) (*IdentityResponse, error) {
	identity := impl.identifier.Identify(ctx)
	return &IdentityResponse{
		Subject:    identity.Subject(),
		Identifier: identity.Broker(),
	}, nil
}
