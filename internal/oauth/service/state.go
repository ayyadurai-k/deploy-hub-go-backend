package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	stateVersion = "oauth.state.v3"
	stateTTL     = 10 * time.Minute
)

var (
	ErrStateMissing  = errors.New("missing state")
	ErrStateInvalid  = errors.New("state signature invalid or expired")
	ErrStateMismatch = errors.New("state mismatch")
)

type StatePayload struct {
	Nonce    string
	Provider string
}

type stateClaims struct {
	Nonce    string `json:"n"`
	Provider string `json:"p"`
	jwt.RegisteredClaims
}

func stateSecret() []byte {
	return []byte(os.Getenv("OAUTH_STATE_SECRET"))
}

func IssueState(provider string) (echoedNonce string, signedEnvelope string, err error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, stateClaims{
		Nonce:    nonce,
		Provider: provider,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    stateVersion,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(stateTTL)),
		},
	})

	signedEnvelope, err = token.SignedString(stateSecret())
	if err != nil {
		return "", "", err
	}
	return nonce, signedEnvelope, nil
}

func VerifyState(echoedNonce, signedEnvelope, expectedProvider string) (StatePayload, error) {
	if echoedNonce == "" || signedEnvelope == "" {
		return StatePayload{}, ErrStateMissing
	}

	var claims stateClaims
	_, err := jwt.ParseWithClaims(
		signedEnvelope,
		&claims,
		func(token *jwt.Token) (any, error) { return stateSecret(), nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(stateVersion),
	)
	if err != nil {
		return StatePayload{}, ErrStateInvalid
	}

	if claims.Provider != expectedProvider {
		return StatePayload{}, ErrStateMismatch
	}
	if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(echoedNonce)) != 1 {
		return StatePayload{}, ErrStateMismatch
	}

	return StatePayload{Nonce: claims.Nonce, Provider: claims.Provider}, nil
}

func generateNonce() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
