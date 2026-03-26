package external

import (
	"context"
	"time"

	"github.com/GerindT/GoFlight/internal/domain"
)

type FlightFetcher interface {
	Fetch(ctx context.Context, flightNumber string) (*domain.FlightDetails, error)
}

type WeatherFetcher interface {
	Fetch(ctx context.Context, flightNumber string) (*domain.WeatherDetails, error)
}

type CacheManager interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Ping(ctx context.Context) error
}
