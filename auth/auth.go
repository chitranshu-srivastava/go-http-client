package auth

import (
	"net/http"
)

type Authenticator interface {
	Apply(req *http.Request) error
}

type Config struct {
	Username     string
	Password     string
	BearerToken  string
	ClientID     string
	ClientSecret string
	TokenURL     string
	Scopes       []string
	CustomHeader string
	CustomValue  string
}

func NewAuthenticator(config Config) (Authenticator, error) {
	if config.Username != "" || config.Password != "" {
		return NewBasicAuth(config.Username, config.Password), nil
	}
	
	if config.BearerToken != "" {
		return NewBearerAuth(config.BearerToken), nil
	}
	
	if config.ClientID != "" && config.ClientSecret != "" && config.TokenURL != "" {
		return NewOAuth2ClientCredentials(config.ClientID, config.ClientSecret, config.TokenURL, config.Scopes)
	}
	
	if config.CustomHeader != "" && config.CustomValue != "" {
		return NewCustomAuth(config.CustomHeader, config.CustomValue), nil
	}
	
	return nil, nil
}