package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	googleAuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleIssuer       = "https://accounts.google.com"
)

var googleScopes = []string{"openid", "email", "profile"}

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

func BuildGoogleAuthorizeURL(state string) string {
	params := url.Values{
		"client_id":              {os.Getenv("GOOGLE_CLIENT_ID")},
		"redirect_uri":           {os.Getenv("GOOGLE_REDIRECT_URI")},
		"response_type":          {"code"},
		"scope":                  {strings.Join(googleScopes, " ")},
		"state":                  {state},
		"include_granted_scopes": {"true"},
	}
	return googleAuthorizeURL + "?" + params.Encode()
}

func ExchangeGoogleCode(code string) (GoogleTokens, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("GOOGLE_CLIENT_ID")},
		"client_secret": {os.Getenv("GOOGLE_CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("GOOGLE_REDIRECT_URI")},
		"grant_type":    {"authorization_code"},
	}

	resp, err := httpClient.PostForm(googleTokenURL, form)
	if err != nil {
		return GoogleTokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return GoogleTokens{}, fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, body)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return GoogleTokens{}, err
	}
	return GoogleTokens{AccessToken: body.AccessToken, IDToken: body.IDToken}, nil
}

var (
	googleProvider     *oidc.Provider
	googleProviderErr  error
	googleProviderOnce sync.Once
)

func getGoogleProvider(ctx context.Context) (*oidc.Provider, error) {
	googleProviderOnce.Do(func() {
		googleProvider, googleProviderErr = oidc.NewProvider(ctx, googleIssuer)
	})
	return googleProvider, googleProviderErr
}

func VerifyGoogleIDToken(ctx context.Context, rawIDToken string) (GoogleIdentity, error) {
	provider, err := getGoogleProvider(ctx)
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("oidc provider init failed: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: os.Getenv("GOOGLE_CLIENT_ID")})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("id_token verify failed: %w", err)
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return GoogleIdentity{}, fmt.Errorf("claims decode failed: %w", err)
	}

	return GoogleIdentity{
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}
