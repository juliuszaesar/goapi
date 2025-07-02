package models

import (
	"time"
)

// Reminder represents a reminder entity
type Reminder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReminderRequest represents the request payload for creating/updating reminders
type ReminderRequest struct {
	Title   string `json:"title" form:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" form:"content" validate:"max=1000"`
}

// ReminderResponse represents the response payload for reminders
type ReminderResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchRequest represents the search request payload
type SearchRequest struct {
	Query string `json:"query" form:"query" validate:"required,min=1"`
	Limit int    `json:"limit" form:"limit"`
}

// SearchResponse represents the search response payload
type SearchResponse struct {
	Results []ReminderResponse `json:"results"`
	Total   int                `json:"total"`
	Query   string             `json:"query"`
}

// ToResponse converts a Reminder to ReminderResponse
func (r *Reminder) ToResponse() ReminderResponse {
	return ReminderResponse{
		ID:        r.ID,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
