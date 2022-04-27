// Package dimension provides an HTTP client for the Cantabular Dimension API
package dimension

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const service = "cantabular-dimension-api"

// Client is a Cantabular Dimension API client
type Client struct {
	hcCli   *health.Client
	baseURL *url.URL
}

type GetAreasInput struct {
	UserAuthToken    string
	ServiceAuthToken string
	DatasetID        string
	AreaTypeID       string
	Text             string
}

// NewClient creates a new instance of Client with a given Dimensions API URL
func NewClient(dimensionsAPIURL string) (*Client, error) {
	client := health.NewClient(service, dimensionsAPIURL)
	return NewWithHealthClient(client)
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client
func NewWithHealthClient(hcCli *health.Client) (*Client, error) {
	client := health.NewClientWithClienter(service, hcCli.URL, hcCli.Client)
	baseURL, err := url.Parse(client.URL)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing URL")
	}

	// The Parse method on `url.URL` uses a trailing slash to determine
	// how relative URLs are joined.
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path = baseURL.Path + "/"
	}

	return &Client{hcCli: client, baseURL: baseURL}, nil
}

// Checker calls recipe api health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

func (c *Client) createGetRequest(ctx context.Context, userAuthToken, serviceAuthToken, urlPath string, urlValues url.Values) (*http.Request, error) {
	areasURL, err := c.baseURL.Parse(urlPath)
	if err != nil {
		return &http.Request{}, dperrors.New(
			errors.Wrap(err, "failed to parse areas URL"),
			http.StatusInternalServerError,
			log.Data{},
		)
	}

	areasURL.RawQuery = urlValues.Encode()
	reqURL := areasURL.String()

	req, err := newRequest(ctx, http.MethodGet, reqURL, nil, userAuthToken, serviceAuthToken)
	if err != nil {
		return &http.Request{}, dperrors.New(
			errors.Wrap(err, "failed to create request"),
			http.StatusBadRequest,
			log.Data{},
		)
	}
	return req, nil
}

func checkGetResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		var errorResp ErrorResp
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return dperrors.New(
				fmt.Errorf("error response from Dimensions API (%d): %w", resp.StatusCode, errorResp),
				http.StatusInternalServerError,
				log.Data{},
			)
		}
	}

	if resp.StatusCode != http.StatusOK {
		// Best effort — an empty body is fine for the error message
		body, _ := io.ReadAll(resp.Body)
		return dperrors.New(
			errors.Errorf("error response from Dimensions API (%d): %s", resp.StatusCode, body),
			http.StatusInternalServerError,
			log.Data{},
		)
	}

	return nil
}

// GetAreaTypes retrieves the Cantabular area-types associated with a dataset
func (c *Client) GetAreaTypes(ctx context.Context, userAuthToken, serviceAuthToken, datasetID string) (GetAreaTypesResponse, error) {
	logData := log.Data{
		"method":     http.MethodGet,
		"dataset_id": datasetID,
	}

	urlPath := "area-types"
	urlValues := url.Values{"dataset": []string{datasetID}}

	req, err := c.createGetRequest(ctx, userAuthToken, serviceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting area types", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Dimensions API"),
			http.StatusInternalServerError,
			logData,
		)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}()

	if err := checkGetResponse(resp); err != nil {
		return GetAreaTypesResponse{}, err
	}

	var areaTypes GetAreaTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&areaTypes); err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize area types response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return areaTypes, nil
}

func (c *Client) GetAreas(ctx context.Context, input GetAreasInput) (GetAreasResponse, error) {
	logData := log.Data{
		"method":       http.MethodGet,
		"dataset_id":   input.DatasetID,
		"area_type_id": input.AreaTypeID,
		"text":         input.Text,
	}

	urlPath := "areas"
	urlValues := url.Values{"dataset": []string{input.DatasetID}, "area-type": []string{input.AreaTypeID}}
	if input.Text != "" {
		urlValues.Add("text", input.Text)
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetAreasResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting areas", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreasResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Dimensions API"),
			http.StatusInternalServerError,
			logData,
		)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}()

	if err := checkGetResponse(resp); err != nil {
		return GetAreasResponse{}, err
	}

	var areas GetAreasResponse
	if err := json.NewDecoder(resp.Body).Decode(&areas); err != nil {
		return GetAreasResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize areas response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return areas, nil
}

// newRequest creates a new http.Request with auth headers
func newRequest(ctx context.Context, method string, url string, body io.Reader, userAuthToken, serviceAuthToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	if err := headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, errors.Wrap(err, "failed to set auth token header")
	}

	if err := headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, errors.Wrap(err, "failed to set service token header")
	}

	return req, nil
}
