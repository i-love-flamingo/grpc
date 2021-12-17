package credentials

import (
	"context"
	"fmt"
	"sync"

	"flamingo.me/flamingo/v3/core/auth"
	"flamingo.me/flamingo/v3/core/auth/oauth"
	"flamingo.me/flamingo/v3/framework/web"
)

type WebOauth2Credentials struct {
	identifier *auth.WebIdentityService
}

func (c *WebOauth2Credentials) Inject(identifier *auth.WebIdentityService) {
	c.identifier = identifier
}

var tokenLock = new(sync.Mutex)

type ErrWebOauth2UnableToIdentify struct {
	msg string
	err error
}

func NewErrWebOauth2UnableToIdentify(msg string, err error) *ErrWebOauth2UnableToIdentify {
	return &ErrWebOauth2UnableToIdentify{
		msg: msg + ": " + err.Error(),
		err: err,
	}
}

func (e *ErrWebOauth2UnableToIdentify) Error() string {
	return e.msg
}

func (e *ErrWebOauth2UnableToIdentify) Unwrap() error {
	return e.err
}

func (c *WebOauth2Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	req := web.RequestFromContext(ctx)
	if req == nil {
		return nil, fmt.Errorf("no associated request")
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()

	identity, err := c.identifier.IdentifyAs(ctx, req, oauth.OAuthTypeChecker)
	if identity == nil || err != nil {
		return nil, NewErrWebOauth2UnableToIdentify("unable to obtain identity", err)
	}

	token, err := identity.(oauth.Identity).TokenSource().Token()
	if err != nil {
		return nil, NewErrWebOauth2UnableToIdentify("unable to obtain token", err)
	}

	return map[string]string{
		"authorization": token.AccessToken,
	}, nil
}

func (*WebOauth2Credentials) RequireTransportSecurity() bool {
	return false
}
