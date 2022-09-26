package credentials

import (
	"context"

	"flamingo.me/flamingo/v3/core/auth/oauth"
	"flamingo.me/grpc"
)

type GrpcOauth2Credentials struct {
	identifier *grpc.IdentityService
}

func (c *GrpcOauth2Credentials) Inject(identifier *grpc.IdentityService) {
	c.identifier = identifier
}

type ErrGrpcOauth2UnableToIdentify struct {
	msg string
	err error
}

func NewErrGrpcOauth2UnableToIdentify(msg string, err error) *ErrGrpcOauth2UnableToIdentify {
	return &ErrGrpcOauth2UnableToIdentify{
		msg: msg + ": " + err.Error(),
		err: err,
	}
}

func (e *ErrGrpcOauth2UnableToIdentify) Error() string {
	return e.msg
}

func (e *ErrGrpcOauth2UnableToIdentify) Unwrap() error {
	return e.err
}

func (c *GrpcOauth2Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	identity, err := c.identifier.IdentifyAs(ctx, oauth.OAuthTypeChecker)
	if identity == nil || err != nil {
		return nil, NewErrGrpcOauth2UnableToIdentify("unable to obtain identity", err)
	}

	token, err := identity.(oauth.Identity).TokenSource().Token()
	if err != nil {
		return nil, NewErrGrpcOauth2UnableToIdentify("unable to obtain token", err)
	}

	return map[string]string{
		"authorization": token.TokenType + " " + token.AccessToken,
	}, nil
}

func (*GrpcOauth2Credentials) RequireTransportSecurity() bool {
	return false
}
