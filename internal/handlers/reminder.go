package handlers

import (
	"fmt"
	"goapi/internal/models"
	"goapi/internal/search"
	"goapi/internal/services"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// formatDate formats a time.Time to a human-readable string
func formatDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	return t.Format("Jan 2, 2006")
}

// renderReminderHTML renders a single reminder as HTML
func renderReminderHTML(reminder models.ReminderResponse) string {
	return fmt.Sprintf(`
		<div class="reminder-item fade-in">
			<div class="reminder-header">
				<div>
					<div class="reminder-title">%s</div>
					<div class="reminder-date">
						<i class="fas fa-clock"></i> %s
					</div>
				</div>
			</div>
			<div class="reminder-content">%s</div>
			<div class="reminder-actions">
				<button class="btn btn-small btn-danger" onclick="deleteReminder(%d)">
					<i class="fas fa-trash"></i> Delete
				</button>
			</div>
		</div>`,
		template.HTMLEscapeString(reminder.Title),
		formatDate(reminder.CreatedAt),
		template.HTMLEscapeString(reminder.Content),
		reminder.ID,
	)
}

// renderRemindersListHTML renders a list of reminders as HTML
func renderRemindersListHTML(reminders []models.ReminderResponse) string {
	if len(reminders) == 0 {
		return `<div class="no-results">
			<i class="fas fa-inbox" style="font-size: 3rem; margin-bottom: 15px; opacity: 0.5;"></i>
			<p>No reminders yet. Create your first reminder above!</p>
		</div>`
	}

	var html strings.Builder
	for _, reminder := range reminders {
		html.WriteString(renderReminderHTML(reminder))
	}
	return html.String()
}

// renderSearchResultsHTML renders search results as HTML
func renderSearchResultsHTML(results *models.SearchResponse) string {
	if results.Total == 0 {
		return `<div class="no-results">
			<i class="fas fa-search" style="font-size: 2rem; margin-bottom: 10px; opacity: 0.5;"></i>
			<p>No reminders found for your search.</p>
		</div>`
	}

	var html strings.Builder
	html.WriteString(fmt.Sprintf(`<h3><i class="fas fa-search-plus"></i> Found %d result(s) for "%s"</h3>`,
		results.Total, template.HTMLEscapeString(results.Query)))

	for _, reminder := range results.Results {
		html.WriteString(renderReminderHTML(reminder))
	}
	return html.String()
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

	// Check if request wants HTML (from HTMX)
	if c.Request().Header.Get("HX-Request") == "true" {
		html := renderReminderHTML(reminder.ToResponse())
		return c.HTML(http.StatusCreated, html)
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

	// Check if request wants HTML (from HTMX)
	if c.Request().Header.Get("HX-Request") == "true" {
		html := renderRemindersListHTML(responses)
		return c.HTML(http.StatusOK, html)
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
		// Return empty results for empty query instead of error
		if c.Request().Header.Get("HX-Request") == "true" {
			return c.HTML(http.StatusOK, "")
		}
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
		if c.Request().Header.Get("HX-Request") == "true" {
			return c.HTML(http.StatusOK, `<div class="no-results">
				<i class="fas fa-exclamation-triangle" style="font-size: 2rem; margin-bottom: 10px; opacity: 0.5;"></i>
				<p>Search failed. Please try again.</p>
			</div>`)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Search failed"})
	}

	// Check if request wants HTML (from HTMX)
	if c.Request().Header.Get("HX-Request") == "true" {
		html := renderSearchResultsHTML(results)
		return c.HTML(http.StatusOK, html)
	}

	return c.JSON(http.StatusOK, results)
}
