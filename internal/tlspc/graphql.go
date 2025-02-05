// Copyright (c) Venafi, Inc.
// SPDX-License-Identifier: MPL-2.0

package tlspc

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"terraform-provider-tlspc/internal/tlspc/graphql"

	gql "github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"
)

func (c *Client) GetGraphQLClient() gql.Client {
	httpClient := http.DefaultClient
	rt := WithHeader(httpClient.Transport)
	rt.Set("tppl-api-key", c.apikey)
	rt.Header.Set("User-Agent", "terraform-provider-tlspc/"+c.version)
	httpClient.Transport = rt

	path := c.Path(`%s/graphql`)
	client := gql.NewClient(path, httpClient)

	return client
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(h.Header) == 0 {
		return h.rt.RoundTrip(req)
	}

	req = req.Clone(req.Context())
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

type CloudProviderGCP struct {
	ID                             string
	IssuerUrl                      string
	Name                           string
	Team                           string
	ServiceAccountEmail            string
	ProjectNumber                  int64
	WorkloadIdentityPoolId         string
	WorkloadIdentityPoolProviderId string
}

func (c *Client) CreateCloudProviderGCP(ctx context.Context, p CloudProviderGCP) (*CloudProviderGCP, error) {
	gql := c.GetGraphQLClient()

	teamid, err := uuid.Parse(p.Team)
	if err != nil {
		return nil, err
	}

	pn := strconv.FormatInt(p.ProjectNumber, 10)

	resp, err := graphql.NewGCPProvider(ctx, gql,
		p.Name,
		teamid,
		p.ServiceAccountEmail,
		pn,
		p.WorkloadIdentityPoolId,
		p.WorkloadIdentityPoolProviderId,
	)

	if err != nil {
		return nil, err
	}

	cfg, ok := resp.CreateCloudProvider.Configuration.(*graphql.NewGCPProviderCreateCloudProviderConfigurationCloudProviderGCPConfiguration)
	if !ok {
		return nil, errors.New("No GCP CloudProvider Configuration returned")
	}

	cpn, err := strconv.ParseInt(cfg.ProjectNumber, 10, 64)
	if err != nil {
		return nil, err
	}

	created := CloudProviderGCP{
		ID:                             resp.CreateCloudProvider.Id.String(),
		IssuerUrl:                      cfg.IssuerUrl,
		Name:                           resp.CreateCloudProvider.Name,
		Team:                           resp.CreateCloudProvider.Team.Id,
		ProjectNumber:                  cpn,
		ServiceAccountEmail:            cfg.ServiceAccountEmail,
		WorkloadIdentityPoolId:         cfg.WorkloadIdentityPoolId,
		WorkloadIdentityPoolProviderId: cfg.WorkloadIdentityPoolProviderId,
	}

	return &created, nil
}

func (c *Client) GetCloudProviderGCP(ctx context.Context, id string) (*CloudProviderGCP, error) {
	gql := c.GetGraphQLClient()

	// No mechanism to retrieve by Id :(
	// (CloudProviderDetails only works if we get a valid connection - we definitely want to be able to retrieve poorly/incomplete setup)
	resp, err := graphql.GCPProviders(ctx, gql)

	if err != nil {
		return nil, err
	}

	var found *graphql.GCPProvidersCloudProvidersCloudProviderConnectionNodesCloudProvider

	for _, v := range resp.CloudProviders.Nodes {
		if v.Id.String() == id {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, errors.New("GCP CloudProvider not found")
	}
	cfg, ok := found.Configuration.(*graphql.GCPProvidersCloudProvidersCloudProviderConnectionNodesCloudProviderConfigurationCloudProviderGCPConfiguration)
	if !ok {
		return nil, errors.New("Expected GCP Configuration not found")
	}

	cpn, err := strconv.ParseInt(cfg.ProjectNumber, 10, 64)
	if err != nil {
		return nil, err
	}

	p := CloudProviderGCP{
		ID:                             found.Id.String(),
		IssuerUrl:                      cfg.IssuerUrl,
		Name:                           found.Name,
		Team:                           found.Team.Id,
		ProjectNumber:                  cpn,
		ServiceAccountEmail:            cfg.ServiceAccountEmail,
		WorkloadIdentityPoolId:         cfg.WorkloadIdentityPoolId,
		WorkloadIdentityPoolProviderId: cfg.WorkloadIdentityPoolProviderId,
	}

	return &p, nil
}

func (c *Client) UpdateCloudProviderGCP(ctx context.Context, p CloudProviderGCP) (*CloudProviderGCP, error) {
	gql := c.GetGraphQLClient()

	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}

	teamid, err := uuid.Parse(p.Team)
	if err != nil {
		return nil, err
	}

	pn := strconv.FormatInt(p.ProjectNumber, 10)

	resp, err := graphql.UpdateGCPProvider(ctx, gql,
		id,
		p.Name,
		teamid,
		pn,
		p.WorkloadIdentityPoolId,
		p.WorkloadIdentityPoolProviderId,
	)
	if err != nil {
		return nil, err
	}
	cfg, ok := resp.UpdateCloudProvider.Configuration.(*graphql.UpdateGCPProviderUpdateCloudProviderConfigurationCloudProviderGCPConfiguration)
	if !ok {
		return nil, errors.New("Error updating GCP Cloud Provider")
	}

	cpn, err := strconv.ParseInt(cfg.ProjectNumber, 10, 64)
	if err != nil {
		return nil, err
	}

	updated := CloudProviderGCP{
		ID:                             resp.UpdateCloudProvider.Id.String(),
		IssuerUrl:                      cfg.IssuerUrl,
		Name:                           resp.UpdateCloudProvider.Name,
		Team:                           resp.UpdateCloudProvider.Team.Id,
		ProjectNumber:                  cpn,
		ServiceAccountEmail:            cfg.ServiceAccountEmail,
		WorkloadIdentityPoolId:         cfg.WorkloadIdentityPoolId,
		WorkloadIdentityPoolProviderId: cfg.WorkloadIdentityPoolProviderId,
	}

	return &updated, nil
}

func (c *Client) DeleteCloudProviderGCP(ctx context.Context, id string) error {
	gql := c.GetGraphQLClient()

	deleteId, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	_, err = graphql.DeleteGCPProvider(ctx, gql, deleteId)

	return err
}

func (c *Client) GetCloudProviderGCPValidation(ctx context.Context, id string) (bool, error) {
	gql := c.GetGraphQLClient()

	cpId, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	resp, err := graphql.GetGCPProviderDetails(ctx, gql, cpId)
	if err != nil {
		return false, err
	}

	details, ok := resp.CloudProviderDetails.(*graphql.GetGCPProviderDetailsCloudProviderDetailsGCPProviderDetails)
	if !ok {
		return false, errors.New("Error retrieving GCP CloudProvider status")
	}

	return details.CloudProvider.Status == graphql.CloudProviderStatusValidated, nil
}

func (c *Client) ValidateCloudProviderGCP(ctx context.Context, id string) (bool, error) {
	gql := c.GetGraphQLClient()

	cpId, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	resp, err := graphql.ValidateGCPProvider(ctx, gql, cpId)
	if err != nil {
		return false, err
	}

	return resp.ValidateCloudProvider.Result == graphql.CloudProviderStatusValidated, nil
}
