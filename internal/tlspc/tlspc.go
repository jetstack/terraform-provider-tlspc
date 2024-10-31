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

func (c *Client) Patch(path string, body []byte) (*http.Response, error) {
	return c.doRequest("PATCH", path, body)
}

func (c *Client) Delete(path string, body []byte) (*http.Response, error) {
	return c.doRequest("DELETE", path, body)
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

func (c *Client) GetTeam(id string) (*Team, error) {
	path := c.Path(`%s/v1/teams/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var team Team
	err = json.Unmarshal(respBody, &team)
	if err != nil {
		return nil, err
	}
	if team.ID == "" {
		return nil, errors.New("Didn't find a Team")
	}

	return &team, nil
}

type updateTeam struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func (c *Client) UpdateTeam(team Team) (*Team, error) {
	id := team.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	team.ID = ""
	path := c.Path(`%s/v1/teams/` + id)

	update := updateTeam{
		Name: team.Name,
		Role: team.Role,
	}
	body, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	resp, err := c.Patch(path, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to update team")
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, err
	}
	if updated.ID == "" {
		return nil, errors.New("Didn't create a team??" + string(body))
	}

	return &updated, nil
}

type updateTeamOwners struct {
	Owners []string `json:"owners"`
}

func (c *Client) AddTeamOwners(id string, owners []string) (*Team, error) {
	path := c.Path(`%s/v1/teams/` + id + `/owners`)
	update := updateTeamOwners{
		Owners: owners,
	}

	body, err := json.Marshal(update)
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
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, err
	}
	if updated.ID == "" {
		return nil, errors.New("Didn't create a team??" + string(body))
	}

	return &updated, nil
}

func (c *Client) RemoveTeamOwners(id string, owners []string) (*Team, error) {
	path := c.Path(`%s/v1/teams/` + id + `/owners`)
	update := updateTeamOwners{
		Owners: owners,
	}

	body, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	resp, err := c.Delete(path, body)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, err
	}
	if updated.ID == "" {
		return nil, errors.New("Didn't create a team??" + string(body))
	}

	return &updated, nil
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

func (c *Client) GetServiceAccount(id string) (*ServiceAccount, error) {
	path := c.Path(`%s/v1/serviceaccounts/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sa ServiceAccount
	err = json.Unmarshal(respBody, &sa)
	if err != nil {
		return nil, err
	}
	if sa.ID == "" {
		return nil, errors.New("Didn't find a Service Account")
	}

	return &sa, nil
}

func (c *Client) UpdateServiceAccount(sa ServiceAccount) error {
	id := sa.ID
	if id == "" {
		return errors.New("Empty ID")
	}
	sa.ID = ""
	path := c.Path(`%s/v1/serviceaccounts/` + id)

	body, err := json.Marshal(sa)
	if err != nil {
		return err
	}

	resp, err := c.Patch(path, body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New("Failed to update service account")
	}

	return nil
}
