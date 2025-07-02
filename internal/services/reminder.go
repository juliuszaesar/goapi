package services

import (
	"goapi/internal/models"
	"gorm.io/gorm"
)

// ReminderService defines the interface for reminder operations
type ReminderService interface {
	CreateReminder(req *models.ReminderRequest) (*models.Reminder, error)
	GetReminders() ([]models.Reminder, error)
	GetReminderByID(id uint) (*models.Reminder, error)
	UpdateReminder(id uint, req *models.ReminderRequest) (*models.Reminder, error)
	DeleteReminder(id uint) error
}

// reminderService implements ReminderService
type reminderService struct {
	db *gorm.DB
}

// NewReminderService creates a new reminder service
func NewReminderService(db *gorm.DB) ReminderService {
	return &reminderService{db: db}
}

// CreateReminder creates a new reminder
func (s *reminderService) CreateReminder(req *models.ReminderRequest) (*models.Reminder, error) {
	reminder := &models.Reminder{
		Title:   req.Title,
		Content: req.Content,
	}

	if err := s.db.Create(reminder).Error; err != nil {
		return nil, err
	}

	return reminder, nil
}

// GetReminders retrieves all reminders
func (s *reminderService) GetReminders() ([]models.Reminder, error) {
	var reminders []models.Reminder
	if err := s.db.Order("created_at DESC").Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

// GetReminderByID retrieves a reminder by ID
func (s *reminderService) GetReminderByID(id uint) (*models.Reminder, error) {
	var reminder models.Reminder
	if err := s.db.First(&reminder, id).Error; err != nil {
		return nil, err
	}
	return &reminder, nil
}

// UpdateReminder updates an existing reminder
func (s *reminderService) UpdateReminder(id uint, req *models.ReminderRequest) (*models.Reminder, error) {
	var reminder models.Reminder
	if err := s.db.First(&reminder, id).Error; err != nil {
		return nil, err
	}

	reminder.Title = req.Title
	reminder.Content = req.Content

	if err := s.db.Save(&reminder).Error; err != nil {
		return nil, err
	}

	return &reminder, nil
}

// DeleteReminder deletes a reminder by ID
func (s *reminderService) DeleteReminder(id uint) error {
	return s.db.Delete(&models.Reminder{}, id).Error
}
