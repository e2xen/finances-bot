package fixer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const (
	latestRatesURL = "https://api.apilayer.com/fixer/latest"
	baseParam      = "base"
	relativesParam = "symbols"
)

type apiKeyGetter interface {
	APIKey() string
}

type Client struct {
	apiKey string
}

type ratesResponse struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
}

func New(getter apiKeyGetter) *Client {
	return &Client{apiKey: getter.APIKey()}
}

func (c *Client) GetRates(ctx context.Context, baseRate string, relativeRates []string) (map[string]float64, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", latestRatesURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fixer client")
	}

	req.Header.Set("apikey", c.apiKey)
	q := req.URL.Query()
	q.Add(baseParam, baseRate)
	q.Add(relativesParam, strings.Join(relativeRates, ","))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fixer client")
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "fixer client")
	}
	log.Printf("new response from fixer: %s\n", string(body))

	rates := ratesResponse{}
	err = json.Unmarshal(body, &rates)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling response")
	}

	if !rates.Success {
		return nil, errors.New("error from fixer (success = false)")
	}

	return rates.Rates, nil
}
