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
			IATA string `json:"iata"`
		} `json:"flight"`
		Airline struct {
			Name string `json:"name"`
		} `json:"airline"`
		Departure struct {
			Airport   string `json:"airport"`
			Scheduled string `json:"scheduled"`
			Terminal  string `json:"terminal"`
			Gate      string `json:"gate"`
		} `json:"departure"`
		Arrival struct {
			Airport   string `json:"airport"`
			Estimated string `json:"estimated"`
			Delay     int    `json:"delay"`
		} `json:"arrival"`
		FlightStatus string `json:"flight_status"`
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
	scheduled, _ := time.Parse(time.RFC3339, entry.Departure.Scheduled)
	estimated, _ := time.Parse(time.RFC3339, entry.Arrival.Estimated)

	return &domain.FlightDetails{
		FlightNumber:   entry.Flight.IATA,
		Airline:        entry.Airline.Name,
		Departure:      entry.Departure.Airport,
		Destination:    entry.Arrival.Airport,
		ScheduledTime:  scheduled,
		EstimatedTime:  estimated,
		Status:         entry.FlightStatus,
		Terminal:       entry.Departure.Terminal,
		Gate:           entry.Departure.Gate,
		DelayInMinutes: entry.Arrival.Delay,
	}, nil
}
