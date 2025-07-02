package handlers

import (
	"goapi/internal/models"
	"goapi/internal/search"
	"goapi/internal/services"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ReminderHandler handles reminder-related HTTP requests
type ReminderHandler struct {
	reminderService services.ReminderService
	searchService   search.SearchService
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService services.ReminderService, searchService search.SearchService) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
		searchService:   searchService,
	}
}



// CreateReminder handles POST /reminders
func (h *ReminderHandler) CreateReminder(c echo.Context) error {
	var req models.ReminderRequest

	// Bind form data or JSON
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate required fields
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
	}

	// Create reminder
	reminder, err := h.reminderService.CreateReminder(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create reminder"})
	}

	// Index in search engine
	if err := h.searchService.IndexReminder(reminder); err != nil {
		// Log error but don't fail the request
		c.Logger().Errorf("Failed to index reminder in search: %v", err)
	}

	return c.JSON(http.StatusCreated, reminder.ToResponse())
}

// GetReminders handles GET /reminders
func (h *ReminderHandler) GetReminders(c echo.Context) error {
	reminders, err := h.reminderService.GetReminders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch reminders"})
	}

	var responses []models.ReminderResponse
	for _, reminder := range reminders {
		responses = append(responses, reminder.ToResponse())
	}

	return c.JSON(http.StatusOK, responses)
}

// GetReminder handles GET /reminders/:id
func (h *ReminderHandler) GetReminder(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid reminder ID"})
	}

	reminder, err := h.reminderService.GetReminderByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Reminder not found"})
	}

	return c.JSON(http.StatusOK, reminder.ToResponse())
}

// UpdateReminder handles PUT /reminders/:id
func (h *ReminderHandler) UpdateReminder(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid reminder ID"})
	}

	var req models.ReminderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title is required"})
	}

	reminder, err := h.reminderService.UpdateReminder(uint(id), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update reminder"})
	}

	// Update in search engine
	if err := h.searchService.UpdateReminder(reminder); err != nil {
		c.Logger().Errorf("Failed to update reminder in search: %v", err)
	}

	return c.JSON(http.StatusOK, reminder.ToResponse())
}

// DeleteReminder handles DELETE /reminders/:id
func (h *ReminderHandler) DeleteReminder(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid reminder ID"})
	}

	if err := h.reminderService.DeleteReminder(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete reminder"})
	}

	// Remove from search engine
	if err := h.searchService.DeleteReminder(uint(id)); err != nil {
		c.Logger().Errorf("Failed to delete reminder from search: %v", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchReminders handles GET /search
func (h *ReminderHandler) SearchReminders(c echo.Context) error {
	query := c.QueryParam("query")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Query parameter is required"})
	}

	limitParam := c.QueryParam("limit")
	limit := 10 // default
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	results, err := h.searchService.SearchReminders(query, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Search failed"})
	}

	return c.JSON(http.StatusOK, results)
}
