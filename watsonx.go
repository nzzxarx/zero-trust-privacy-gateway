/*
 * Author: Nizar Akkioui
 * Description: IBM watsonx.ai client with IAM token caching.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type WatsonxClient struct {
	APIKey     string
	ProjectID  string
	IAMURL     string
	Token      string
	Expiration time.Time
	mu         sync.Mutex
}

func NewWatsonxClient(apiKey, projectID string) *WatsonxClient {
	return &WatsonxClient{
		APIKey:    apiKey,
		ProjectID: projectID,
		IAMURL:    "https://iam.cloud.ibm.com/identity/token",
	}
}

// GetToken fetches a new IAM token or returns the cached one if valid
func (c *WatsonxClient) GetToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.Expiration) {
		return c.Token, nil // Return cached token
	}

	return c.fetchTokenFromIBM()
}

func (c *WatsonxClient) fetchTokenFromIBM() (string, error) {
	// Payload for IBM IAM (Simplified for this file)
	payload := []byte(fmt.Sprintf("grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=%s", c.APIKey))
	req, _ := http.NewRequest("POST", c.IAMURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	c.Token = result.AccessToken
	// Subtract 1 minute as buffer
	c.Expiration = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)

	return c.Token, nil
}
