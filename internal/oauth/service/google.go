package service

import (
	"net/url"
	"os"
	"strings"
)

const GOOGLE_AUTHORIZE_URL = "https://accounts.google.com/o/oauth2/v2/auth"

var (
	GOOGLE_OAUTH_CLIENT_ID     = os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	GOOGLE_OAUTH_CLIENT_SECRET = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	GOOGLE_OAUTH_REDIRECT_URI  = os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	SCOPES                     = []string{"openid", "email", "profile"}
)

func BuildGoogleAuthorizeURL(nonce string) string {
	params := url.Values{
		"client_id":     []string{GOOGLE_OAUTH_CLIENT_ID},
		"redirect_uri":  []string{GOOGLE_OAUTH_REDIRECT_URI},
		"response_type": []string{"code"},
		"scope":         []string{strings.Join(SCOPES, " ")},
		"state":         []string{nonce},
		"access_type":   []string{"offline"},
	}

	return GOOGLE_AUTHORIZE_URL + "?" + params.Encode()
}
