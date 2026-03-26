package handlers

import (
	"errors"
	"net/http"

	"github.com/GerindT/GoFlight/internal/apierrors"
	"github.com/GerindT/GoFlight/internal/services"
	"github.com/gin-gonic/gin"
)

type FlightHandler struct {
	aggregator *services.Aggregator
}

func NewFlightHandler(aggregator *services.Aggregator) *FlightHandler {
	return &FlightHandler{aggregator: aggregator}
}

func (h *FlightHandler) GetDashboard(c *gin.Context) {
	flight := c.Param("flight")
	if flight == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight parameter is required"})
		return
	}

	resp, err := h.aggregator.GetFlightDash(c.Request.Context(), flight)
	if err != nil {
		status := apierrors.StatusFor(err)
		msg := "service unavailable"
		if errors.Is(err, apierrors.ErrFlightNotFound) {
			msg = "flight not found"
		} else if errors.Is(err, apierrors.ErrUpstreamTimeout) {
			msg = "upstream timeout"
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}
