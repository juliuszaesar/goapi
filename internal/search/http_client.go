package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goapi/internal/config"
	"goapi/internal/models"
	"io"
	"net/http"
	"strconv"
	"time"
)

// httpSearchService implements SearchService using HTTP client
type httpSearchService struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewHTTPSearchService creates a new HTTP-based search service
func NewHTTPSearchService(cfg *config.TypesenseConfig) SearchService {
	return &httpSearchService{
		baseURL: cfg.GetTypesenseURL(),
		apiKey:  cfg.APIKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitializeSchema creates the reminders collection schema
func (s *httpSearchService) InitializeSchema() error {
	// Check if collection exists
	req, err := http.NewRequest("GET", s.baseURL+"/collections/"+RemindersCollection, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		return nil // Collection exists
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Create collection
	schema := map[string]interface{}{
		"name": RemindersCollection,
		"fields": []map[string]interface{}{
			{"name": "id", "type": "string"},
			{"name": "title", "type": "string"},
			{"name": "content", "type": "string"},
			{"name": "created_at", "type": "int64"},
		},
		"default_sorting_field": "created_at",
	}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", s.baseURL+"/collections", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err = s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IndexReminder indexes a reminder in Typesense
func (s *httpSearchService) IndexReminder(reminder *models.Reminder) error {
	document := map[string]interface{}{
		"id":         strconv.Itoa(int(reminder.ID)),
		"title":      reminder.Title,
		"content":    reminder.Content,
		"created_at": reminder.CreatedAt.Unix(),
	}

	jsonData, err := json.Marshal(document)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.baseURL+"/collections/"+RemindersCollection+"/documents", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to index reminder: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to index reminder: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SearchReminders searches for reminders using Typesense
func (s *httpSearchService) SearchReminders(query string, limit int) (*models.SearchResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	url := fmt.Sprintf("%s/collections/%s/documents/search?q=%s&query_by=title,content&per_page=%d&sort_by=created_at:desc",
		s.baseURL, RemindersCollection, query, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search reminders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Hits  []struct {
			Document map[string]interface{} `json:"document"`
		} `json:"hits"`
		Found int `json:"found"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var reminders []models.ReminderResponse
	for _, hit := range result.Hits {
		var reminder models.ReminderResponse

		if idStr, ok := hit.Document["id"].(string); ok {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				reminder.ID = uint(id)
			}
		}

		if title, ok := hit.Document["title"].(string); ok {
			reminder.Title = title
		}

		if content, ok := hit.Document["content"].(string); ok {
			reminder.Content = content
		}

		reminders = append(reminders, reminder)
	}

	return &models.SearchResponse{
		Results: reminders,
		Total:   result.Found,
		Query:   query,
	}, nil
}

// DeleteReminder removes a reminder from the search index
func (s *httpSearchService) DeleteReminder(id uint) error {
	req, err := http.NewRequest("DELETE", s.baseURL+"/collections/"+RemindersCollection+"/documents/"+strconv.Itoa(int(id)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete reminder from search index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete reminder: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateReminder updates a reminder in the search index
func (s *httpSearchService) UpdateReminder(reminder *models.Reminder) error {
	document := map[string]interface{}{
		"id":         strconv.Itoa(int(reminder.ID)),
		"title":      reminder.Title,
		"content":    reminder.Content,
		"created_at": reminder.CreatedAt.Unix(),
	}

	jsonData, err := json.Marshal(document)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", s.baseURL+"/collections/"+RemindersCollection+"/documents/"+strconv.Itoa(int(reminder.ID)), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update reminder in search index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update reminder: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
