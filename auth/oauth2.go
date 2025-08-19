package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OAuth2ClientCredentials struct {
	clientID     string
	clientSecret string
	tokenURL     string
	scopes       []string
	token        string
	expiry       time.Time
	mutex        sync.RWMutex
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewOAuth2ClientCredentials(clientID, clientSecret, tokenURL string, scopes []string) (*OAuth2ClientCredentials, error) {
	if clientID == "" || clientSecret == "" || tokenURL == "" {
		return nil, fmt.Errorf("clientID, clientSecret, and tokenURL are required")
	}
	
	return &OAuth2ClientCredentials{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		scopes:       scopes,
	}, nil
}

func (o *OAuth2ClientCredentials) Apply(req *http.Request) error {
	token, err := o.getValidToken()
	if err != nil {
		return fmt.Errorf("failed to get OAuth2 token: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (o *OAuth2ClientCredentials) getValidToken() (string, error) {
	o.mutex.RLock()
	if o.token != "" && time.Now().Before(o.expiry) {
		token := o.token
		o.mutex.RUnlock()
		return token, nil
	}
	o.mutex.RUnlock()
	
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	if o.token != "" && time.Now().Before(o.expiry) {
		return o.token, nil
	}
	
	return o.fetchToken()
}

func (o *OAuth2ClientCredentials) fetchToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", o.clientID)
	data.Set("client_secret", o.clientSecret)
	
	if len(o.scopes) > 0 {
		data.Set("scope", strings.Join(o.scopes, " "))
	}
	
	req, err := http.NewRequest("POST", o.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %s", resp.Status)
	}
	
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}
	
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}
	
	o.token = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		o.expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	} else {
		o.expiry = time.Now().Add(55 * time.Minute)
	}
	
	return o.token, nil
}