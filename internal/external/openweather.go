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

type OpenWeatherClient struct {
	baseURL     string
	apiKey      string
	defaultCity string
	client      *http.Client
}

func NewOpenWeatherClient(baseURL, apiKey, defaultCity string, client *http.Client) *OpenWeatherClient {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &OpenWeatherClient{
		baseURL:     baseURL,
		apiKey:      apiKey,
		defaultCity: defaultCity,
		client:      client,
	}
}

type openWeatherResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func (o *OpenWeatherClient) Fetch(ctx context.Context, _ string) (*domain.WeatherDetails, error) {
	u, err := url.Parse(o.baseURL + "/weather")
	if err != nil {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	q := u.Query()
	q.Set("appid", o.apiKey)
	q.Set("q", o.defaultCity)
	q.Set("units", "metric")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamTimeout)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	var decoded openWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}
	if decoded.Name == "" {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	condition := "unknown"
	if len(decoded.Weather) > 0 {
		condition = decoded.Weather[0].Description
	}

	return &domain.WeatherDetails{
		Location:        decoded.Name,
		TemperatureC:    decoded.Main.Temp,
		FeelsLikeC:      decoded.Main.FeelsLike,
		Condition:       condition,
		WindSpeedKph:    decoded.Wind.Speed * 3.6,
		HumidityPercent: decoded.Main.Humidity,
	}, nil
}
