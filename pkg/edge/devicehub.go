package edge

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrUnauthorized is returned when the DeviceHub API responds with HTTP 401.
var ErrUnauthorized = errors.New("unauthorized: invalid or expired API Token")

// DeviceHubClient queries the Litmus Edge DeviceHub GraphQL API for topic discovery.
type DeviceHubClient interface {
	SearchTopics(ctx context.Context, query string) ([]string, error)
}

type deviceHubClient struct {
	httpClient *http.Client
	endpoint   string
	authHeader string
}

// NewDeviceHubClient creates a client that queries the DeviceHub GraphQL API.
func NewDeviceHubClient(hostname, apiToken string) DeviceHubClient {
	return &deviceHubClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Litmus Edge uses self-signed certs on local network
			},
		},
		endpoint:   fmt.Sprintf("https://%s/devicehub/v2", hostname),
		authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte(apiToken+":")),
	}
}

const listRegistersQuery = `query ListRegistersFromAllDevices($input: ListRegistersFromAllDevicesRequest!) {
  ListRegistersFromAllDevices(input: $input) {
    Registers {
      Topics {
        Format
        Topic
      }
    }
  }
}`

type graphQLRequest struct {
	OperationName string      `json:"operationName"`
	Query         string      `json:"query"`
	Variables     interface{} `json:"variables"`
}

type listRegistersVariables struct {
	Input listRegistersInput `json:"input"`
}

type listRegistersInput struct {
	TagPattern             string `json:"TagPattern"`
	TagPatternSearchOption string `json:"TagPatternSearchOption"`
	Limit                  int    `json:"Limit"`
}

type graphQLResponse struct {
	Data   graphQLData    `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLData struct {
	ListRegistersFromAllDevices listRegistersResult `json:"ListRegistersFromAllDevices"`
}

type listRegistersResult struct {
	Registers []register `json:"Registers"`
}

type register struct {
	Topics []topicInfo `json:"Topics"`
}

type topicInfo struct {
	Format string `json:"Format"`
	Topic  string `json:"Topic"`
}

type graphQLError struct {
	Message string `json:"message"`
}

func (c *deviceHubClient) SearchTopics(ctx context.Context, query string) ([]string, error) {
	body := graphQLRequest{
		OperationName: "ListRegistersFromAllDevices",
		Query:         listRegistersQuery,
		Variables: listRegistersVariables{
			Input: listRegistersInput{
				TagPattern:             query,
				TagPatternSearchOption: "CONTAINS",
				Limit:                  15,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling DeviceHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading DeviceHub response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeviceHub API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("parsing DeviceHub response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("DeviceHub GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	seen := make(map[string]struct{})
	var topics []string
	for _, reg := range gqlResp.Data.ListRegistersFromAllDevices.Registers {
		for _, t := range reg.Topics {
			if t.Format != "Raw" {
				continue
			}
			if _, ok := seen[t.Topic]; ok {
				continue
			}
			seen[t.Topic] = struct{}{}
			topics = append(topics, t.Topic)
		}
	}

	if topics == nil {
		topics = []string{}
	}

	return topics, nil
}
