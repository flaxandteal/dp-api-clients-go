package config

type NLP struct {
	BerlinAPIEndpoint   string `envconfig:"NLP_BERLIN_API_ENDPOINT"`
	BerlinAPIURL        string `envconfig:"NLP_BERLIN_API_URL"`
	CategoryAPIEndpoint string `envconfig:"NLP_CATEGORY_API_ENDPOINT"`
	CategoryAPIURL      string `envconfig:"NLP_CATEGORY_API_URL"`
	ScrubberAPIEndpoint string `envconfig:"NLP_SCRUBBER_API_ENDPOINT"`
	ScrubberAPIURL      string `envconfig:"NLP_SCRUBBER_API_URL"`
}
