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
	baseURL         string
	defaultLocation string
	lat             float64
	lon             float64
	client          *http.Client
}

func NewOpenWeatherClient(baseURL, defaultLocation string, lat, lon float64, client *http.Client) *OpenWeatherClient {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &OpenWeatherClient{
		baseURL:         baseURL,
		defaultLocation: defaultLocation,
		lat:             lat,
		lon:             lon,
		client:          client,
	}
}

type openWeatherResponse struct {
	Current struct {
		Temp             float64 `json:"temperature_2m"`
		FeelsLike        float64 `json:"apparent_temperature"`
		Humidity         int     `json:"relative_humidity_2m"`
		WindSpeed        float64 `json:"wind_speed_10m"`
		WeatherCode      int     `json:"weather_code"`
		IsDay            int     `json:"is_day"`
		ObservationTime  string  `json:"time"`
	} `json:"current"`
}

func (o *OpenWeatherClient) Fetch(ctx context.Context, _ string) (*domain.WeatherDetails, error) {
	u, err := url.Parse(o.baseURL + "/forecast")
	if err != nil {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.6f", o.lat))
	q.Set("longitude", fmt.Sprintf("%.6f", o.lon))
	q.Set("current", "temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m,is_day")
	q.Set("timezone", "auto")
	q.Set("wind_speed_unit", "kmh")
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
	if decoded.Current.ObservationTime == "" {
		return nil, fmt.Errorf("openweather: %w", apierrors.ErrUpstreamBadData)
	}

	return &domain.WeatherDetails{
		Location:        o.defaultLocation,
		TemperatureC:    decoded.Current.Temp,
		FeelsLikeC:      decoded.Current.FeelsLike,
		Condition:       weatherCodeToText(decoded.Current.WeatherCode, decoded.Current.IsDay == 1),
		WindSpeedKph:    decoded.Current.WindSpeed,
		HumidityPercent: decoded.Current.Humidity,
	}, nil
}

func weatherCodeToText(code int, day bool) string {
	switch code {
	case 0:
		if day {
			return "clear sky"
		}
		return "clear night"
	case 1:
		if day {
			return "mainly clear"
		}
		return "mainly clear night"
	case 2:
		return "partly cloudy"
	case 3:
		return "overcast"
	case 45, 48:
		return "fog"
	case 51, 53, 55:
		return "drizzle"
	case 56, 57:
		return "freezing drizzle"
	case 61, 63, 65:
		return "rain"
	case 66, 67:
		return "freezing rain"
	case 71, 73, 75, 77:
		return "snow"
	case 80, 81, 82:
		return "rain showers"
	case 85, 86:
		return "snow showers"
	case 95:
		return "thunderstorm"
	case 96, 99:
		return "thunderstorm with hail"
	default:
		return "unknown"
	}
}
