package cantabular

import (
	"context"
	"io/ioutil"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/ONSdigital/log.go/v2/log"
	dperrors "github.com/ONSdigital/dp-api-clients-go/errors"
)

// Variable represents a 'codebook' object returned from Cantabular
type Codebook []Variable

// GetCodebook gets a Codebook from cantabular.
func (c *Client) GetCodebook(ctx context.Context, req GetCodebookRequest) (*GetCodebookResponse, error){
	var vars string
	for _, v := range req.Variables{
		vars += "&v=" + v
	}

	url := fmt.Sprintf("%s/v8/codebook/%s?cats=%v%s", c.host, req.DatasetName, req.Categories, vars)

	res, err := c.httpGet(ctx, url)
	if err != nil{
		return nil, dperrors.New(
			fmt.Errorf("failed to get response from Cantabular API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK{
		return nil, c.errorResponse(res)
	}

	var resp GetCodebookResponse

	b, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			res.StatusCode,
			log.Data{
				"response_body": string(b),
			},
		)
	}

	if len(b) == 0{
		b = []byte("[response body empty]")
	}

	if err := json.Unmarshal(b, &resp); err != nil{
		return nil, dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{
				"response_body": string(b),
			},
		)
	}

	return &resp, nil
}