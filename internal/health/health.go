package health

import (
	"net/http"
	"time"

	"github.com/shubie/trading/internal/aggregator"
)

type Handler struct {
	agg     *aggregator.Aggregator
	timeout time.Duration
}

func NewHandler(agg *aggregator.Aggregator, timeout time.Duration) *Handler {
	return &Handler{
		agg:     agg,
		timeout: timeout,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lastData := h.agg.GetLastDataTime()
	if time.Since(lastData) > h.timeout {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("UNHEALTHY: No recent data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
