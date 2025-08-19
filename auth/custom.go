package auth

import (
	"net/http"
)

type CustomAuth struct {
	header string
	value  string
}

func NewCustomAuth(header, value string) *CustomAuth {
	return &CustomAuth{
		header: header,
		value:  value,
	}
}

func (c *CustomAuth) Apply(req *http.Request) error {
	if c.header != "" && c.value != "" {
		req.Header.Set(c.header, c.value)
	}
	return nil
}