package cantabular

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

// QueryStaticDataset is the graphQL query to obtain static dataset counts (variables with categories and counts)
const QueryStaticDataset = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				count
				variable { name label }
				categories { code label }
			}
			values
			error
		}
	}
}`

// QueryDimensionOptions is the graphQL query to obtain static dataset dimension options (variables with categories)
const QueryDimensionOptions = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				variable { name label }
				categories { code label }
			}
			values
			error
		}
	}
}`

// QueryDimensions is the graphQL query to obtain dimensions (variables without categories)
const QueryDimensions = `
query($dataset: String!) {
	dataset(name: $dataset) {
		variables {
			edges {
				node {
					name
					mapFrom {
						edges {
							node {
								filterOnly
								label
								name
							}
						}
					}
					label
					categories {
						totalCount
					}
				}
			}
		}
	}
}`

// QueryDimensionsByName is the graphQL query to obtain dimensions by name (subset of variables, without categories)
const QueryDimensionsByName = `
query($dataset: String!, $variables: [String!]!) {
	dataset(name: $dataset) {
		variables(names: $variables) {
			edges {
				node {
					name
					mapFrom {
						edges {
							node {
								filterOnly
								label
								name
							}
						}
					}
					label
					categories {
						totalCount
					}
				}
			}
		}
	}
}`

// QueryGeographyDimensions is the graphQL query to obtain geography dimensions (subset of variables, without categories)
const QueryGeographyDimensions = `
query($dataset: String!) {
	dataset(name: $dataset) {
		ruleBase {
			name
			isSourceOf {
				edges {
					node {
						name
						mapFrom {
							edges {
								node {
									filterOnly
									label
									name
								}
							}
						}
						label
						categories{
							totalCount
						}
					}
				}
			}
		}
	}
}`

const QueryDimensionsSearch = `
query($dataset: String!, $text: String!) {
	dataset(name: $dataset) {
		variables {
			search(text: $text) {
				edges {
					node {
						name
						label
						mapFrom {
							totalCount
							edges {
								node {
									name
									label
								}
							}
						}
					}
				}
			}
		}
	}
}`

// QueryDimensionsByName is the graphQL query to obtain dimensions by name (subset of variables, without categories)
const QueryAreasByArea = `
query($dataset: String!, $text: String!) {
	dataset(name: $dataset) {
		ruleBase 
		{
		  isSourceOf {
			categorySearch(text: $text){
		   		edges {
			  		node { 
						code
						label
						variable {
				  			mapFrom{
								edges{
					  				node{
										name
										label
										 }
									}
				 				 }
							 name 
						}
					  } 
	  				}
				}
	  		}
	  	}
	  }
  }
`

// QueryData holds all the possible required variables to encode any of the graphql queries defined in this file.
type QueryData struct {
	Dataset   string
	Text      string
	Variables []string
	Filters   []Filter
}

// Filter holds the fields for the Cantabular GraphQL 'Filter' object used for specifying categories
// returned in tables
type Filter struct {
	Codes    []string `json:"codes"`
	Variable string   `json:"variable"`
}

// Encode the provided graphQL query with the data in QueryData
// returns a byte buffer with the encoded query, along with any encoding error that might happen
func (data *QueryData) Encode(query string) (bytes.Buffer, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	vars := map[string]interface{}{
		"dataset":   data.Dataset,
		"variables": data.Variables,
		"text":      data.Text,
	}
	if len(data.Filters) > 0 {
		vars["filters"] = data.Filters
	}
	if err := enc.Encode(map[string]interface{}{
		"query":     query,
		"variables": vars,
	}); err != nil {
		return b, fmt.Errorf("failed to encode GraphQL query: %w", err)
	}

	return b, nil
}

// queryUnmarshal uses postQuery to perform a graphQL query and then un-marshals the response body to the provided value pointer v
// This method handles the response body closing.
func (c *Client) queryUnmarshal(ctx context.Context, graphQLQuery string, data QueryData, v interface{}) error {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	res, err := c.postQuery(ctx, graphQLQuery, data)
	defer closeResponseBody(ctx, res)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			res.StatusCode,
			log.Data{
				"url": url,
			},
		)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{
				"url":           url,
				"response_body": string(b),
			},
		)
	}

	return nil
}

// postQuery performs a query against the Cantabular Extended API
// using the /graphql endpoint and the http client directly
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) postQuery(ctx context.Context, graphQLQuery string, data QueryData) (*http.Response, error) {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url":        url,
		"query_data": data,
	}

	b, err := data.Encode(graphQLQuery)
	if err != nil {
		return nil, dperrors.New(err, http.StatusInternalServerError, logData)
	}

	// Do a POST call to graphQL endpoint
	res, err := c.httpPost(ctx, url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	// Check status code and return error
	if res.StatusCode != http.StatusOK {
		closeResponseBody(ctx, res)
		return nil, c.errorResponse(url, res)
	}

	return res, nil
}