package mock

import (
	"context"

	"github.com/GerindT/GoFlight/internal/domain"
)

type MockFlightFetcher struct {
	ReturnFlight *domain.FlightDetails
	ReturnErr    error
	Calls        int
}

func (m *MockFlightFetcher) Fetch(_ context.Context, _ string) (*domain.FlightDetails, error) {
	m.Calls++
	return m.ReturnFlight, m.ReturnErr
}
