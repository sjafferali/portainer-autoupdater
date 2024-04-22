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

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog"
)

var defaultRequestTimeout = time.Minute * 2

type Client interface {
	Endpoints(ctx context.Context, ll zerolog.Logger) ([]portainer.Endpoint, error)
	Stacks(ctx context.Context, ll zerolog.Logger) ([]Stack, error)
	Stack(ctx context.Context, stackID int, ll zerolog.Logger) (*Stack, error)
	StackFileContent(ctx context.Context, stackID int, ll zerolog.Logger) (string, error)
	StackImageStatus(ctx context.Context, stackID int, ll zerolog.Logger) (string, error)
	UpdateStack(ctx context.Context, stackID int, ll zerolog.Logger) error
	UpdateService(ctx context.Context, serviceID string, endpoint int, ll zerolog.Logger) error
	ContainersForStack(ctx context.Context, stack Stack, ll zerolog.Logger) ([]dockertypes.Container, error)
	Containers(ctx context.Context, endpointID int, ll zerolog.Logger) ([]dockertypes.Container, error)
	ContainerImageStatus(ctx context.Context, containerID string, endpoint int, ll zerolog.Logger) (string, error)
	ServicesForStack(ctx context.Context, stack Stack, ll zerolog.Logger) ([]swarm.Service, error)
	Services(ctx context.Context, endpointID int, ll zerolog.Logger) ([]swarm.Service, error)
	ServiceImageStatus(ctx context.Context, serviceID string, endpoint int, ll zerolog.Logger) (string, error)
}

type PortainerAPI struct {
	client *http.Client
	token  string
	host   string
}

func (c *PortainerAPI) do(ctx context.Context, method, endpoint string, queryMap map[string]string, body []byte, ll zerolog.Logger) (*http.Response, error) {
	baseURL := fmt.Sprintf("%s/%s", c.host, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range queryMap {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", c.token)

	return c.client.Do(req)
}

func (c *PortainerAPI) put(ctx context.Context, endpoint string, body []byte, ll zerolog.Logger) ([]byte, error) {
	res, err := c.do(ctx, http.MethodPut, endpoint, nil, body, ll)
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

func (c *PortainerAPI) get(ctx context.Context, endpoint string, queryMap map[string]string, ll zerolog.Logger) ([]byte, error) {
	res, err := c.do(ctx, http.MethodGet, endpoint, queryMap, nil, ll)
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

func (c *PortainerAPI) Endpoints(ctx context.Context, ll zerolog.Logger) ([]portainer.Endpoint, error) {
	response, err := c.get(ctx, "api/endpoints", nil, ll)
	if err != nil {
		return nil, err
	}

	var results []portainer.Endpoint
	if err := json.Unmarshal(response, &results); err != nil {
		return nil, err
	}

	return results, nil
}

type Stack struct {
	portainer.Stack
	Webhook string `json:"Webhook"`
}

func (c *PortainerAPI) Stacks(ctx context.Context, ll zerolog.Logger) ([]Stack, error) {
	response, err := c.get(ctx, "api/stacks", nil, ll)
	if err != nil {
		return nil, err
	}

	var results []Stack
	if err := json.Unmarshal(response, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *PortainerAPI) Stack(ctx context.Context, stackID int, ll zerolog.Logger) (*Stack, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d", stackID), nil, ll)
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

func (c *PortainerAPI) StackFileContent(ctx context.Context, stackID int, ll zerolog.Logger) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d/file", stackID), nil, ll)
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

func (c *PortainerAPI) StackImageStatus(ctx context.Context, stackID int, ll zerolog.Logger) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/stacks/%d/images_status?refresh=true", stackID), nil, ll)
	if err != nil {
		return "", err
	}

	result := new(imageStatusResponse)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func (c *PortainerAPI) ContainerImageStatus(ctx context.Context, containerID string, endpoint int, ll zerolog.Logger) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/docker/%d/containers/%s/image_status?refresh=true", endpoint, containerID), nil, ll)
	if err != nil {
		return "", err
	}

	result := new(imageStatusResponse)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func (c *PortainerAPI) ServiceImageStatus(ctx context.Context, serviceID string, endpoint int, ll zerolog.Logger) (string, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/docker/%d/services/%s/image_status?refresh=true", endpoint, serviceID), nil, ll)
	if err != nil {
		return "", err
	}

	result := new(imageStatusResponse)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func (c *PortainerAPI) ContainersForStack(ctx context.Context, stack Stack, ll zerolog.Logger) ([]dockertypes.Container, error) {
	query := make(map[string]string)
	args := filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.compose.project=%s", stack.Name)))
	filtersStr, err := filters.ToJSON(args)
	if err != nil {
		return nil, err
	}

	query["filters"] = filtersStr
	response, err := c.get(ctx, fmt.Sprintf("api/endpoints/%d/docker/containers/json", stack.EndpointID), query, ll)
	if err != nil {
		return nil, err
	}

	var result []dockertypes.Container
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PortainerAPI) Containers(ctx context.Context, endpointID int, ll zerolog.Logger) ([]dockertypes.Container, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/endpoints/%d/docker/containers/json", endpointID), nil, ll)
	if err != nil {
		return nil, err
	}

	var result []dockertypes.Container
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PortainerAPI) Services(ctx context.Context, endpointID int, ll zerolog.Logger) ([]swarm.Service, error) {
	response, err := c.get(ctx, fmt.Sprintf("api/endpoints/%d/docker/services", endpointID), nil, ll)
	if err != nil {
		return nil, err
	}

	var result []swarm.Service
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PortainerAPI) ServicesForStack(ctx context.Context, stack Stack, ll zerolog.Logger) ([]swarm.Service, error) {
	query := make(map[string]string)
	args := filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.stack.namespace=%s", stack.Name)))
	filtersStr, err := filters.ToJSON(args)
	if err != nil {
		return nil, err
	}

	query["filters"] = filtersStr
	response, err := c.get(ctx, fmt.Sprintf("api/endpoints/%d/docker/services", stack.EndpointID), query, ll)
	if err != nil {
		return nil, err
	}

	var result []swarm.Service
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, err
	}

	return result, nil
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

func (c *PortainerAPI) updateGitStack(ctx context.Context, stack *Stack, ll zerolog.Logger) error {
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
		return errors.Wrap(err, "marshalling request body to json")
	}

	if _, err := c.put(ctx, fmt.Sprintf(
		"api/stacks/%d/git/redeploy?endpointId=%d",
		stack.ID,
		stack.EndpointID,
	), jsonRequest, ll); err != nil {
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

func (c *PortainerAPI) updateFileStack(ctx context.Context, stack *Stack, ll zerolog.Logger) error {
	fileContents, err := c.StackFileContent(ctx, int(stack.ID), ll)
	if err != nil {
		return errors.Wrap(err, "getting stack file contents")
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
		return errors.Wrap(err, "marshalling request body to json")
	}

	if _, err := c.put(ctx, fmt.Sprintf(
		"api/stacks/%d?endpointId=%d",
		stack.ID,
		stack.EndpointID,
	), jsonRequest, ll); err != nil {
		return errors.Wrap(err, "updating stack")
	}
	return nil
}

func (c *PortainerAPI) UpdateStack(ctx context.Context, stackID int, ll zerolog.Logger) error {
	stack, err := c.Stack(ctx, stackID, ll)
	if err != nil {
		return err
	}

	switch {
	case stack.GitConfig != nil:
		return c.updateGitStack(ctx, stack, ll)
	default:
		return c.updateFileStack(ctx, stack, ll)
	}
}

type forceUpdateServiceRequest struct {
	PullImage bool   `json:"pullImage"`
	ServiceID string `json:"serviceID"`
}

func (c *PortainerAPI) UpdateService(ctx context.Context, serviceID string, endpointID int, ll zerolog.Logger) error {
	request := forceUpdateServiceRequest{
		PullImage: true,
		ServiceID: serviceID,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "marshalling request body to json")
	}

	if _, err := c.put(ctx, fmt.Sprintf(
		"api/endpoints/%d/forceupdateservice",
		endpointID,
	), jsonRequest, ll); err != nil {
		return errors.Wrap(err, "updating service")
	}
	return nil
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
