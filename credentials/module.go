package credentials

import (
	"context"
	"fmt"
	"sync"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/core/auth/oauth"
	"flamingo.me/flamingo/v3/framework/web"
)

type WebOidcCredentials struct {
	identifier *auth.WebIdentityService
}

func (c *WebOidcCredentials) Inject(identifier *auth.WebIdentityService) {
	c.identifier = identifier
}

var tokenLock = new(sync.Mutex)

type ErrWebOidcUnableToIdentify struct {
	msg string
	err error
}

func NewErrWebOidcUnableToIdentify(msg string, err error) *ErrWebOidcUnableToIdentify {
	return &ErrWebOidcUnableToIdentify{
		msg: msg + ": " + err.Error(),
		err: err,
	}
}

func (e *ErrWebOidcUnableToIdentify) Error() string {
	return e.msg
}

func (e *ErrWebOidcUnableToIdentify) Unwrap() error {
	return e.err
}

func (c *WebOidcCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	req := web.RequestFromContext(ctx)
	if req == nil {
		return nil, fmt.Errorf("no associated request")
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()

	identity, err := c.identifier.IdentifyAs(ctx, req, oauth.OpenIDTypeChecker)
	if identity == nil || err != nil {
		return nil, NewErrWebOidcUnableToIdentify("unable to obtain identity", err)
	}

	token, err := identity.(oauth.OpenIDIdentity).TokenSource().Token()
	if err != nil {
		return nil, NewErrWebOidcUnableToIdentify("unable to obtain token", err)
	}

	return map[string]string{
		"identity": token.AccessToken,
	}, nil
}

func (*WebOidcCredentials) RequireTransportSecurity() bool {
	return false
}
