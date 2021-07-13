package grpc

import (
	"context"
	"fmt"
	"time"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/core/auth/oauth"
	"flamingo.me/flamingo/v3/framework/config"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
)

type oauth2CallIdentifier struct {
	identifier string
	metakey    string
	provider   *oidc.Provider
	clientID   string
}

var _ CallIdentifier = new(oauth2CallIdentifier)

type oidcConfig struct {
	Identifier  string `json:"identifier"`
	Issuer      string `json:"issuer"`
	ClientID    string `json:"clientID"`
	MetadataKey string `json:"metadatakey"`
}

func oauth2Factory(cfg config.Map) (CallIdentifier, error) {
	var oidcConfig oidcConfig

	if err := cfg.MapInto(&oidcConfig); err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), oidcConfig.Issuer)
	if err != nil {
		return nil, err
	}

	if oidcConfig.MetadataKey == "" {
		oidcConfig.MetadataKey = "authorization"
	}

	return &oauth2CallIdentifier{
		identifier: oidcConfig.Identifier,
		provider:   provider,
		clientID:   oidcConfig.ClientID,
		metakey:    oidcConfig.MetadataKey,
	}, nil
}

func (identifier *oauth2CallIdentifier) Identifier() string {
	return identifier.identifier
}

type oauth2Identity struct {
	identifier string
	token      *oidc.IDToken
	rawToken   string
}

var _ oauth.Identity = new(oauth2Identity)

func (identity *oauth2Identity) Broker() string {
	return identity.identifier
}

func (identity *oauth2Identity) Subject() string {
	return identity.token.Subject
}

type staticTokenSource struct {
	identity *oauth2Identity
}

func (s staticTokenSource) Token() (*oauth2.Token, error) {
	if time.Now().After(s.identity.token.Expiry) {
		return nil, fmt.Errorf("token already timed out")
	}
	return &oauth2.Token{
		AccessToken: s.identity.rawToken,
		TokenType:   "Bearer",
		Expiry:      s.identity.token.Expiry,
	}, nil
}

func (identity *oauth2Identity) TokenSource() oauth2.TokenSource {
	return staticTokenSource{
		identity: identity,
	}
}

func (identity *oauth2Identity) AccessTokenClaims(into interface{}) error {
	return identity.token.Claims(into)
}

func (identifier *oauth2CallIdentifier) Identify(ctx context.Context) (auth.Identity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata available")
	}

	verifier := identifier.provider.Verifier(&oidc.Config{
		ClientID: identifier.clientID,
	})

	var err error
	var token *oidc.IDToken
	for _, rawToken := range md.Get(identifier.metakey) {
		token, err = verifier.Verify(ctx, rawToken)
		if err == nil {
			return &oauth2Identity{
				identifier: identifier.identifier,
				token:      token,
				rawToken:   rawToken,
			}, nil
		}
	}

	return nil, fmt.Errorf("can not identify call, last error: %#w", err)
}
