package service

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleIssuer = "https://accounts.google.com"

type GoogleTokens struct {
	AccessToken string
	IDToken     string
}

type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

func googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URI"),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func BuildGoogleAuthorizeURL(state string) string {
	return googleOAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("include_granted_scopes", "true"))
}

func ExchangeGoogleCode(ctx context.Context, code string) (*GoogleTokens, error) {
	token, err := googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("token response did not include an id_token")
	}

	return &GoogleTokens{AccessToken: token.AccessToken, IDToken: rawIDToken}, nil
}

var (
	googleProvider     *oidc.Provider
	googleProviderErr  error
	googleProviderOnce sync.Once
)

func googleOIDCProvider(ctx context.Context) (*oidc.Provider, error) {
	googleProviderOnce.Do(func() {
		googleProvider, googleProviderErr = oidc.NewProvider(ctx, googleIssuer)
	})
	return googleProvider, googleProviderErr
}

func VerifyGoogleIDToken(ctx context.Context, rawIDToken string) (*GoogleIdentity, error) {
	provider, err := googleOIDCProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("oidc provider init failed: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: os.Getenv("GOOGLE_OAUTH_CLIENT_ID")})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("id_token verify failed: %w", err)
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("claims decode failed: %w", err)
	}

	return &GoogleIdentity{
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}
