package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Globals for URL generation
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var maxShortSize = 10

// CustomURL provides the model to interact with our DB
type CustomURL struct {
	ID        uint   `gorm:"primaryKey"`
	OldURL    string `gorm:"not null;uniqueIndex;size:255"`
	ShortURL  string `gorm:"not null;uniqueIndex;size:255"`
	Visits    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// App holds application dependencies for testability
type App struct {
	DB     *gorm.DB
	Config Config
}

// NewApp creates a new App instance
func NewApp(db *gorm.DB, config Config) *App {
	return &App{
		DB:     db,
		Config: config,
	}
}

// connectToDB establishes a database connection based on the configuration
// Supports postgres, mysql, and sqlite engines
func connectToDB(config Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch config.DBEngine {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.DBHost, config.DBUsername, config.DBPassword, config.DBName, config.DBPort)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.DBUsername, config.DBPassword, config.DBHost, config.DBPort, config.DBName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(config.DBName), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database engine: %s", config.DBEngine)
	}

	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Auto migrate the schema (safe - only creates/updates, never drops)
	err = db.AutoMigrate(&CustomURL{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// GetAllURLs returns all URLs ordered by created_at descending
func (app *App) GetAllURLs() ([]CustomURL, error) {
	var urls []CustomURL
	result := app.DB.Order("created_at desc").Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}
	return urls, nil
}

// CreateURL validates the URL, checks for duplicates, and creates a new short URL
// Returns the existing URL if a duplicate is found
func (app *App) CreateURL(originalURL string) (*CustomURL, error) {
	// Validate the URL
	if !validURL(originalURL) {
		return nil, fmt.Errorf("invalid URL provided")
	}

	// Check for duplicate
	var existing CustomURL
	result := app.DB.Where("old_url = ?", originalURL).First(&existing)
	if result.Error == nil {
		// URL already exists, return the existing record
		return &existing, nil
	}

	// Create new short URL
	shortURL := CustomURL{
		OldURL:   originalURL,
		ShortURL: app.generateShortURL(),
		Visits:   0,
	}

	result = app.DB.Create(&shortURL)
	if result.Error != nil {
		return nil, result.Error
	}

	return &shortURL, nil
}

// DeleteURL deletes a URL by its ID
func (app *App) DeleteURL(id uint) error {
	result := app.DB.Delete(&CustomURL{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no URL found with ID %d", id)
	}
	return nil
}

// SearchURLs searches for URLs containing the query string in the OldURL field
func (app *App) SearchURLs(query string) ([]CustomURL, error) {
	var urls []CustomURL
	searchPattern := "%" + strings.ToLower(query) + "%"
	result := app.DB.Where("LOWER(old_url) LIKE ?", searchPattern).Order("created_at desc").Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}
	return urls, nil
}

// validURL validates URLs using net/url package
// Returns true if the URL has http or https scheme and a valid host
func validURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	// Must have http or https scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	// Must have a host
	if parsed.Host == "" {
		return false
	}
	return true
}

// generateShortURL generates a random 10-character alphanumeric string
// prefixed with the configured redirect URL
func (app *App) generateShortURL() string {
	b := make([]rune, maxShortSize)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return app.Config.RedirectPrefixURL + string(b)
}
