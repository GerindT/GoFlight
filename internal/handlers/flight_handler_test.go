package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GerindT/GoFlight/internal/apierrors"
	"github.com/GerindT/GoFlight/internal/domain"
	"github.com/GerindT/GoFlight/internal/services"
	"github.com/GerindT/GoFlight/mock"
	"github.com/gin-gonic/gin"
)

type testCache struct {
	store map[string]string
}

func (t *testCache) Get(_ context.Context, key string) (string, error) {
	v, ok := t.store[key]
	if !ok {
		return "", apierrors.ErrCacheMiss
	}
	return v, nil
}

func (t *testCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	t.store[key] = value
	return nil
}

func (t *testCache) Ping(_ context.Context) error { return nil }

func TestGetDashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid flight returns 200", func(t *testing.T) {
		flight := &mock.MockFlightFetcher{ReturnFlight: &domain.FlightDetails{FlightNumber: "LH123"}}
		weather := &mock.MockWeatherFetcher{ReturnWeather: &domain.WeatherDetails{Location: "Frankfurt"}}
		agg := services.NewAggregator(flight, weather, &testCache{store: map[string]string{}}, 5*time.Minute)
		h := NewFlightHandler(agg)

		r := gin.New()
		r.GET("/api/v1/dashboard/:flight", h.GetDashboard)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/LH123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", w.Code)
		}
	})

	t.Run("unknown flight returns 404", func(t *testing.T) {
		flight := &mock.MockFlightFetcher{ReturnErr: apierrors.ErrFlightNotFound}
		weather := &mock.MockWeatherFetcher{ReturnWeather: &domain.WeatherDetails{Location: "Frankfurt"}}
		agg := services.NewAggregator(flight, weather, &testCache{store: map[string]string{}}, 5*time.Minute)
		h := NewFlightHandler(agg)

		r := gin.New()
		r.GET("/api/v1/dashboard/:flight", h.GetDashboard)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/ZZ999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", w.Code)
		}
	})

	t.Run("upstream error returns 503 style error JSON", func(t *testing.T) {
		flight := &mock.MockFlightFetcher{ReturnErr: apierrors.ErrUpstreamBadData}
		weather := &mock.MockWeatherFetcher{ReturnWeather: &domain.WeatherDetails{Location: "Frankfurt"}}
		agg := services.NewAggregator(flight, weather, &testCache{store: map[string]string{}}, 5*time.Minute)
		h := NewFlightHandler(agg)

		r := gin.New()
		r.GET("/api/v1/dashboard/:flight", h.GetDashboard)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/LH123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadGateway {
			t.Fatalf("expected 502 got %d", w.Code)
		}

		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("expected JSON body: %v", err)
		}
		if body["error"] == "" {
			t.Fatalf("expected redacted error message")
		}
	})
}
