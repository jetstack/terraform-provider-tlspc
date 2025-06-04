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
	version  string
}

func NewClient(apikey, endpoint, version string) (*Client, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	return &Client{
		apikey:   apikey,
		endpoint: endpoint,
		version:  version,
	}, nil
}

func (c *Client) doRequest(method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tppl-api-key", c.apikey)
	req.Header.Set("User-Agent", "terraform-provider-tlspc/"+c.version)

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

func (c *Client) Put(path string, body []byte) (*http.Response, error) {
	return c.doRequest("PUT", path, body)
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
	ID      string           `json:"id"`
	Name    string           `json:"productName"`
	Details CAProductDetails `json:"productDetails"`
}

type CAProductDetails struct {
	Template CAProductTemplate `json:"productTemplate"`
}

type CAProductTemplate struct {
	CertificateAuthority string   `json:"certificateAuthority"`
	ProductName          string   `json:"productName"`
	ProductTypes         []string `json:"productTypes"`
	ValidityPeriod       string   `json:"validityPeriod"`
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

func (c *Client) GetCAProductOption(kind, name, option string) (*CAProductOption, *CAAccount, error) {
	path := c.Path(`%s/v1/certificateauthorities/` + kind + "/accounts")

	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting ca product: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var accounts caAccounts
	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return nil, nil, fmt.Errorf("Error decoding response: %s", string(body))
	}
	for _, acc := range accounts.Accounts {
		acct := acc.Account
		if acct.Name != name {
			continue
		}
		for _, opt := range acc.ProductOptions {
			if opt.Name == option {
				return &opt, &acct, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("Specified CA product option not found.")
}

func (c *Client) GetCAProductOptionByID(kind, option_id string) (*CAProductOption, error) {
	path := c.Path(`%s/v1/certificateauthorities/` + kind + "/accounts")

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting ca product: %s", err)
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
		for _, opt := range acc.ProductOptions {
			if opt.ID == option_id {
				return &opt, nil
			}
		}
	}

	return nil, fmt.Errorf("Specified CA product option not found.")
}

type KeyType struct {
	Type       string   `json:"keyType"`
	KeyLengths []int32  `json:"keyLengths,omitempty"`
	KeyCurves  []string `json:"keyCurves,omitempty"`
}

type CertificateTemplate struct {
	ID                                  string            `json:"id,omitempty"`
	Name                                string            `json:"name"`
	CertificateAuthorityType            string            `json:"certificateAuthority"`
	CertificateAuthorityProductOptionID string            `json:"certificateAuthorityProductOptionId"`
	KeyReuse                            bool              `json:"keyReuse"`
	KeyTypes                            []KeyType         `json:"keyTypes"`
	Product                             CAProductTemplate `json:"product"`
	SANRegexes                          []string          `json:"sanRegexes"`
	SubjectCNRegexes                    []string          `json:"subjectCNRegexes"`
	SubjectCValues                      []string          `json:"subjectCValues"`
	SubjectLRegexes                     []string          `json:"subjectLRegexes"`
	SubjectORegexes                     []string          `json:"subjectORegexes"`
	SubjectOURegexes                    []string          `json:"subjectOURegexes"`
	SubjectSTRegexes                    []string          `json:"subjectSTRegexes"`
}

type certificateTemplates struct {
	Templates []CertificateTemplate `json:"certificateIssuingTemplates"`
}

func (c *Client) CreateCertificateTemplate(ct CertificateTemplate) (*CertificateTemplate, error) {
	path := c.Path(`%s/v1/certificateissuingtemplates`)

	body, err := json.Marshal(ct)
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
	var created certificateTemplates
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if len(created.Templates) != 1 {
		return nil, fmt.Errorf("Unexpected number of templates returned (%d): %s %s", len(created.Templates), string(respBody), string(body))
	}
	if created.Templates[0].ID == "" {
		return nil, fmt.Errorf("Didn't create a template; response was: %s", string(respBody))
	}

	return &created.Templates[0], nil
}

func (c *Client) GetCertificateTemplate(id string) (*CertificateTemplate, error) {
	path := c.Path(`%s/v1/certificateissuingtemplates/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting certificate template: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var ct CertificateTemplate
	err = json.Unmarshal(respBody, &ct)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if ct.ID == "" {
		return nil, fmt.Errorf("Didn't find a Certificate Template; response was: %s", string(respBody))
	}

	return &ct, nil
}

func (c *Client) UpdateCertificateTemplate(ct CertificateTemplate) (*CertificateTemplate, error) {
	id := ct.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	ct.ID = ""
	path := c.Path(`%s/v1/certificateissuingtemplates/` + id)

	body, err := json.Marshal(ct)
	if err != nil {
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Put(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error patching request: %s", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Failed to update certificate template; response was: %s", string(respBody))
	}

	var updated CertificateTemplate
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteCertificateTemplate(id string) error {
	path := c.Path(`%s/v1/certificateissuingtemplates/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete certificate template; response was: %s", string(respBody))
	}

	return nil
}

type OwnerAndType struct {
	ID   string `json:"ownerId"`
	Type string `json:"ownerType"`
}

type Application struct {
	ID                   string            `json:"id,omitempty"`
	Name                 string            `json:"name"`
	Owners               []OwnerAndType    `json:"ownerIdsAndTypes"`
	CertificateTemplates map[string]string `json:"certificateIssuingTemplateAliasIdMap"`
	FQDNs                []string          `json:"fqdns"`
	InternalPorts        []string          `json:"internalPorts"`
	IPRanges             []string          `json:"ipRanges"`
	Ports                []string          `json:"ports"`
}

type applications struct {
	Applications []Application `json:"applications"`
}

func (c *Client) CreateApplication(app Application) (*Application, error) {
	path := c.Path(`%s/outagedetection/v1/applications`)

	body, err := json.Marshal(app)
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
	var created applications
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if len(created.Applications) != 1 {
		return nil, fmt.Errorf("Unexpected number of applications returned (%d): %s %s", len(created.Applications), string(respBody), string(body))
	}
	if created.Applications[0].ID == "" {
		return nil, fmt.Errorf("Didn't create a application; response was: %s", string(respBody))
	}

	return &created.Applications[0], nil
}

func (c *Client) GetApplication(id string) (*Application, error) {
	path := c.Path(`%s/outagedetection/v1/applications/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting application: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var app Application
	err = json.Unmarshal(respBody, &app)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if app.ID == "" {
		return nil, fmt.Errorf("Didn't find a Application; response was: %s", string(respBody))
	}

	return &app, nil
}

func (c *Client) UpdateApplication(app Application) (*Application, error) {
	id := app.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	app.ID = ""
	path := c.Path(`%s/outagedetection/v1/applications/` + id)

	body, err := json.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("Error encoding request: %s", err)
	}

	resp, err := c.Put(path, body)
	if err != nil {
		return nil, fmt.Errorf("Error patching request: %s", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Failed to update application; response was: %s", string(respBody))
	}

	var updated Application
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteApplication(id string) error {
	path := c.Path(`%s/outagedetection/v1/applications/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete certificate template; response was: %s", string(respBody))
	}

	return nil
}

type CertificateTemplates struct {
	Templates []CertificateTemplate `json:"certificateIssuingTemplates"`
}

func (c *Client) GetCertTemplates() ([]CertificateTemplate, error) {
	path := c.Path(`%s/v1/certificateissuingtemplates/`)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting certificate template: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var ct CertificateTemplates
	err = json.Unmarshal(respBody, &ct)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return ct.Templates, nil

}

type FireflyConfig struct {
	ID                string          `json:"id,omitempty"`
	Name              string          `json:"name"`
	PolicyIds         []string        `json:"policyIds"`
	Policies          []FireflyPolicy `json:"policies,omitempty"`
	ServiceAccountIds []string        `json:"serviceAccountIds"`
	SubCAProviderId   string          `json:"subCaProviderId"`
	MinTLSVersion     string          `json:"minTlsVersion"`
	//ClientAuthentication ClientAuthentication `json:"clientAuthentication,omitempty"`
	CloudProviders CloudProviders `json:"cloudProviders"`
}

type CloudProviders struct{}

type ClientAuthentication struct {
	Type string `json:"type,omitempty"`
}

func (c *Client) CreateFireflyConfig(ff FireflyConfig) (*FireflyConfig, error) {
	path := c.Path(`%s/v1/distributedissuers/configurations`)

	body, err := json.Marshal(ff)
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
	var created FireflyConfig
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if created.ID == "" {
		return nil, fmt.Errorf("Didn't create a Firefly Config; response was: %s", string(respBody))
	}

	return &created, nil
}

func (c *Client) GetFireflyConfig(id string) (*FireflyConfig, error) {
	path := c.Path(`%s/v1/distributedissuers/configurations/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting Firefly Config: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var got FireflyConfig
	err = json.Unmarshal(respBody, &got)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if got.ID == "" {
		return nil, fmt.Errorf("Didn't find a Firefly Config; response was: %s", string(respBody))
	}

	return &got, nil
}

func (c *Client) UpdateFireflyConfig(ff FireflyConfig) (*FireflyConfig, error) {
	id := ff.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	ff.ID = ""
	path := c.Path(`%s/v1/distributedissuers/configurations/` + id)

	body, err := json.Marshal(ff)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Failed to update Firefly Config; response was: %s", string(respBody))
	}

	var updated FireflyConfig
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteFireflyConfig(id string) error {
	path := c.Path(`%s/v1/distributedissuers/configurations/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete Firefly Config; response was: %s", string(respBody))
	}

	return nil
}

type FireflySubCAProvider struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	CAType            string `json:"caType"`
	CAAccountID       string `json:"caAccountId"`
	CAProductOptionID string `json:"caProductOptionId"`
	CommonName        string `json:"commonName"`
	KeyAlgorithm      string `json:"keyAlgorithm"`
	ValidityPeriod    string `json:"validityPeriod"`
}

func (c *Client) CreateFireflySubCAProvider(ff FireflySubCAProvider) (*FireflySubCAProvider, error) {
	path := c.Path(`%s/v1/distributedissuers/subcaproviders`)

	body, err := json.Marshal(ff)
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
	var created FireflySubCAProvider
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if created.ID == "" {
		return nil, fmt.Errorf("Didn't create a Firefly SubCAProvider; response was: %s", string(respBody))
	}

	return &created, nil
}

func (c *Client) GetFireflySubCAProvider(id string) (*FireflySubCAProvider, error) {
	path := c.Path(`%s/v1/distributedissuers/subcaproviders/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting Firefly SubCAProvider: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var got FireflySubCAProvider
	err = json.Unmarshal(respBody, &got)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if got.ID == "" {
		return nil, fmt.Errorf("Didn't find a Firefly SubCAProvider; response was: %s", string(respBody))
	}

	return &got, nil
}

func (c *Client) UpdateFireflySubCAProvider(ff FireflySubCAProvider) (*FireflySubCAProvider, error) {
	id := ff.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	ff.ID = ""
	path := c.Path(`%s/v1/distributedissuers/subcaproviders/` + id)

	body, err := json.Marshal(ff)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Failed to update Firefly SubCAProvider; response was: %s", string(respBody))
	}

	var updated FireflySubCAProvider
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteFireflySubCAProvider(id string) error {
	path := c.Path(`%s/v1/distributedissuers/subcaproviders/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete Firefly SubCAProvider; response was: %s", string(respBody))
	}

	return nil
}

type FireflyPolicy struct {
	ID                string               `json:"id,omitempty"`
	Name              string               `json:"name"`
	ExtendedKeyUsages []string             `json:"extendedKeyUsages"`
	KeyAlgorithm      KeyAlgorithm         `json:"keyAlgorithm"`
	KeyUsages         []string             `json:"keyUsages"`
	SANs              SANs                 `json:"sans"`
	Subject           FireflyPolicySubject `json:"subject"`
	ValidityPeriod    string               `json:"validityPeriod"`
}

type KeyAlgorithm struct {
	AllowedValues []string `json:"allowedValues"`
	DefaultValue  string   `json:"defaultValue"`
}

type SANs struct {
	DNSNames    PolicyDetails `json:"dnsNames"`
	IPAddresses PolicyDetails `json:"ipAddresses"`
	RFC822Names PolicyDetails `json:"rfc822Names"`
	URIs        PolicyDetails `json:"uniformResourceIdentifiers"`
}

type PolicyDetails struct {
	AllowedValues  []string `json:"allowedValues"`
	DefaultValues  []string `json:"defaultValues"`
	MaxOccurrences int32    `json:"maxOccurrences"`
	MinOccurrences int32    `json:"minOccurrences"`
	Type           string   `json:"type"`
}

type FireflyPolicySubject struct {
	CommonName         PolicyDetails `json:"commonName"`
	Country            PolicyDetails `json:"country"`
	Locality           PolicyDetails `json:"locality"`
	Organization       PolicyDetails `json:"organization"`
	OrganizationalUnit PolicyDetails `json:"organizationalUnit"`
	StateOrProvince    PolicyDetails `json:"stateOrProvince"`
}

func (c *Client) CreateFireflyPolicy(ff FireflyPolicy) (*FireflyPolicy, error) {
	path := c.Path(`%s/v1/distributedissuers/policies`)

	body, err := json.Marshal(ff)
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
	var created FireflyPolicy
	err = json.Unmarshal(respBody, &created)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if created.ID == "" {
		return nil, fmt.Errorf("Didn't create a Firefly Policy; response was: %s", string(respBody))
	}

	return &created, nil
}

func (c *Client) GetFireflyPolicy(id string) (*FireflyPolicy, error) {
	path := c.Path(`%s/v1/distributedissuers/policies/` + id)

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("Error getting Firefly Policy: %s", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}
	var got FireflyPolicy
	err = json.Unmarshal(respBody, &got)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}
	if got.ID == "" {
		return nil, fmt.Errorf("Didn't find a Firefly Policy; response was: %s", string(respBody))
	}

	return &got, nil
}

func (c *Client) UpdateFireflyPolicy(ff FireflyPolicy) (*FireflyPolicy, error) {
	id := ff.ID
	if id == "" {
		return nil, errors.New("Empty ID")
	}
	ff.ID = ""
	path := c.Path(`%s/v1/distributedissuers/policies/` + id)

	body, err := json.Marshal(ff)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Failed to update Firefly Policy; response was: %s", string(respBody))
	}

	var updated FireflyPolicy
	err = json.Unmarshal(respBody, &updated)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %s", string(respBody))
	}

	return &updated, nil
}

func (c *Client) DeleteFireflyPolicy(id string) error {
	path := c.Path(`%s/v1/distributedissuers/policies/` + id)

	resp, err := c.Delete(path, nil)
	if err != nil {
		return fmt.Errorf("Error with delete request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		// returning an error here anyway, no more information if we couldn't read the body
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete Firefly Policy; response was: %s", string(respBody))
	}

	return nil
}
