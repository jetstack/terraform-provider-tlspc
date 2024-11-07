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
		return nil, fmt.Errorf("Error getting user: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var users Users
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(body))
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
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error posting request: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var created Team
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if created.ID == "" {
		return nil, fmt.Errorf("Didn't create a team; response was: %s", string(respBody))
	}

	return &created, nil
}

func (c *Client) GetTeam(id string) (*Team, error) {
	path := c.Path(`%s/v1/teams/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting team: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var team Team
	err = json.Unmarshal(respBody, &team)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if team.ID == "" {
		return nil, fmt.Errorf("Didn't find a Team; response was: %s", string(respBody))
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
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Patch(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error patching request: %s", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to update Team; response was: %s", string(respBody))
	}
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if updated.ID == "" {
		return nil, fmt.Errorf("Didn't get a Team ID; response was: %s", string(respBody))
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
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error posting request: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if updated.ID == "" {
		return nil, fmt.Errorf("Didn't get a Team ID; response was: %s", string(respBody))
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
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Delete(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error with delete request: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var updated Team
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if updated.ID == "" {
		return nil, fmt.Errorf("Didn't get a Team ID; response was: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteTeam(id string) error {
	path := c.Path(`%s/v1/teams/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	// https://developer.venafi.com/tlsprotectcloud/reference/teams_delete says 204, but we get a 200 back
	// so accept either, in case behaviour gets fixed to match the docs in the future
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete team; response was: %s", string(respBody))
	}

	return nil
}

type ServiceAccount struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name"`
	Owner              string   `json:"owner"`
	Scopes             []string `json:"scopes"`
	CredentialLifetime int32    `json:"credentialLifetime,omitempty"`
	PublicKey          string   `json:"publicKey,omitempty"`
	AuthenticationType string   `json:"authenticationType,omitempty"`
	OciAccountName     string   `json:"ociAccountName,omitempty"`
	OciRegistryToken   string   `json:"ociRegistryToken,omitempty"`
	JwksURI            string   `json:"jwksURI,omitempty"`
	IssuerURL          string   `json:"issuerURL,omitempty"`
	Audience           string   `json:"audience,omitempty"`
	Subject            string   `json:"subject,omitempty"`
	Applications       []string `json:"applications,omitempty"`
}

func (c *Client) CreateServiceAccount(sa ServiceAccount) (*ServiceAccount, error) {
	path := c.Path(`%s/v1/serviceaccounts`)

	body, err := json.Marshal(sa)
	if err != nil {
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error posting request: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var created ServiceAccount
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if created.ID == "" {
		return nil, fmt.Errorf("Didn't create a service account; response was: %s", string(respBody))
	}

	return &created, nil
}

func (c *Client) GetServiceAccount(id string) (*ServiceAccount, error) {
	path := c.Path(`%s/v1/serviceaccounts/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting service account: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var sa ServiceAccount
	err = json.Unmarshal(respBody, &sa)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if sa.ID == "" {
		return nil, fmt.Errorf("Didn't find a Service Account; response was: %s", string(respBody))
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
		return fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Patch(path, body)
	if err != nil {
		return fmt.Errorf("Error patching request: %s", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to update Service Account; response was: %s", string(respBody))
	}

	return nil
}

func (c *Client) DeleteServiceAccount(id string) error {
	path := c.Path(`%s/v1/serviceaccounts/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete Service Account; response was: %s", string(respBody))
	}

	return nil
}

type Plugin struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"pluginType"`
	Manifest any    `json:"manifest"`
}

type plugins struct {
	Plugins []Plugin `json:"plugins"`
}

func (c *Client) CreatePlugin(p Plugin) (*Plugin, error) {
	path := c.Path(`%s/v1/plugins`)

	body, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Post(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error posting request: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var created plugins
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if len(created.Plugins) != 1 {
		return nil, fmt.Errorf("Unexpected number of plugins returned (%d): %s", len(created.Plugins), string(respBody))
	}
	if created.Plugins[0].ID == "" {
		return nil, fmt.Errorf("Didn't create a plugin; response was: %s", string(respBody))
	}

	return &created.Plugins[0], nil
}

func (c *Client) GetPlugin(id string) (*Plugin, error) {
	path := c.Path(`%s/v1/plugins/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting plugin: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var plugin Plugin
	err = json.Unmarshal(respBody, &plugin)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if plugin.ID == "" {
		return nil, fmt.Errorf("Didn't find a Plugin; response was: %s", string(respBody))
	}

	return &plugin, nil
}

func (c *Client) UpdatePlugin(p Plugin) error {
	id := p.ID
	if id == "" {
		return errors.New("Empty ID")
	}
	p.ID = ""
	path := c.Path(`%s/v1/plugins/` + id)

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Patch(path, body)
	if err != nil {
		return fmt.Errorf("Error patching request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to update Plugin; response was: %s", string(respBody))
	}

	return nil
}

func (c *Client) DeletePlugin(id string) error {
	path := c.Path(`%s/v1/plugins/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete Plugin; response was: %s", string(respBody))
	}

	return nil
}

type CAProductOption struct {
	ID   string `json:"id"`
	Name string `json:"productName"`
}

type CAAccount struct {
	ID   string `json:"id"`
	Name string `json:"key"`
}

type caAccounts struct {
	Accounts []caAccount `json:"accounts"`
}

type caAccount struct {
	Account        CAAccount         `json:"account"`
	ProductOptions []CAProductOption `json:"productOptions"`
}

func (c *Client) GetCAProductOption(kind, name, option string) (*CAProductOption, error) {
	path := c.Path(`%s/v1/certificateauthorities/` + kind + "/accounts")

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting user: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var accounts caAccounts
	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(body))
	}
	for _, acc := range accounts.Accounts {
		acct := acc.Account
		if acct.Name != name {
			continue
		}
		for _, opt := range acc.ProductOptions {
			if opt.Name == option {
				return &opt, nil
			}
		}
	}

	return nil, fmt.Errorf("Specified CA product option not found.")
}
