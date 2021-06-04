package grpc

import (
	"context"
	"fmt"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/framework/config"
)

type CallIdentifierFactory func(config config.Map) (CallIdentifier, error)

type CallIdentifier interface {
	Identifier() string
	Identify(ctx context.Context) (auth.Identity, error)
}

type IdentityService struct {
	identityProviders []CallIdentifier
}

func (service *IdentityService) Inject(providers []CallIdentifier) *IdentityService {
	service.identityProviders = providers
	return service
}

func (service *IdentityService) Identify(ctx context.Context) auth.Identity {
	if service == nil {
		return nil
	}

	for _, provider := range service.identityProviders {
		if identity, _ := provider.Identify(ctx); identity != nil {
			return identity
		}
	}

	return nil
}

func (service *IdentityService) IdentifyFor(ctx context.Context, identifier string) (auth.Identity, error) {
	if service == nil {
		return nil, fmt.Errorf("grpc identity service is nil")
	}

	for _, provider := range service.identityProviders {
		if provider.Identifier() == identifier {
			return provider.Identify(ctx)
		}
	}

	return nil, fmt.Errorf("no identifier with code %q found", identifier)
}

func (service *IdentityService) IdentifyAll(ctx context.Context) []auth.Identity {
	if service == nil {
		return nil
	}

	var identities []auth.Identity

	for _, provider := range service.identityProviders {
		if identity, _ := provider.Identify(ctx); identity != nil {
			identities = append(identities, identity)
		}
	}

	return identities
}

// IdentifyAs returns an identity for a given interface
// identity, err := s.IdentifyAs(ctx, request, OpenIDTypeChecker)
// identity.(oauth.OpenIDIdentity)
func (service *IdentityService) IdentifyAs(ctx context.Context, checkType auth.IdentityTypeChecker) (auth.Identity, error) {
	if service == nil {
		return nil, fmt.Errorf("grpc identity service is nil")
	}

	for _, provider := range service.identityProviders {
		if identity, _ := provider.Identify(ctx); identity != nil {
			if checkType(identity) {
				return identity, nil
			}
		}
	}

	return nil, fmt.Errorf("no identity for type %T found", checkType)
}
