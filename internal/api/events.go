package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultEventsPageSize = 200

// GetFilteredUsageEvents fetches paginated, filterable usage events.
func (c *Client) GetFilteredUsageEvents(req EventsRequest) (*EventsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/dashboard/get-filtered-usage-events",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	respBody, err := c.CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	var eventsResp EventsResponse
	if err := json.NewDecoder(respBody).Decode(&eventsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &eventsResp, nil
}

// GetAllFilteredUsageEvents fetches every page of usage events for req.
func (c *Client) GetAllFilteredUsageEvents(req EventsRequest) (*EventsResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = defaultEventsPageSize
	}
	req.Page = 1

	var all []UsageEvent
	var total int
	for {
		resp, err := c.GetFilteredUsageEvents(req)
		if err != nil {
			return nil, err
		}
		total = resp.TotalUsageEventsCount
		all = append(all, resp.UsageEventsDisplay...)
		if len(all) >= total || len(resp.UsageEventsDisplay) == 0 {
			break
		}
		req.Page++
		time.Sleep(200 * time.Millisecond)
	}

	return &EventsResponse{
		TotalUsageEventsCount: total,
		UsageEventsDisplay:    all,
	}, nil
}
