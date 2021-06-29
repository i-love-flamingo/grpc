package grpc

import "flamingo.me/flamingo/v3/core/auth/oauth"

type keycloakStructure struct {
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

func KeycloakRealmRoles(token oauth.Identity) ([]string, error) {
	var tkn keycloakStructure
	if err := token.AccessTokenClaims(&tkn); err != nil {
		return nil, err
	}
	return tkn.RealmAccess.Roles, nil
}

func KeycloakClientRoles(token oauth.Identity, client string) ([]string, error) {
	var tkn keycloakStructure
	if err := token.AccessTokenClaims(&tkn); err != nil {
		return nil, err
	}
	return tkn.ResourceAccess[client].Roles, nil
}

func KeycloakClients(token oauth.Identity) (map[string][]string, error) {
	var tkn keycloakStructure
	if err := token.AccessTokenClaims(&tkn); err != nil {
		return nil, err
	}
	res := make(map[string][]string)
	for client, roles := range tkn.ResourceAccess {
		res[client] = roles.Roles
	}
	return res, nil
}
