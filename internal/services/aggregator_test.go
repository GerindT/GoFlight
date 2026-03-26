package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/GerindT/GoFlight/internal/apierrors"
	"github.com/GerindT/GoFlight/internal/domain"
	"github.com/GerindT/GoFlight/mock"
)

type mockCache struct {
	store         map[string]string
	missFirstRead bool
	readCount     int
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	m.readCount++
	if m.missFirstRead && m.readCount == 1 {
		return "", apierrors.ErrCacheMiss
	}
	v, ok := m.store[key]
	if !ok {
		return "", apierrors.ErrCacheMiss
	}
	return v, nil
}

func (m *mockCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockCache) Ping(_ context.Context) error {
	return nil
}

func TestGetFlightDash_TableDriven(t *testing.T) {
	baseFlight := &domain.FlightDetails{FlightNumber: "LH123", Status: "scheduled"}
	baseWeather := &domain.WeatherDetails{Location: "Frankfurt", Condition: "clear"}

	makeCachePayload := func() string {
		payload, _ := json.Marshal(domain.AggregatedResponse{
			Flight:      baseFlight,
			Weather:     baseWeather,
			Cached:      false,
			GeneratedAt: time.Now().UTC(),
		})
		return string(payload)
	}

	tests := []struct {
		name             string
		cacheValue       string
		flightErr        error
		expectErr        bool
		expectCached     bool
		expectFlightCall int
		expectWeathCall  int
	}{
		{
			name:             "happy path",
			expectErr:        false,
			expectCached:     false,
			expectFlightCall: 1,
			expectWeathCall:  1,
		},
		{
			name:             "cache hit short circuits fetchers",
			cacheValue:       makeCachePayload(),
			expectErr:        false,
			expectCached:     true,
			expectFlightCall: 0,
			expectWeathCall:  0,
		},
		{
			name:             "upstream error serves stale cache",
			cacheValue:       makeCachePayload(),
			flightErr:        apierrors.ErrUpstreamTimeout,
			expectErr:        false,
			expectCached:     true,
			expectFlightCall: 1,
			expectWeathCall:  1,
		},
		{
			name:             "upstream error without stale cache",
			flightErr:        apierrors.ErrUpstreamTimeout,
			expectErr:        true,
			expectCached:     false,
			expectFlightCall: 1,
			expectWeathCall:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mockCache{store: map[string]string{}}
			if tt.cacheValue != "" {
				c.store["flight_dash:LH123"] = tt.cacheValue
			}
			if tt.cacheValue != "" && tt.flightErr != nil {
				c.missFirstRead = true
			}

			fetcher := &mock.MockFlightFetcher{ReturnFlight: baseFlight, ReturnErr: tt.flightErr}
			weather := &mock.MockWeatherFetcher{ReturnWeather: baseWeather}
			agg := NewAggregator(fetcher, weather, c, 5*time.Minute)

			resp, err := agg.GetFlightDash(context.Background(), "LH123")
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if err == nil && resp.Cached != tt.expectCached {
				t.Fatalf("expected cached=%v got=%v", tt.expectCached, resp.Cached)
			}
			if fetcher.Calls != tt.expectFlightCall {
				t.Fatalf("expected flight calls=%d got=%d", tt.expectFlightCall, fetcher.Calls)
			}
			if weather.Calls != tt.expectWeathCall {
				t.Fatalf("expected weather calls=%d got=%d", tt.expectWeathCall, weather.Calls)
			}
			if tt.flightErr != nil && tt.expectErr {
				if !errors.Is(err, apierrors.ErrUpstreamTimeout) {
					t.Fatalf("expected timeout error, got %v", err)
				}
			}
		})
	}
}
