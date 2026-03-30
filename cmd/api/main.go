package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/GerindT/GoFlight/internal/external"
	"github.com/GerindT/GoFlight/internal/handlers"
	"github.com/GerindT/GoFlight/internal/middleware"
	"github.com/GerindT/GoFlight/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	middleware.RegisterMetrics()

	ttl := 300 * time.Second
	if raw := os.Getenv("CACHE_TTL"); raw != "" {
		if sec, err := strconv.Atoi(raw); err == nil {
			ttl = time.Duration(sec) * time.Second
		}
	}

	redisCache, err := external.NewRedisCache(env("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		log.Fatalf("redis setup failed: %v", err)
	}

	httpClient := &http.Client{Timeout: 3 * time.Second}
	weatherLat, err := strconv.ParseFloat(env("OPENWEATHER_LAT", "50.1109"), 64)
	if err != nil {
		log.Fatalf("invalid OPENWEATHER_LAT: %v", err)
	}
	weatherLon, err := strconv.ParseFloat(env("OPENWEATHER_LON", "8.6821"), 64)
	if err != nil {
		log.Fatalf("invalid OPENWEATHER_LON: %v", err)
	}

	flightClient := external.NewAviationStackClient(
		env("AVIATIONSTACK_BASE_URL", "http://api.aviationstack.com/v1"),
		env("AVIATIONSTACK_API_KEY", ""),
		httpClient,
	)
	weatherClient := external.NewOpenWeatherClient(
		env("OPENWEATHER_BASE_URL", "https://api.open-meteo.com/v1"),
		env("DEFAULT_WEATHER_CITY", "Frankfurt"),
		weatherLat,
		weatherLon,
		httpClient,
	)

	aggregator := services.NewAggregator(flightClient, weatherClient, redisCache, ttl)
	handler := handlers.NewFlightHandler(aggregator)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.PrometheusMiddleware())

	r.GET("/api/v1/dashboard/:flight", handler.GetDashboard)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()
		if err := redisCache.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	srv := &http.Server{
		Addr:    ":" + env("PORT", "8080"),
		Handler: r,
	}

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("server stopped")
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
