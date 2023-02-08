package portainerapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	portainer "github.com/portainer/portainer/api"
)

var defaultRequestTimeout = 30 * time.Second

type Client interface {
	Stacks(ctx context.Context) ([]Stack, error)
	Stack(ctx context.Context, stackID int) (*Stack, error)
	StackFileContent(ctx context.Context, stackID int) (string, error)
	StackImageStatus(ctx context.Context, stackID int) (string, error)
	UpdateStack(ctx context.Context, stackID int) error
}

type PortainerAPI struct {
	client *http.Client
	token  string
	host   string
}

func (c *PortainerAPI) do(ctx context.Context, method, endpoint string, body []byte) (*http.Response, error) {
	baseURL := fmt.Sprintf("%s/%s", c.host, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", c.token)

	return c.client.Do(req)
}

func (c *PortainerAPI) put(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	res, err := c.do(ctx, http.MethodPut, endpoint, body)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	respbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("non 2xx response code received: %d", res.StatusCode)
	}

	return respbody, nil
}

func (c *PortainerAPI) get(ctx context.Context, endpoint string) ([]byte, error) {
	res, err := c.do(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("non 2xx response code received: %d", res.StatusCode)
	}

	return body, nil
}

type Stack struct {
	portainer.Stack
	Webhook string `json:"Webhook"`
}

func (c *PortainerAPI) Stacks(ctx context.Context) ([]Stack, error) {
	response, err := c.get(ctx, "api/stacks")
	if err != nil {
		return nil, err
	}

	var results []Stack
	if err := json.Unmarshal(response, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *PortainerAPI) Stack(ctx context.Context, stackID int) (*Stack, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d", stackID))
	if err != nil {
		return nil, err
	}

	result := new(Stack)
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, err
	}

	return result, nil
}

type stackFileContentsResponse struct {
	StackFileContent string `json:"StackFileContent"`
}

func (c *PortainerAPI) StackFileContent(ctx context.Context, stackID int) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d/file", stackID))
	if err != nil {
		return "", err
	}

	result := new(stackFileContentsResponse)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	return result.StackFileContent, nil
}

type imageStatusResponse struct {
	Status string `json:"Status"`
}

func (c *PortainerAPI) StackImageStatus(ctx context.Context, stackID int) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d/images_status", stackID))
	if err != nil {
		return "", err
	}

	result := new(imageStatusResponse)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	return result.Status, nil
}

type updateGitStackRequest struct {
	Env                       []portainer.Pair `json:"env"`
	Prune                     bool             `json:"prune"`
	PullImage                 bool             `json:"pullImage"`
	RepositoryAuthentication  bool             `json:"repositoryAuthentication"`
	RepositoryGitCredentialID int              `json:"repositoryGitCredentialID"`
	RepositoryPassword        string           `json:"repositoryPassword"`
	RepositoryReferenceName   string           `json:"repositoryReferenceName"`
	RepositoryUsername        string           `json:"repositoryUsername"`
}

func (c *PortainerAPI) updateGitStack(ctx context.Context, stack *Stack) error {
	request := updateGitStackRequest{
		Env:                       stack.Env,
		Prune:                     true,
		PullImage:                 true,
		RepositoryAuthentication:  false,
		RepositoryGitCredentialID: 0,
		RepositoryPassword:        "",
		RepositoryReferenceName:   stack.GitConfig.ReferenceName,
		RepositoryUsername:        "",
	}

	if stack.GitConfig.Authentication != nil {
		request.RepositoryAuthentication = true
		request.RepositoryGitCredentialID = stack.GitConfig.Authentication.GitCredentialID
		request.RepositoryPassword = stack.GitConfig.Authentication.Password
		request.RepositoryUsername = stack.GitConfig.Authentication.Username
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if _, err := c.put(ctx, fmt.Sprintf("api/stacks/%d/git/redeploy", stack.ID), jsonRequest); err != nil {
		return err
	}
	return nil
}

type updateFileStackRequest struct {
	Env              []portainer.Pair `json:"env"`
	Prune            bool             `json:"prune"`
	PullImage        bool             `json:"pullImage"`
	StackFileContent string           `json:"stackFileContent"`
	Webhook          string           `json:"webhook"`
}

func (c *PortainerAPI) updateFileStack(ctx context.Context, stack *Stack) error {
	fileContents, err := c.StackFileContent(ctx, int(stack.ID))
	if err != nil {
		return err
	}

	request := updateFileStackRequest{
		Env:              stack.Env,
		Prune:            true,
		PullImage:        true,
		StackFileContent: fileContents,
		Webhook:          stack.Webhook,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if _, err := c.put(ctx, fmt.Sprintf("api/stacks/%d", stack.ID), jsonRequest); err != nil {
		return err
	}
	return nil
}

func (c *PortainerAPI) UpdateStack(ctx context.Context, stackID int) error {
	stack, err := c.Stack(ctx, stackID)
	if err != nil {
		return err
	}

	switch {
	case stack.GitConfig != nil:
		return c.updateGitStack(ctx, stack)
	default:
		return c.updateFileStack(ctx, stack)
	}
}

func NewPortainerAPIClient(token, host string) *PortainerAPI {
	httpClient := &http.Client{
		Timeout: defaultRequestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &PortainerAPI{
		client: httpClient,
		token:  token,
		host:   host,
	}
}
