package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

// LabelHandler handles HTTP requests for labels
type LabelHandler struct {
	service *service.LabelService
	logger  *zap.Logger
}

// NewLabelHandler creates a new LabelHandler
func NewLabelHandler(service *service.LabelService, logger *zap.Logger) *LabelHandler {
	return &LabelHandler{
		service: service,
		logger:  logger,
	}
}

// List handles GET /labels
func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	labels, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("failed to list labels", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list labels", nil)
		return
	}

	writeList(w, labels, len(labels))
}

// Get handles GET /labels/{id}
func (h *LabelHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid label ID", nil)
		return
	}

	label, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Label not found", nil)
			return
		}
		h.logger.Error("failed to get label", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get label", nil)
		return
	}

	writeJSON(w, http.StatusOK, label)
}

// Create handles POST /labels
func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.CreateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	label, err := h.service.Create(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrCodeExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Label name already exists", nil)
			return
		}
		h.logger.Error("failed to create label", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create label", nil)
		return
	}

	writeJSON(w, http.StatusCreated, label)
}

// Update handles PUT /labels/{id}
func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid label ID", nil)
		return
	}

	var req domain.UpdateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	label, err := h.service.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Label not found", nil)
			return
		}
		h.logger.Error("failed to update label", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update label", nil)
		return
	}

	writeJSON(w, http.StatusOK, label)
}

// Delete handles DELETE /labels/{id}
func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid label ID", nil)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Label not found", nil)
			return
		}
		h.logger.Error("failed to delete label", zap.String("id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete label", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignToLine handles PUT /lines/{id}/labels
func (h *LabelHandler) AssignToLine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	lineID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	var req domain.AssignLabelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.service.AssignToLine(ctx, lineID, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line or label not found", nil)
			return
		}
		h.logger.Error("failed to assign labels", zap.String("line_id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to assign labels", nil)
		return
	}

	// Return updated labels for the line
	labels, err := h.service.GetLabelsForLine(ctx, lineID)
	if err != nil {
		h.logger.Error("failed to get labels after assignment", zap.String("line_id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get labels", nil)
		return
	}

	writeJSON(w, http.StatusOK, labels)
}

// GetLabelsForLine handles GET /lines/{id}/labels
func (h *LabelHandler) GetLabelsForLine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	lineID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid production line ID", nil)
		return
	}

	labels, err := h.service.GetLabelsForLine(ctx, lineID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to get labels for line", zap.String("line_id", idStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get labels", nil)
		return
	}

	writeList(w, labels, len(labels))
}
