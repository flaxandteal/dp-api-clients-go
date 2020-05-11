package zebedee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

const service = "zebedee"

// Client represents a zebedee client
type Client struct {
	cli dphttp.Clienter
	url string
}

// ErrInvalidZebedeeResponse is returned when zebedee does not respond
// with a valid status
type ErrInvalidZebedeeResponse struct {
	actualCode int
	uri        string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidZebedeeResponse) Error() string {
	return fmt.Sprintf("invalid response from zebedee - should be 2.x.x or 3.x.x, got: %d, path: %s",
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidZebedeeResponse{}

var (
	errCastingCollectionID = errors.New("error casting collection ID cookie to string")
	errCastingLocalCode    = errors.New("error casting locale code to string")
)

// New creates a new Zebedee Client, set ZEBEDEE_REQUEST_TIMEOUT_SECOND
// environment variable to modify default client timeout as zebedee can often be slow
// to respond
func New(zebedeeURL string) *Client {
	timeout, err := strconv.Atoi(os.Getenv("ZEBEDEE_REQUEST_TIMEOUT_SECONDS"))
	if timeout == 0 || err != nil {
		timeout = 5
	}
	hcClient := healthcheck.NewClient(service, zebedeeURL)
	hcClient.Client.SetTimeout(time.Duration(timeout) * time.Second)

	return &Client{
		cli: hcClient.Client,
		url: zebedeeURL,
	}
}

// Checker calls zebedee health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	hcClient := healthcheck.Client{
		Client: c.cli,
		URL:    c.url,
		Name:   service,
	}

	return hcClient.Checker(ctx, check)
}

// Get returns a response for the requested uri in zebedee
func (c *Client) Get(ctx context.Context, userAccessToken, path string) ([]byte, error) {
	b, _, err := c.get(ctx, userAccessToken, path)
	return b, err
}

// GetWithHeaders returns a response for the requested uri in zebedee, providing the headers too
func (c *Client) GetWithHeaders(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	return c.get(ctx, userAccessToken, path)
}

// GetDatasetLandingPage returns a DatasetLandingPage populated with data from a zebedee response. If an error
// is returned there is a chance that a partly completed DatasetLandingPage is returned
func (c *Client) GetDatasetLandingPage(ctx context.Context, userAccessToken, path string) (DatasetLandingPage, error) {
	reqURL := c.createRequestURL(ctx, "/data", "uri="+path)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return DatasetLandingPage{}, err
	}

	var dlp DatasetLandingPage
	if err = json.Unmarshal(b, &dlp); err != nil {
		return dlp, err
	}

	related := [][]Related{
		dlp.RelatedDatasets,
		dlp.RelatedDocuments,
		dlp.RelatedMethodology,
		dlp.RelatedMethodologyArticle,
	}

	//Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	sem := make(chan int, 10)

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e Related, element []Related) {
				defer func() {
					<-sem
					wg.Done()
				}()
				t, _ := c.GetPageTitle(ctx, userAccessToken, e.URI)
				element[i].Title = t.Title
			}(i, e, element)
		}
	}
	wg.Wait()

	return dlp, nil
}

func (c *Client) get(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", c.url+path, nil)
	if err != nil {
		return nil, nil, err
	}

	common.AddFlorenceHeader(req, userAccessToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, nil, ErrInvalidZebedeeResponse{resp.StatusCode, req.URL.Path}
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, resp.Header, err
}

// GetBreadcrumb returns a Breadcrumb
func (c *Client) GetBreadcrumb(ctx context.Context, userAccessToken, uri string) ([]Breadcrumb, error) {
	b, _, err := c.get(ctx, userAccessToken, "/parents?uri="+uri)
	if err != nil {
		return nil, err
	}

	var parentsJSON []Breadcrumb
	if err = json.Unmarshal(b, &parentsJSON); err != nil {
		return nil, err
	}

	return parentsJSON, nil
}

// GetDataset returns details about a dataset from zebedee
func (c *Client) GetDataset(ctx context.Context, userAccessToken, uri string) (Dataset, error) {
	reqURL := c.createRequestURL(ctx, "/data", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	if err != nil {
		return Dataset{}, err
	}

	var d Dataset
	if err = json.Unmarshal(b, &d); err != nil {
		return d, err
	}

	downloads := make([]Download, 0)

	for _, v := range d.Downloads {
		fs, err := c.GetFileSize(ctx, userAccessToken, uri+"/"+v.File)
		if err != nil {
			return d, err
		}

		downloads = append(downloads, Download{
			File: v.File,
			Size: strconv.Itoa(fs.Size),
		})
	}

	d.Downloads = downloads

	supplementaryFiles := make([]SupplementaryFile, 0)
	for _, v := range d.SupplementaryFiles {
		fs, err := c.GetFileSize(ctx, userAccessToken, uri+"/"+v.File)
		if err != nil {
			return d, err
		}

		supplementaryFiles = append(supplementaryFiles, SupplementaryFile{
			File:  v.File,
			Title: v.Title,
			Size:  strconv.Itoa(fs.Size),
		})
	}

	d.SupplementaryFiles = supplementaryFiles

	return d, nil
}

// GetFileSize retrieves a given filesize from zebedee
func (c *Client) GetFileSize(ctx context.Context, userAccessToken, uri string) (FileSize, error) {
	reqURL := c.createRequestURL(ctx, "/filesize", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return FileSize{}, err
	}

	var fs FileSize
	if err = json.Unmarshal(b, &fs); err != nil {
		return fs, err
	}

	return fs, nil
}

// GetPageTitle retrieves a page title from zebedee
func (c *Client) GetPageTitle(ctx context.Context, userAccessToken, uri string) (PageTitle, error) {
	reqURL := c.createRequestURL(ctx, "/data", "uri="+uri+"&title")
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return PageTitle{}, err
	}

	var pt PageTitle
	if err = json.Unmarshal(b, &pt); err != nil {
		return pt, err
	}

	return pt, nil
}

func (c *Client) GetTimeseriesMainFigure(ctx context.Context, userAccessToken, uri string) (TimeseriesMainFigure, error) {
	reqURL := c.createRequestURL(ctx, "/data", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	if err != nil {
		return TimeseriesMainFigure{}, err
	}

	var ts TimeseriesMainFigure
	if err = json.Unmarshal(b, &ts); err != nil {
		return ts, err
	}

	return ts, nil
}

func (c *Client) createRequestURL(ctx context.Context, path, query string) string {
	// Check if collection ID is set in context
	if ctx.Value(common.CollectionIDHeaderKey) != nil {
		collectionID, ok := ctx.Value(common.CollectionIDHeaderKey).(string)
		if !ok {
			log.Event(ctx, "unable to find collection ID header key", log.ERROR, log.Error(errCastingCollectionID))
		}
		path += "/" + collectionID
	}

	path += "?" + url.PathEscape(query)

	// Check if locale code is set in context and add lang query param to url
	if ctx.Value(common.LocaleHeaderKey) != nil {
		localeCode, ok := ctx.Value(common.LocaleHeaderKey).(string)
		if !ok {
			log.Event(ctx, "unable to find local header key", log.ERROR, log.Error(errCastingLocalCode))
		}
		path += "&lang=" + localeCode
	}

	return path
}
