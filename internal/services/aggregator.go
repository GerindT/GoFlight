package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GerindT/GoFlight/internal/domain"
	"github.com/GerindT/GoFlight/internal/external"
	"github.com/sony/gobreaker"
)

type Aggregator struct {
	flightFetcher  external.FlightFetcher
	weatherFetcher external.WeatherFetcher
	cache          external.CacheManager
	cacheTTL       time.Duration
	flightCB       *gobreaker.CircuitBreaker
	weatherCB      *gobreaker.CircuitBreaker
}

func NewAggregator(f external.FlightFetcher, w external.WeatherFetcher, c external.CacheManager, ttl time.Duration) *Aggregator {
	return &Aggregator{
		flightFetcher:  f,
		weatherFetcher: w,
		cache:          c,
		cacheTTL:       ttl,
		flightCB: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "aviationstack",
			MaxRequests: 3,
			Timeout:     30 * time.Second,
		}),
		weatherCB: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "openweather",
			MaxRequests: 3,
			Timeout:     30 * time.Second,
		}),
	}
}

func (a *Aggregator) cacheKey(flightNumber string) string {
	return fmt.Sprintf("flight_dash:%s", flightNumber)
}

func (a *Aggregator) GetFlightDash(ctx context.Context, flightNumber string) (*domain.AggregatedResponse, error) {
	key := a.cacheKey(flightNumber)
	if cached, err := a.cache.Get(ctx, key); err == nil {
		var out domain.AggregatedResponse
		if json.Unmarshal([]byte(cached), &out) == nil {
			out.Cached = true
			return &out, nil
		}
	}

	type flightResult struct {
		value *domain.FlightDetails
		err   error
	}
	type weatherResult struct {
		value *domain.WeatherDetails
		err   error
	}

	flightCh := make(chan flightResult, 1)
	weatherCh := make(chan weatherResult, 1)

	go func() {
		v, err := a.flightCB.Execute(func() (interface{}, error) {
			return a.flightFetcher.Fetch(ctx, flightNumber)
		})
		if err != nil {
			flightCh <- flightResult{err: err}
			return
		}
		flightCh <- flightResult{value: v.(*domain.FlightDetails)}
	}()

	go func() {
		v, err := a.weatherCB.Execute(func() (interface{}, error) {
			return a.weatherFetcher.Fetch(ctx, flightNumber)
		})
		if err != nil {
			weatherCh <- weatherResult{err: err}
			return
		}
		weatherCh <- weatherResult{value: v.(*domain.WeatherDetails)}
	}()

	var flight *domain.FlightDetails
	var weather *domain.WeatherDetails

	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			return a.tryStale(ctx, key, ctx.Err())
		case fr := <-flightCh:
			if fr.err != nil {
				return a.tryStale(ctx, key, fr.err)
			}
			flight = fr.value
		case wr := <-weatherCh:
			if wr.err != nil {
				return a.tryStale(ctx, key, wr.err)
			}
			weather = wr.value
		}
	}

	result := &domain.AggregatedResponse{
		Flight:      flight,
		Weather:     weather,
		Cached:      false,
		GeneratedAt: time.Now().UTC(),
	}
	encoded, err := json.Marshal(result)
	if err == nil {
		_ = a.cache.Set(ctx, key, string(encoded), a.cacheTTL)
	}
	return result, nil
}

func (a *Aggregator) tryStale(ctx context.Context, key string, sourceErr error) (*domain.AggregatedResponse, error) {
	cached, err := a.cache.Get(ctx, key)
	if err != nil {
		return nil, sourceErr
	}
	var out domain.AggregatedResponse
	if err := json.Unmarshal([]byte(cached), &out); err != nil {
		return nil, sourceErr
	}
	out.Cached = true
	return &out, nil
}
