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

func (c *WebOidcCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	req := web.RequestFromContext(ctx)
	if req == nil {
		return nil, fmt.Errorf("no associated request")
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()

	identity, err := c.identifier.IdentifyAs(ctx, req, oauth.OpenIDTypeChecker)
	if identity == nil || err != nil {
		return nil, fmt.Errorf("unable to obtain identity: %#w", err)
	}

	token, err := identity.(oauth.OpenIDIdentity).TokenSource().Token()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain token: %#w", err)
	}

	return map[string]string{
		"identity": token.AccessToken,
	}, nil
}

func (*WebOidcCredentials) RequireTransportSecurity() bool {
	return false
}
