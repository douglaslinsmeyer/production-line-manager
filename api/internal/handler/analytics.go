package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

// AnalyticsHandler handles HTTP requests for analytics
type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *zap.Logger
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(service *service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

// GetAggregateMetrics handles GET /analytics/aggregate
func (h *AnalyticsHandler) GetAggregateMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query, err := h.parseAnalyticsQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error(), nil)
		return
	}

	metrics, err := h.service.GetAggregateMetrics(ctx, query)
	if err != nil {
		h.logger.Error("failed to get aggregate metrics", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get metrics", nil)
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// GetLineMetrics handles GET /analytics/lines
func (h *AnalyticsHandler) GetLineMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query, err := h.parseAnalyticsQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error(), nil)
		return
	}

	metrics, err := h.service.GetLineMetrics(ctx, query)
	if err != nil {
		h.logger.Error("failed to get line metrics", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get metrics", nil)
		return
	}

	writeList(w, metrics, len(metrics))
}

// GetLabelMetrics handles GET /analytics/labels
func (h *AnalyticsHandler) GetLabelMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query, err := h.parseAnalyticsQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error(), nil)
		return
	}

	metrics, err := h.service.GetLabelMetrics(ctx, query)
	if err != nil {
		h.logger.Error("failed to get label metrics", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get metrics", nil)
		return
	}

	writeList(w, metrics, len(metrics))
}

// GetDailyKPIs handles GET /analytics/lines/{id}/daily
func (h *AnalyticsHandler) GetDailyKPIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	lineID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	query, err := h.parseAnalyticsQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error(), nil)
		return
	}

	// Default to 7d for daily KPIs if not specified
	if query.Timeframe == "" {
		query.Timeframe = "7d"
	}

	kpis, err := h.service.GetDailyKPIs(ctx, lineID, query)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to get daily KPIs", zap.String("line_id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get KPIs", nil)
		return
	}

	writeList(w, kpis, len(kpis))
}

// parseAnalyticsQuery parses query parameters into AnalyticsQuery
func (h *AnalyticsHandler) parseAnalyticsQuery(r *http.Request) (domain.AnalyticsQuery, error) {
	query := domain.AnalyticsQuery{
		Timeframe: r.URL.Query().Get("timeframe"),
	}

	// Parse start_time if provided
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return query, errors.New("invalid start_time format (use RFC3339)")
		}
		query.StartTime = &start
	}

	// Parse end_time if provided
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return query, errors.New("invalid end_time format (use RFC3339)")
		}
		query.EndTime = &end
	}

	// Parse label_ids if provided
	if labelIDsStr := r.URL.Query().Get("label_ids"); labelIDsStr != "" {
		labelIDs, err := parseUUIDList(labelIDsStr)
		if err != nil {
			return query, errors.New("invalid label_ids format")
		}
		query.LabelIDs = labelIDs
	}

	// Parse line_ids if provided
	if lineIDsStr := r.URL.Query().Get("line_ids"); lineIDsStr != "" {
		lineIDs, err := parseUUIDList(lineIDsStr)
		if err != nil {
			return query, errors.New("invalid line_ids format")
		}
		query.LineIDs = lineIDs
	}

	return query, nil
}

// parseUUIDList parses comma-separated UUID string into []uuid.UUID
func parseUUIDList(s string) ([]uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	uuids := make([]uuid.UUID, 0, len(parts))

	for _, part := range parts {
		id, err := uuid.Parse(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, id)
	}

	return uuids, nil
}
