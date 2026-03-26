package domain

import "time"

type AggregatedResponse struct {
	Flight      *FlightDetails  `json:"flight"`
	Weather     *WeatherDetails `json:"weather"`
	Cached      bool            `json:"cached"`
	GeneratedAt time.Time       `json:"generated_at"`
}

type FlightDetails struct {
	FlightNumber   string    `json:"flight_number"`
	Airline        string    `json:"airline"`
	Departure      string    `json:"departure"`
	Destination    string    `json:"destination"`
	ScheduledTime  time.Time `json:"scheduled_time"`
	EstimatedTime  time.Time `json:"estimated_time"`
	Status         string    `json:"status"`
	Terminal       string    `json:"terminal,omitempty"`
	Gate           string    `json:"gate,omitempty"`
	DelayInMinutes int       `json:"delay_in_minutes"`
}

type WeatherDetails struct {
	Location        string  `json:"location"`
	TemperatureC    float64 `json:"temperature_c"`
	FeelsLikeC      float64 `json:"feels_like_c"`
	Condition       string  `json:"condition"`
	WindSpeedKph    float64 `json:"wind_speed_kph"`
	HumidityPercent int     `json:"humidity_percent"`
}
