package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetUsageSummary fetches billing cycle info, plan limits, and current usage totals.
func (c *Client) GetUsageSummary() (*UsageSummary, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/usage-summary", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	body, err := c.CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	var summary UsageSummary
	if err := json.NewDecoder(body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &summary, nil
}
