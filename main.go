package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Reminder struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ReminderService interface {
	CreateReminder(title, content string) (*Reminder, error)
	GetReminders() ([]Reminder, error)
}

type reminderService struct {
	db *gorm.DB
}

func NewReminderService(db *gorm.DB) ReminderService {
	return &reminderService{db: db}
}

func (s *reminderService) CreateReminder(title, content string) (*Reminder, error) {
	reminder := &Reminder{
		Title:   title,
		Content: content,
	}

	if err := s.db.Create(reminder).Error; err != nil {
		return nil, err
	}

	return reminder, nil
}

func (s *reminderService) GetReminders() ([]Reminder, error) {
	var reminders []Reminder
	if err := s.db.Find(&reminders).Error; err != nil {
		return nil, err
	}
	return reminders, nil
}

func main() {
	e := echo.New()

	dsn := "host=localhost user=goapi password=password dbname=goapi"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")

	db.AutoMigrate(&Reminder{})

	reminderService := NewReminderService(db)

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	e.POST("/reminders", func(c echo.Context) error {
		return createReminder(c, reminderService)
	})

	e.GET("/reminders", func(c echo.Context) error {
		return getReminders(c, reminderService)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func createReminder(c echo.Context, service ReminderService) error {
	title := c.FormValue("title")
	content := c.FormValue("content")

	reminder, err := service.CreateReminder(title, content)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, reminder)
}

func getReminders(c echo.Context, service ReminderService) error {
	reminders, err := service.GetReminders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, reminders)
}
