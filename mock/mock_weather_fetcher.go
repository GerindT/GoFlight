package mock

import (
	"context"

	"github.com/GerindT/GoFlight/internal/domain"
)

type MockWeatherFetcher struct {
	ReturnWeather *domain.WeatherDetails
	ReturnErr     error
	Calls         int
}

func (m *MockWeatherFetcher) Fetch(_ context.Context, _ string) (*domain.WeatherDetails, error) {
	m.Calls++
	return m.ReturnWeather, m.ReturnErr
}
