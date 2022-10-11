package fixer

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	latestRatesUrl = "https://api.apilayer.com/fixer/latest"
	baseParam      = "base"
	relativesParam = "symbols"
)

type apiKeyGetter interface {
	ApiKey() string
}

type Client struct {
	apiKey string
}

type ratesResponse struct {
	Base string `json:"base"`
	//Date      time.Time          `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
}

func New(getter apiKeyGetter) *Client {
	return &Client{apiKey: getter.ApiKey()}
}

func (c *Client) GetRates(baseRate string, relativeRates []string) (map[string]float64, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", latestRatesUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.apiKey)
	q := req.URL.Query()
	q.Add(baseParam, baseRate)
	q.Add(relativesParam, strings.Join(relativeRates, ","))
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Println(fmt.Sprintf("new response from fixer: %s", string(body)))

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
