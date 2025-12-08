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

// ScheduleHandler handles HTTP requests for schedules
type ScheduleHandler struct {
	service *service.ScheduleService
	logger  *zap.Logger
}

// NewScheduleHandler creates a new ScheduleHandler
func NewScheduleHandler(service *service.ScheduleService, logger *zap.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		service: service,
		logger:  logger,
	}
}

// ========== Schedule CRUD ==========

// List returns all schedules
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	schedules, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("failed to list schedules", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list schedules", nil)
		return
	}

	writeList(w, schedules, len(schedules))
}

// Get returns a schedule by ID
func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	schedule, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule not found", nil)
			return
		}
		h.logger.Error("failed to get schedule", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get schedule", nil)
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

// Create creates a new schedule
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	schedule, err := h.service.Create(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNameExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Schedule name already exists", nil)
			return
		}
		h.logger.Error("failed to create schedule", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create schedule", nil)
		return
	}

	writeJSON(w, http.StatusCreated, schedule)
}

// Update updates a schedule
func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	var req domain.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	schedule, err := h.service.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule not found", nil)
			return
		}
		if errors.Is(err, domain.ErrScheduleNameExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Schedule name already exists", nil)
			return
		}
		h.logger.Error("failed to update schedule", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update schedule", nil)
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

// Delete soft deletes a schedule
func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule not found", nil)
			return
		}
		h.logger.Error("failed to delete schedule", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete schedule", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Schedule Day Operations ==========

// GetDay returns a schedule day
func (h *ScheduleHandler) GetDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dayID, err := uuid.Parse(chi.URLParam(r, "dayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid day ID", nil)
		return
	}

	day, err := h.service.GetDay(ctx, dayID)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleDayNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule day not found", nil)
			return
		}
		h.logger.Error("failed to get schedule day", zap.String("day_id", dayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get schedule day", nil)
		return
	}

	writeJSON(w, http.StatusOK, day)
}

// UpdateDay updates a schedule day
func (h *ScheduleHandler) UpdateDay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dayID, err := uuid.Parse(chi.URLParam(r, "dayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid day ID", nil)
		return
	}

	var req domain.UpdateScheduleDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	day, err := h.service.UpdateDay(ctx, dayID, req)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleDayNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule day not found", nil)
			return
		}
		h.logger.Error("failed to update schedule day", zap.String("day_id", dayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update schedule day", nil)
		return
	}

	writeJSON(w, http.StatusOK, day)
}

// SetDayBreaks replaces all breaks for a day
func (h *ScheduleHandler) SetDayBreaks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dayID, err := uuid.Parse(chi.URLParam(r, "dayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid day ID", nil)
		return
	}

	var req domain.SetScheduleBreaksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	breaks, err := h.service.SetDayBreaks(ctx, dayID, req)
	if err != nil {
		h.logger.Error("failed to set day breaks", zap.String("day_id", dayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to set day breaks", nil)
		return
	}

	writeJSON(w, http.StatusOK, breaks)
}

// ========== Holiday Operations ==========

// ListHolidays returns all holidays for a schedule
func (h *ScheduleHandler) ListHolidays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	holidays, err := h.service.GetHolidays(ctx, scheduleID)
	if err != nil {
		h.logger.Error("failed to list holidays", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list holidays", nil)
		return
	}

	writeList(w, holidays, len(holidays))
}

// CreateHoliday creates a holiday
func (h *ScheduleHandler) CreateHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	var req domain.CreateHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	holiday, err := h.service.CreateHoliday(ctx, scheduleID, req)
	if err != nil {
		if errors.Is(err, domain.ErrHolidayDateExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Holiday already exists for this date", nil)
			return
		}
		h.logger.Error("failed to create holiday", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create holiday", nil)
		return
	}

	writeJSON(w, http.StatusCreated, holiday)
}

// GetHoliday returns a holiday by ID
func (h *ScheduleHandler) GetHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	holidayID, err := uuid.Parse(chi.URLParam(r, "holidayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid holiday ID", nil)
		return
	}

	holiday, err := h.service.GetHoliday(ctx, holidayID)
	if err != nil {
		if errors.Is(err, domain.ErrHolidayNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Holiday not found", nil)
			return
		}
		h.logger.Error("failed to get holiday", zap.String("holiday_id", holidayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get holiday", nil)
		return
	}

	writeJSON(w, http.StatusOK, holiday)
}

// UpdateHoliday updates a holiday
func (h *ScheduleHandler) UpdateHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	holidayID, err := uuid.Parse(chi.URLParam(r, "holidayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid holiday ID", nil)
		return
	}

	var req domain.UpdateHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	holiday, err := h.service.UpdateHoliday(ctx, holidayID, req)
	if err != nil {
		if errors.Is(err, domain.ErrHolidayNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Holiday not found", nil)
			return
		}
		h.logger.Error("failed to update holiday", zap.String("holiday_id", holidayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update holiday", nil)
		return
	}

	writeJSON(w, http.StatusOK, holiday)
}

// DeleteHoliday deletes a holiday
func (h *ScheduleHandler) DeleteHoliday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	holidayID, err := uuid.Parse(chi.URLParam(r, "holidayId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid holiday ID", nil)
		return
	}

	if err := h.service.DeleteHoliday(ctx, holidayID); err != nil {
		if errors.Is(err, domain.ErrHolidayNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Holiday not found", nil)
			return
		}
		h.logger.Error("failed to delete holiday", zap.String("holiday_id", holidayID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete holiday", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Schedule Exception Operations ==========

// ListExceptions returns all exceptions for a schedule
func (h *ScheduleHandler) ListExceptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	exceptions, err := h.service.GetExceptions(ctx, scheduleID)
	if err != nil {
		h.logger.Error("failed to list exceptions", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list exceptions", nil)
		return
	}

	writeList(w, exceptions, len(exceptions))
}

// CreateException creates a schedule exception
func (h *ScheduleHandler) CreateException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	var req domain.CreateScheduleExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	exception, err := h.service.CreateException(ctx, scheduleID, req)
	if err != nil {
		if errors.Is(err, domain.ErrExceptionDatesOverlap) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Exception dates overlap with existing exception", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, "Start date must be before or equal to end date", nil)
			return
		}
		h.logger.Error("failed to create exception", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create exception", nil)
		return
	}

	writeJSON(w, http.StatusCreated, exception)
}

// GetException returns a schedule exception by ID
func (h *ScheduleHandler) GetException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	exception, err := h.service.GetException(ctx, exceptionID)
	if err != nil {
		if errors.Is(err, domain.ErrExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Exception not found", nil)
			return
		}
		h.logger.Error("failed to get exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get exception", nil)
		return
	}

	writeJSON(w, http.StatusOK, exception)
}

// UpdateException updates a schedule exception
func (h *ScheduleHandler) UpdateException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	var req domain.UpdateScheduleExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	exception, err := h.service.UpdateException(ctx, exceptionID, req)
	if err != nil {
		if errors.Is(err, domain.ErrExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Exception not found", nil)
			return
		}
		h.logger.Error("failed to update exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update exception", nil)
		return
	}

	writeJSON(w, http.StatusOK, exception)
}

// DeleteException deletes a schedule exception
func (h *ScheduleHandler) DeleteException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	if err := h.service.DeleteException(ctx, exceptionID); err != nil {
		if errors.Is(err, domain.ErrExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Exception not found", nil)
			return
		}
		h.logger.Error("failed to delete exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete exception", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Line Schedule Exception Operations ==========

// ListLineExceptions returns all line-specific exceptions for a schedule
func (h *ScheduleHandler) ListLineExceptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	exceptions, err := h.service.GetLineExceptions(ctx, scheduleID)
	if err != nil {
		h.logger.Error("failed to list line exceptions", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list line exceptions", nil)
		return
	}

	writeList(w, exceptions, len(exceptions))
}

// CreateLineException creates a line-specific exception
func (h *ScheduleHandler) CreateLineException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	var req domain.CreateLineScheduleExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	exception, err := h.service.CreateLineException(ctx, scheduleID, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDateRange) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, "Start date must be before or equal to end date", nil)
			return
		}
		if errors.Is(err, domain.ErrMissingLinesToException) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, "At least one line must be specified", nil)
			return
		}
		h.logger.Error("failed to create line exception", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create line exception", nil)
		return
	}

	writeJSON(w, http.StatusCreated, exception)
}

// GetLineException returns a line-specific exception by ID
func (h *ScheduleHandler) GetLineException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	exception, err := h.service.GetLineException(ctx, exceptionID)
	if err != nil {
		if errors.Is(err, domain.ErrLineExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Line exception not found", nil)
			return
		}
		h.logger.Error("failed to get line exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get line exception", nil)
		return
	}

	writeJSON(w, http.StatusOK, exception)
}

// UpdateLineException updates a line-specific exception
func (h *ScheduleHandler) UpdateLineException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	var req domain.UpdateLineScheduleExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	exception, err := h.service.UpdateLineException(ctx, exceptionID, req)
	if err != nil {
		if errors.Is(err, domain.ErrLineExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Line exception not found", nil)
			return
		}
		h.logger.Error("failed to update line exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update line exception", nil)
		return
	}

	writeJSON(w, http.StatusOK, exception)
}

// DeleteLineException deletes a line-specific exception
func (h *ScheduleHandler) DeleteLineException(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid exception ID", nil)
		return
	}

	if err := h.service.DeleteLineException(ctx, exceptionID); err != nil {
		if errors.Is(err, domain.ErrLineExceptionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Line exception not found", nil)
			return
		}
		h.logger.Error("failed to delete line exception", zap.String("exception_id", exceptionID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete line exception", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Line Assignment ==========

// GetLinesForSchedule returns all lines assigned to a schedule
func (h *ScheduleHandler) GetLinesForSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scheduleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid schedule ID", nil)
		return
	}

	lines, err := h.service.GetLinesForSchedule(ctx, scheduleID)
	if err != nil {
		h.logger.Error("failed to get lines for schedule", zap.String("schedule_id", scheduleID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get lines for schedule", nil)
		return
	}

	writeList(w, lines, len(lines))
}

// AssignScheduleToLine assigns a schedule to a production line
func (h *ScheduleHandler) AssignScheduleToLine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid line ID", nil)
		return
	}

	var req domain.AssignScheduleToLineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.service.AssignScheduleToLine(ctx, lineID, req.ScheduleID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		if errors.Is(err, domain.ErrScheduleNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule not found", nil)
			return
		}
		h.logger.Error("failed to assign schedule to line", zap.String("line_id", lineID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to assign schedule to line", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Effective Schedule ==========

// GetEffectiveSchedule returns the effective schedule for a line on a date
func (h *ScheduleHandler) GetEffectiveSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lineIDStr := r.URL.Query().Get("line_id")
	if lineIDStr == "" {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "line_id is required", nil)
		return
	}

	lineID, err := uuid.Parse(lineIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid line ID", nil)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "date is required", nil)
		return
	}

	effective, err := h.service.GetEffectiveSchedule(ctx, lineID, date)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		h.logger.Error("failed to get effective schedule", zap.String("line_id", lineID.String()), zap.String("date", date), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get effective schedule", nil)
		return
	}

	writeJSON(w, http.StatusOK, effective)
}

// GetEffectiveScheduleRange returns effective schedules for a line over a date range
func (h *ScheduleHandler) GetEffectiveScheduleRange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lineIDStr := r.URL.Query().Get("line_id")
	if lineIDStr == "" {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "line_id is required", nil)
		return
	}

	lineID, err := uuid.Parse(lineIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid line ID", nil)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	if startDate == "" {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "start_date is required", nil)
		return
	}

	endDate := r.URL.Query().Get("end_date")
	if endDate == "" {
		writeError(w, http.StatusBadRequest, ErrCodeValidation, "end_date is required", nil)
		return
	}

	effectives, err := h.service.GetEffectiveScheduleRange(ctx, lineID, startDate, endDate)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Production line not found", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidDateRange) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, "Start date must be before or equal to end date", nil)
			return
		}
		h.logger.Error("failed to get effective schedule range", zap.String("line_id", lineID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get effective schedule range", nil)
		return
	}

	writeList(w, effectives, len(effectives))
}
