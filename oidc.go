package grpc

import (
	"context"
	"fmt"
	"log"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/framework/config"
	"github.com/coreos/go-oidc"
	"google.golang.org/grpc/metadata"
)

type oidcCallIdentifier struct {
	identifier string
	metakey    string
	provider   *oidc.Provider
	clientID   string
}

var _ CallIdentifier = new(oidcCallIdentifier)

type oidcConfig struct {
	Identifier  string `json:"identifier"`
	Issuer      string `json:"issuer"`
	ClientID    string `json:"clientID"`
	MetadataKey string `json:"metadatakey"`
}

func oidcFactory(cfg config.Map) (CallIdentifier, error) {
	var oidcConfig oidcConfig

	if err := cfg.MapInto(&oidcConfig); err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), oidcConfig.Issuer)
	if err != nil {
		return nil, err
	}

	return &oidcCallIdentifier{
		identifier: oidcConfig.Identifier,
		provider:   provider,
		clientID:   oidcConfig.ClientID,
		metakey:    oidcConfig.MetadataKey,
	}, nil
}

func (identifier *oidcCallIdentifier) Identifier() string {
	return identifier.identifier
}

type oidcIdentity struct {
	identifier string
	token      *oidc.IDToken
}

func (identity *oidcIdentity) Broker() string {
	return identity.identifier
}

func (identity *oidcIdentity) Subject() string {
	return identity.token.Subject
}

func (identifier *oidcCallIdentifier) Identify(ctx context.Context) (auth.Identity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata available")
	}

	verifier := identifier.provider.Verifier(&oidc.Config{
		ClientID: identifier.clientID,
	})

	var err error
	var token *oidc.IDToken
	log.Println(identifier.metakey, md.Get(identifier.metakey))
	for _, line := range md.Get(identifier.metakey) {
		log.Println("LINE", line)
		token, err = verifier.Verify(ctx, line)
		if err == nil {
			return &oidcIdentity{
				identifier: identifier.identifier,
				token:      token,
			}, nil
		}
	}

	return nil, fmt.Errorf("can not identify call, last error: %#w", err)
}
