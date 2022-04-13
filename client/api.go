package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type APIClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewAPIClient(baseURL string) (*APIClient, error) {
	b, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	return &APIClient{
		httpClient: &http.Client{},
		baseURL:    b,
	}, nil
}

func (c *APIClient) Prices(ctx context.Context) (map[string]float64, error) {
	u, err := c.baseURL.Parse("/asset/live")
	if err != nil {
		return nil, fmt.Errorf("parse url ref: %w", err)
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	var body struct {
		Data []struct {
			Denom string
			Price float64 `json:"priceOracle"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	prices := map[string]float64{}
	for _, coin := range body.Data {
		prices[coin.Denom] = coin.Price
	}
	return prices, nil
}
