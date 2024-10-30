// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package tlspc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const DefaultEndpoint = "https://api.venafi.cloud"

type Client struct {
	apikey   string
	endpoint string
}

func NewClient(apikey, endpoint string) (*Client, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	return &Client{
		apikey:   apikey,
		endpoint: endpoint,
	}, nil
}

func (c *Client) doRequest(method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tppl-api-key", c.apikey)

	client := http.Client{}
	return client.Do(req)
}

func (c *Client) Path(tmpl string) string {
	return fmt.Sprintf(tmpl, c.endpoint)
}

func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", path, nil)
}

func (c *Client) Post(path string, body []byte) (*http.Response, error) {
	return c.doRequest("POST", path, body)
}

type User struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

type Users struct {
	Users []User `json:"users"`
}

func (c *Client) GetUser(email string) (*User, error) {
	path := c.Path(`%s/v1/users/username/` + email)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var users Users
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}
	if len(users.Users) != 1 {
		return nil, fmt.Errorf("Unexpected number of users returned (%d)", len(users.Users))
	}

	return &users.Users[0], nil
}

type Team struct {
	ID      string   `json:"id,omitempty"`
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Owners  []string `json:"owners"`
	Members []string `json:"members"`
}

func (c *Client) CreateTeam(team Team) (*Team, error) {
	path := c.Path(`%s/v1/teams`)

	body, err := json.Marshal(team)
	if err != nil {
		return nil, err
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var created Team
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, err
	}
	if created.ID == "" {
		return nil, errors.New("Didn't create a team??" + string(body))
	}

	return &created, nil
}

type ServiceAccount struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name"`
	Owner              string   `json:"owner"`
	Scopes             []string `json:"scopes"`
	CredentialLifetime int32    `json:"credentialLifetime"`
	PublicKey          string   `json:"publicKey"`
	AuthenticationType string   `json:"authenticationType"`
}

func (c *Client) CreateServiceAccount(sa ServiceAccount) (*ServiceAccount, error) {
	path := c.Path(`%s/v1/serviceaccounts`)

	body, err := json.Marshal(sa)
	if err != nil {
		return nil, err
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var created ServiceAccount
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, err
	}
	if created.ID == "" {
		return nil, errors.New("Didn't create a Service Account??" + string(body))
	}

	return &created, nil
}
