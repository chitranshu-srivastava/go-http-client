package auth

import (
	"net/http"
)

type BearerAuth struct {
	token string
}

func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		token: token,
	}
}

func (b *BearerAuth) Apply(req *http.Request) error {
	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}
	return nil
}