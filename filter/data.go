package filter

import (
	"time"
)

const (
	EventFilterOutputQueryStart  = "FilterOutputQueryStart"
	EventFilterOutputQueryEnd    = "FilterOutputQueryEnd"
	EventFilterOutputCSVGenStart = "FilterOutputCSVGenStart"
	EventFilterOutputCSVGenEnd   = "FilterOutputCSVGenEnd"
)

// Dimensions represents a dimensions response from the filter api
type Dimensions struct {
	Items      []Dimension `json:"items"`
	Count      int         `json:"count"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
}

// Dimension represents a dimension response from the filter api
type Dimension struct {
	Name                  string   `json:"name"`
	ID                    string   `json:"id,omitempty"`
	Label                 string   `json:"label,omitempty"`
	URI                   string   `json:"dimension_url"`
	IsAreaType            *bool    `json:"is_area_type,omitempty"`
	Options               []string `json:"options,omitempty"`
	DefaultCategorisation string   `json:"default_categorisation"`
	FilterByParent        string   `json:"filter_by_parent,omitempty"`
}

// DimensionOption represents a dimension option from the filter api
type DimensionOption struct {
	DimensionOptionsURL string `json:"dimension_option_url"`
	Option              string `json:"option"`
}

// DimensionOptions represents a list of dimension options from the filter api
type DimensionOptions struct {
	Items      []DimensionOption `json:"items"`
	Count      int               `json:"count"`
	Offset     int               `json:"offset"`
	Limit      int               `json:"limit"`
	TotalCount int               `json:"total_count"`
}

// createBlueprint represents the fields required to create a filter blueprint
type createBlueprint struct {
	Dataset    Dataset          `json:"dataset"`
	Dimensions []ModelDimension `json:"dimensions"`
	FilterID   string           `json:"filter_id"`
}

type createFlexBlueprintRequest struct {
	Dataset        Dataset          `json:"dataset"`
	Dimensions     []ModelDimension `json:"dimensions"`
	PopulationType string           `json:"population_type"`
}

type createFlexBlueprintResponse struct {
	FilterID string `json:"filter_id"`
}

// createFlexDimensionRequest represents the fields required to add a dimension to a flex filter
type createFlexDimensionRequest struct {
	Name       string   `json:"name"`
	IsAreaType bool     `json:"is_area_type"`
	Options    []string `json:"options"`
}

// Dataset represents the dataset fields required to create a filter blueprint
type Dataset struct {
	DatasetID string `json:"id"`
	Edition   string `json:"edition"`
	Version   int    `json:"version"`
}

// Model represents a model returned from the filter api
type Model struct {
	FilterID       string              `json:"filter_id"`
	InstanceID     string              `json:"instance_id"`
	Links          Links               `json:"links"`
	DatasetID      string              `json:"dataset_id"`
	Dataset        Dataset             `json:"dataset,omitempty"`
	Edition        string              `json:"edition"`
	Version        string              `json:"version"`
	State          string              `json:"state"`
	Dimensions     []ModelDimension    `json:"dimensions,omitempty"`
	Downloads      map[string]Download `json:"downloads,omitempty"`
	Events         []Event             `json:"events,omitempty"`
	IsPublished    bool                `json:"published"`
	PopulationType string              `json:"population_type,omitempty"`
}

// Links represents a links object on the filter api response
type Links struct {
	Version         Link `json:"version,omitempty"`
	FilterOutputs   Link `json:"filter_output,omitempty"`
	FilterBlueprint Link `json:"filter_blueprint,omitempty"`
}

// Link represents a single link within a links object
type Link struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
}

// ModelDimension represents a dimension to be filtered upon
type ModelDimension struct {
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	URI            string   `json:"dimension_url,omitempty"`
	IsAreaType     *bool    `json:"is_area_type,omitempty"`
	Options        []string `json:"options"`
	Values         []string `json:"values"`
	FilterByParent string   `json:"filter_by_parent,omitempty"`
}

// Download represents a download within a filter from api response
type Download struct {
	URL     string `json:"href"`
	Size    string `json:"size"`
	Public  string `json:"public,omitempty"`
	Private string `json:"private,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
}

// Event represents an event from a filter api response
type Event struct {
	Time time.Time `json:"time"`
	Type string    `json:"type"`
}

// Preview represents a preview document returned from the filter api
type Preview struct {
	Headers         []string   `json:"headers"`
	NumberOfRows    int        `json:"number_of_rows"`
	NumberOfColumns int        `json:"number_of_columns"`
	Rows            [][]string `json:"rows"`
}

type SubmitFilterRequest struct {
	FilterID       string             `json:"filter_id"`
	Dimensions     []DimensionOptions `json:"dimension_options,omitempty"`
	PopulationType string             `json:"population_type"`
}

type SubmitFilterResponse struct {
	InstanceID     string      `json:"instance_id"`
	FilterOutputID string      `json:"filter_output_id"`
	Dataset        Dataset     `json:"dataset"`
	Links          FilterLinks `json:"links"`
	PopulationType string      `json:"population_type"`
}
