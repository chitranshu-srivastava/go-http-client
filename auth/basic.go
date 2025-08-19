package auth

import (
	"encoding/base64"
	"net/http"
)

type BasicAuth struct {
	username string
	password string
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}

func (b *BasicAuth) Apply(req *http.Request) error {
	if b.username != "" || b.password != "" {
		req.SetBasicAuth(b.username, b.password)
	}
	return nil
}

func (b *BasicAuth) EncodeCredentials() string {
	credentials := b.username + ":" + b.password
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}