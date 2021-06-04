package grpc

import (
	"context"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/framework/config"
)

type mockCallIdentifier struct {
	identifier string
	subject    string
}

var _ CallIdentifier = new(mockCallIdentifier)

func mockFactory(cfg config.Map) (CallIdentifier, error) {
	var config struct {
		Identifier string
		Subject    string
	}

	if err := cfg.MapInto(&config); err != nil {
		return nil, err
	}

	return &mockCallIdentifier{
		identifier: config.Identifier,
		subject:    config.Subject,
	}, nil
}

func (identifier *mockCallIdentifier) Identifier() string {
	return identifier.identifier
}

type mockIdentity struct {
	identifier string
	subject    string
}

func (identity *mockIdentity) Broker() string {
	return identity.identifier
}

func (identity *mockIdentity) Subject() string {
	return identity.subject
}

func (identifier *mockCallIdentifier) Identify(ctx context.Context) (auth.Identity, error) {
	return &mockIdentity{
		identifier: identifier.identifier,
		subject:    identifier.subject,
	}, nil
}
