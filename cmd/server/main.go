package main

import (
	"fmt"
	"goapi/internal/config"
	"goapi/internal/handlers"
	"goapi/internal/models"
	"goapi/internal/search"
	"goapi/internal/services"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database")

	// Auto-migrate database schema
	if err := db.AutoMigrate(&models.Reminder{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")

	// Initialize search service
	searchService := search.NewHTTPSearchService(&cfg.Typesense)
	if err := searchService.InitializeSchema(); err != nil {
		log.Printf("Warning: Failed to initialize search schema: %v", err)
		log.Println("Search functionality may not work properly")
	} else {
		log.Println("Search service initialized")
	}

	// Initialize services
	reminderService := services.NewReminderService(db)

	// Initialize handlers
	reminderHandler := handlers.NewReminderHandler(reminderService, searchService)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	setupRoutes(e, reminderHandler)

	// Start server
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", serverAddr)
	e.Logger.Fatal(e.Start(serverAddr))
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func setupRoutes(e *echo.Echo, reminderHandler *handlers.ReminderHandler) {
	// Static files
	e.Static("/", "web/static")

	// API routes
	api := e.Group("/api")
	{
		api.POST("/reminders", reminderHandler.CreateReminder)
		api.GET("/reminders", reminderHandler.GetReminders)
		api.GET("/reminders/:id", reminderHandler.GetReminder)
		api.PUT("/reminders/:id", reminderHandler.UpdateReminder)
		api.DELETE("/reminders/:id", reminderHandler.DeleteReminder)
		api.GET("/search", reminderHandler.SearchReminders)
	}

	// Legacy routes for backward compatibility
	e.POST("/reminders", reminderHandler.CreateReminder)
	e.GET("/reminders", reminderHandler.GetReminders)
	e.DELETE("/reminders/:id", reminderHandler.DeleteReminder)
	e.GET("/search", reminderHandler.SearchReminders)
}
