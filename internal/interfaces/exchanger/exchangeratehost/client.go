package exchangeratehost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ExchangeRateHostResponse struct {
	Rates map[string]float64 `json:"rates"`
}

type ExchangeRateHostClient struct {
	baseURL    string
	httpClient *http.Client
}

func New() *ExchangeRateHostClient {
	return &ExchangeRateHostClient{
		baseURL: "https://api.frankfurter.app",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ExchangeRateHostClient) GetRate(ctx context.Context, from, to string) (float64, error) {
	url := fmt.Sprintf("%s/latest?from=%s&to=%s", c.baseURL, from, to)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var apiResponse ExchangeRateHostResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return 0, err
	}

	rate, exists := apiResponse.Rates[to]
	if !exists {
		return 0, fmt.Errorf("rate not found")
	}

	return rate, nil
}

func (c *ExchangeRateHostClient) GetName() string {
	return "Frankfurter"
}

func (c *ExchangeRateHostClient) IsAvailable() bool {
	return true
}
