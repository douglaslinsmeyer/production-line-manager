package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

// LineHandler handles HTTP requests for production lines
type LineHandler struct {
	service *service.LineService
	logger  *zap.Logger
}

// NewLineHandler creates a new LineHandler
func NewLineHandler(service *service.LineService, logger *zap.Logger) *LineHandler {
	return &LineHandler{
		service: service,
		logger:  logger,
	}
}

// List godoc
// @Summary List production lines
// @Description Get all active production lines, optionally filtered by code
// @Tags lines
// @Accept json
// @Produce json
// @Param code query string false "Filter by line code"
// @Success 200 {object} Response{data=[]domain.ProductionLine}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines [get]
func (h *LineHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if filtering by code
	code := r.URL.Query().Get("code")
	if code != "" {
		// Get single line by code
		line, err := h.service.GetByCode(ctx, code)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
				return
			}
			h.logger.Error("failed to get production line by code", zap.String("code", code), zap.Error(err))
			writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get production line", nil)
			return
		}
		// Return as array to match expected response format
		writeList(w, []domain.ProductionLine{*line}, 1)
		return
	}

	// List all lines
	lines, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("failed to list production lines", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list production lines", nil)
		return
	}

	writeList(w, lines, len(lines))
}

// Get godoc
// @Summary Get production line
// @Description Get a production line by ID
// @Tags lines
// @Accept json
// @Produce json
// @Param id path string true "Production Line ID (UUID)"
// @Success 200 {object} Response{data=domain.ProductionLine}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines/{id} [get]
func (h *LineHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	// Get production line
	line, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to get production line", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get production line", nil)
		return
	}

	writeJSON(w, http.StatusOK, line)
}

// Create godoc
// @Summary Create production line
// @Description Create a new production line
// @Tags lines
// @Accept json
// @Produce json
// @Param line body domain.CreateLineRequest true "Production line details"
// @Success 201 {object} Response{data=domain.ProductionLine}
// @Failure 400 {object} Response{error=APIError}
// @Failure 409 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines [post]
func (h *LineHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req domain.CreateLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Create production line
	line, err := h.service.Create(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrCodeExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Production line code already exists", nil)
			return
		}
		h.logger.Error("failed to create production line", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create production line", nil)
		return
	}

	writeJSON(w, http.StatusCreated, line)
}

// Update godoc
// @Summary Update production line
// @Description Update a production line's details
// @Tags lines
// @Accept json
// @Produce json
// @Param id path string true "Production Line ID (UUID)"
// @Param line body domain.UpdateLineRequest true "Updated details"
// @Success 200 {object} Response{data=domain.ProductionLine}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines/{id} [put]
func (h *LineHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	// Parse request body
	var req domain.UpdateLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Update production line
	line, err := h.service.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to update production line", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update production line", nil)
		return
	}

	writeJSON(w, http.StatusOK, line)
}

// Delete godoc
// @Summary Delete production line
// @Description Soft delete a production line
// @Tags lines
// @Accept json
// @Produce json
// @Param id path string true "Production Line ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines/{id} [delete]
func (h *LineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	// Delete production line
	if err := h.service.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to delete production line", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete production line", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetStatus godoc
// @Summary Set production line status
// @Description Set the status of a production line
// @Tags lines
// @Accept json
// @Produce json
// @Param id path string true "Production Line ID (UUID)"
// @Param status body domain.SetStatusRequest true "Status details"
// @Success 200 {object} Response{data=domain.ProductionLine}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines/{id}/status [post]
func (h *LineHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	// Parse request body
	var req domain.SetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	// Validate status
	if !req.Status.IsValid() {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "Invalid status value", nil)
		return
	}

	// Set status
	sourceDetail := map[string]interface{}{
		"user_agent": r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	}

	line, err := h.service.SetStatus(ctx, id, req.Status, "api", sourceDetail)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, "Invalid status value", nil)
			return
		}
		h.logger.Error("failed to set production line status", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to set status", nil)
		return
	}

	writeJSON(w, http.StatusOK, line)
}

// GetStatusHistory godoc
// @Summary Get status history
// @Description Get the status change history for a production line
// @Tags lines
// @Accept json
// @Produce json
// @Param id path string true "Production Line ID (UUID)"
// @Param limit query int false "Maximum number of records to return (default: 100)"
// @Success 200 {object} Response{data=[]domain.StatusChange}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /lines/{id}/status/history [get]
func (h *LineHandler) GetStatusHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	// Parse limit from query params
	limit := 100 // Default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get status history
	history, err := h.service.GetStatusHistory(ctx, id, limit)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to get status history", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get status history", nil)
		return
	}

	writeList(w, history, len(history))
}
