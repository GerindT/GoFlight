package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/GerindT/GoFlight/internal/apierrors"
	"github.com/GerindT/GoFlight/internal/domain"
)

type AviationStackClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewAviationStackClient(baseURL, apiKey string, client *http.Client) *AviationStackClient {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &AviationStackClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  client,
	}
}

type aviationStackResponse struct {
	Data []struct {
		Flight struct {
			IATA *string `json:"iata"`
		} `json:"flight"`
		Airline struct {
			Name *string `json:"name"`
		} `json:"airline"`
		Departure struct {
			Airport   *string `json:"airport"`
			Scheduled *string `json:"scheduled"`
			Terminal  *string `json:"terminal"`
			Gate      *string `json:"gate"`
		} `json:"departure"`
		Arrival struct {
			Airport   *string `json:"airport"`
			Estimated *string `json:"estimated"`
			Delay     *int    `json:"delay"`
		} `json:"arrival"`
		FlightStatus *string `json:"flight_status"`
	} `json:"data"`
}

func (a *AviationStackClient) Fetch(ctx context.Context, flightNumber string) (*domain.FlightDetails, error) {
	u, err := url.Parse(a.baseURL + "/flights")
	if err != nil {
		return nil, fmt.Errorf("aviationstack: %w", apierrors.ErrUpstreamBadData)
	}

	q := u.Query()
	q.Set("access_key", a.apiKey)
	q.Set("flight_iata", flightNumber)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("aviationstack: %w", apierrors.ErrUpstreamBadData)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aviationstack: %w", apierrors.ErrUpstreamTimeout)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("aviationstack: %w", apierrors.ErrUpstreamBadData)
	}

	var decoded aviationStackResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("aviationstack: %w", apierrors.ErrUpstreamBadData)
	}
	if len(decoded.Data) == 0 {
		return nil, apierrors.ErrFlightNotFound
	}

	entry := decoded.Data[0]
	scheduled := parseRFC3339Ptr(entry.Departure.Scheduled)
	estimated := parseRFC3339Ptr(entry.Arrival.Estimated)

	return &domain.FlightDetails{
		FlightNumber:   valueOrEmpty(entry.Flight.IATA),
		Airline:        valueOrEmpty(entry.Airline.Name),
		Departure:      valueOrEmpty(entry.Departure.Airport),
		Destination:    valueOrEmpty(entry.Arrival.Airport),
		ScheduledTime:  scheduled,
		EstimatedTime:  estimated,
		Status:         valueOrEmpty(entry.FlightStatus),
		Terminal:       valueOrEmpty(entry.Departure.Terminal),
		Gate:           valueOrEmpty(entry.Departure.Gate),
		DelayInMinutes: valueOrZero(entry.Arrival.Delay),
	}, nil
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func parseRFC3339Ptr(v *string) time.Time {
	if v == nil || *v == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, *v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
