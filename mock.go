package grpc

import (
	"context"
	"encoding/json"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/framework/config"
	"golang.org/x/oauth2"
)

type mockCallIdentifier struct {
	identifier string
	subject    string
	claims     []byte
}

var _ CallIdentifier = new(mockCallIdentifier)

func mockFactory(cfg config.Map) (CallIdentifier, error) {
	var config struct {
		Identifier string
		Subject    string
		Claims     string
	}

	if err := cfg.MapInto(&config); err != nil {
		return nil, err
	}

	return &mockCallIdentifier{
		identifier: config.Identifier,
		subject:    config.Subject,
		claims:     []byte(config.Claims),
	}, nil
}

func (identifier *mockCallIdentifier) Identifier() string {
	return identifier.identifier
}

type mockIdentity struct {
	identifier string
	subject    string
	claims     []byte
}

func (identity *mockIdentity) Broker() string {
	return identity.identifier
}

func (identity *mockIdentity) Subject() string {
	return identity.subject
}

func (identity *mockIdentity) AccessTokenClaims(into interface{}) error {
	return json.Unmarshal(identity.claims, into)
}

func (identity *mockIdentity) TokenSource() oauth2.TokenSource {
	return nil
}

func (identifier *mockCallIdentifier) Identify(ctx context.Context) (auth.Identity, error) {
	return &mockIdentity{
		identifier: identifier.identifier,
		subject:    identifier.subject,
		claims:     identifier.claims,
	}, nil
}
