package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

var errMissingDateRange = errors.New("start_date and end_date are required")

// ComplianceHandler handles HTTP requests for compliance endpoints
type ComplianceHandler struct {
	complianceService *service.ComplianceService
	logger            *zap.Logger
}

// NewComplianceHandler creates a new ComplianceHandler
func NewComplianceHandler(complianceService *service.ComplianceService, logger *zap.Logger) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
		logger:            logger,
	}
}

// GetAggregateCompliance handles GET /api/v1/analytics/compliance
func (h *ComplianceHandler) GetAggregateCompliance(w http.ResponseWriter, r *http.Request) {
	query, err := h.parseComplianceQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := h.complianceService.GetAggregateCompliance(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to get aggregate compliance", zap.Error(err))
		// Check if it's a validation error
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "cannot exceed") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to calculate compliance metrics")
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// GetLineCompliance handles GET /api/v1/analytics/compliance/lines
func (h *ComplianceHandler) GetLineCompliance(w http.ResponseWriter, r *http.Request) {
	query, err := h.parseComplianceQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := h.complianceService.GetLineCompliance(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to get line compliance", zap.Error(err))
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "cannot exceed") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to calculate line compliance")
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// GetDailyComplianceKPIs handles GET /api/v1/analytics/lines/{id}/compliance/daily
func (h *ComplianceHandler) GetDailyComplianceKPIs(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lineID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid line ID")
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		respondError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	kpis, err := h.complianceService.GetDailyComplianceKPIs(r.Context(), lineID, startDate, endDate)
	if err != nil {
		h.logger.Error("failed to get daily compliance KPIs", zap.Error(err))
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "cannot exceed") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Line not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to calculate daily compliance")
		return
	}

	respondJSON(w, http.StatusOK, kpis)
}

// parseComplianceQuery extracts compliance query parameters from request
func (h *ComplianceHandler) parseComplianceQuery(r *http.Request) (domain.ComplianceQuery, error) {
	query := domain.ComplianceQuery{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	if query.StartDate == "" || query.EndDate == "" {
		return query, errMissingDateRange
	}

	// Parse line_ids if provided
	if lineIDsStr := r.URL.Query().Get("line_ids"); lineIDsStr != "" {
		ids := strings.Split(lineIDsStr, ",")
		for _, idStr := range ids {
			id, err := uuid.Parse(strings.TrimSpace(idStr))
			if err != nil {
				continue
			}
			query.LineIDs = append(query.LineIDs, id)
		}
	}

	// Parse label_ids if provided
	if labelIDsStr := r.URL.Query().Get("label_ids"); labelIDsStr != "" {
		ids := strings.Split(labelIDsStr, ",")
		for _, idStr := range ids {
			id, err := uuid.Parse(strings.TrimSpace(idStr))
			if err != nil {
				continue
			}
			query.LabelIDs = append(query.LabelIDs, id)
		}
	}

	return query, nil
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
