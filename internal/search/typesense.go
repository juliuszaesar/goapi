package search

import (
	"context"
	"encoding/json"
	"fmt"
	"goapi/internal/config"
	"goapi/internal/models"
	"log"
	"strconv"
	"time"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
)

const (
	RemindersCollection = "reminders"
)

// SearchService defines the interface for search operations
type SearchService interface {
	InitializeSchema() error
	IndexReminder(reminder *models.Reminder) error
	SearchReminders(query string, limit int) (*models.SearchResponse, error)
	DeleteReminder(id uint) error
	UpdateReminder(reminder *models.Reminder) error
}

// typesenseService implements SearchService
type typesenseService struct {
	client *typesense.Client
}

// NewSearchService creates a new search service
func NewSearchService(cfg *config.TypesenseConfig) SearchService {
	client := typesense.NewClient(
		typesense.WithServer(cfg.GetTypesenseURL()),
		typesense.WithAPIKey(cfg.APIKey),
		typesense.WithConnectionTimeout(30),
		typesense.WithCircuitBreakerMaxRequests(50),
		typesense.WithCircuitBreakerInterval(2),
		typesense.WithCircuitBreakerTimeout(10),
		typesense.WithHealthcheckInterval(30),
		typesense.WithNumRetries(3),
	)

	return &typesenseService{client: client}
}

// InitializeSchema creates the reminders collection schema
func (s *typesenseService) InitializeSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if collection exists
	_, err := s.client.Collection(RemindersCollection).Retrieve(ctx)
	if err == nil {
		// Collection exists, no need to create
		log.Printf("Typesense collection %s already exists", RemindersCollection)
		return nil
	}

	// Create collection schema
	defaultSortingField := "created_at"
	schema := &api.CollectionSchema{
		Name: RemindersCollection,
		Fields: []api.Field{
			{
				Name: "id",
				Type: "string",
			},
			{
				Name: "title",
				Type: "string",
			},
			{
				Name: "content",
				Type: "string",
			},
			{
				Name: "created_at",
				Type: "int64",
			},
		},
		DefaultSortingField: &defaultSortingField,
	}

	_, err = s.client.Collections().Create(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	log.Printf("Created Typesense collection: %s", RemindersCollection)
	return nil
}

// IndexReminder indexes a reminder in Typesense
func (s *typesenseService) IndexReminder(reminder *models.Reminder) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	document := map[string]interface{}{
		"id":         strconv.Itoa(int(reminder.ID)),
		"title":      reminder.Title,
		"content":    reminder.Content,
		"created_at": reminder.CreatedAt.Unix(),
	}

	_, err := s.client.Collection(RemindersCollection).Documents().Create(ctx, document)
	if err != nil {
		return fmt.Errorf("failed to index reminder: %w", err)
	}

	return nil
}

// SearchReminders searches for reminders using Typesense
func (s *typesenseService) SearchReminders(query string, limit int) (*models.SearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 10
	}

	queryBy := "title,content"
	sortBy := "created_at:desc"
	searchParams := &api.SearchCollectionParams{
		Q:       &query,
		QueryBy: &queryBy,
		PerPage: &limit,
		SortBy:  &sortBy,
	}

	result, err := s.client.Collection(RemindersCollection).Documents().Search(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search reminders: %w", err)
	}

	var reminders []models.ReminderResponse
	for _, hit := range *result.Hits {
		var reminder models.ReminderResponse
		
		// Parse the document
		docBytes, err := json.Marshal(hit.Document)
		if err != nil {
			continue
		}
		
		var doc map[string]interface{}
		if err := json.Unmarshal(docBytes, &doc); err != nil {
			continue
		}

		// Convert to ReminderResponse
		if idStr, ok := doc["id"].(string); ok {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				reminder.ID = uint(id)
			}
		}
		
		if title, ok := doc["title"].(string); ok {
			reminder.Title = title
		}
		
		if content, ok := doc["content"].(string); ok {
			reminder.Content = content
		}

		reminders = append(reminders, reminder)
	}

	total := 0
	if result.Found != nil {
		total = int(*result.Found)
	}

	return &models.SearchResponse{
		Results: reminders,
		Total:   total,
		Query:   query,
	}, nil
}

// DeleteReminder removes a reminder from the search index
func (s *typesenseService) DeleteReminder(id uint) error {
	_, err := s.client.Collection(RemindersCollection).Document(strconv.Itoa(int(id))).Delete(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete reminder from search index: %w", err)
	}
	return nil
}

// UpdateReminder updates a reminder in the search index
func (s *typesenseService) UpdateReminder(reminder *models.Reminder) error {
	document := map[string]interface{}{
		"id":         strconv.Itoa(int(reminder.ID)),
		"title":      reminder.Title,
		"content":    reminder.Content,
		"created_at": reminder.CreatedAt.Unix(),
	}

	_, err := s.client.Collection(RemindersCollection).Document(strconv.Itoa(int(reminder.ID))).Update(context.Background(), document)
	if err != nil {
		return fmt.Errorf("failed to update reminder in search index: %w", err)
	}

	return nil
}
