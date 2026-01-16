package main

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDBTestApp creates an App instance with an in-memory SQLite database for testing
func setupDBTestApp(t *testing.T) *App {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&CustomURL{})

	config := Config{
		RedirectPrefixURL: "http://test.local/r/",
	}

	return NewApp(db, config)
}

func TestValidURL_DB(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Valid URLs
		{"valid http", "http://example.com", true},
		{"valid https", "https://example.com", true},
		{"valid https with path and query", "https://example.com/path?query=1", true},

		// Invalid URLs
		{"ftp scheme", "ftp://example.com", false},
		{"not a url", "not-a-url", false},
		{"empty string", "", false},
		{"http only", "http://", false},
		{"missing scheme", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validURL(tt.url)
			if result != tt.expected {
				t.Errorf("validURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestGenerateShortURL_DB(t *testing.T) {
	app := setupDBTestApp(t)

	t.Run("has correct prefix", func(t *testing.T) {
		shortURL := app.generateShortURL()
		if !strings.HasPrefix(shortURL, app.Config.RedirectPrefixURL) {
			t.Errorf("generateShortURL() = %q, want prefix %q", shortURL, app.Config.RedirectPrefixURL)
		}
	})

	t.Run("short code is exactly 10 characters", func(t *testing.T) {
		shortURL := app.generateShortURL()
		shortCode := strings.TrimPrefix(shortURL, app.Config.RedirectPrefixURL)
		if len(shortCode) != 10 {
			t.Errorf("generateShortURL() short code length = %d, want 10", len(shortCode))
		}
	})

	t.Run("contains only alphanumeric characters", func(t *testing.T) {
		shortURL := app.generateShortURL()
		shortCode := strings.TrimPrefix(shortURL, app.Config.RedirectPrefixURL)

		for _, char := range shortCode {
			isAlphanumeric := (char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9')
			if !isAlphanumeric {
				t.Errorf("generateShortURL() contains non-alphanumeric character: %c", char)
			}
		}
	})

	t.Run("generates unique URLs", func(t *testing.T) {
		generated := make(map[string]bool)
		for i := 0; i < 100; i++ {
			shortURL := app.generateShortURL()
			if generated[shortURL] {
				t.Errorf("generateShortURL() generated duplicate: %s", shortURL)
			}
			generated[shortURL] = true
		}
	})
}

func TestCreateURL(t *testing.T) {
	t.Run("creates new URL successfully", func(t *testing.T) {
		app := setupDBTestApp(t)

		url, err := app.CreateURL("https://example.com/new-page")
		if err != nil {
			t.Fatalf("CreateURL() error = %v, want nil", err)
		}
		if url == nil {
			t.Fatal("CreateURL() returned nil URL")
		}
		if url.OldURL != "https://example.com/new-page" {
			t.Errorf("CreateURL() OldURL = %q, want %q", url.OldURL, "https://example.com/new-page")
		}
		if url.ShortURL == "" {
			t.Error("CreateURL() ShortURL is empty")
		}
		if url.ID == 0 {
			t.Error("CreateURL() ID is 0, expected non-zero after save")
		}
	})

	t.Run("returns existing URL for duplicates", func(t *testing.T) {
		app := setupDBTestApp(t)

		// Create first URL
		first, err := app.CreateURL("https://example.com/duplicate")
		if err != nil {
			t.Fatalf("CreateURL() first call error = %v", err)
		}

		// Try to create the same URL again
		second, err := app.CreateURL("https://example.com/duplicate")
		if err != nil {
			t.Fatalf("CreateURL() second call error = %v", err)
		}

		if second.ID != first.ID {
			t.Errorf("CreateURL() for duplicate returned different ID: got %d, want %d", second.ID, first.ID)
		}
		if second.ShortURL != first.ShortURL {
			t.Errorf("CreateURL() for duplicate returned different ShortURL: got %q, want %q", second.ShortURL, first.ShortURL)
		}
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		app := setupDBTestApp(t)

		_, err := app.CreateURL("not-a-valid-url")
		if err == nil {
			t.Error("CreateURL() with invalid URL should return error")
		}
		if !strings.Contains(err.Error(), "invalid") {
			t.Errorf("CreateURL() error = %q, want error containing 'invalid'", err.Error())
		}
	})

	t.Run("returns error for empty URL", func(t *testing.T) {
		app := setupDBTestApp(t)

		_, err := app.CreateURL("")
		if err == nil {
			t.Error("CreateURL() with empty URL should return error")
		}
	})

	t.Run("returns error for ftp URL", func(t *testing.T) {
		app := setupDBTestApp(t)

		_, err := app.CreateURL("ftp://example.com")
		if err == nil {
			t.Error("CreateURL() with ftp URL should return error")
		}
	})
}

func TestDeleteURL(t *testing.T) {
	t.Run("deletes existing URL", func(t *testing.T) {
		app := setupDBTestApp(t)

		// Create a URL to delete
		url, err := app.CreateURL("https://example.com/to-delete")
		if err != nil {
			t.Fatalf("CreateURL() error = %v", err)
		}

		// Delete the URL
		err = app.DeleteURL(url.ID)
		if err != nil {
			t.Fatalf("DeleteURL() error = %v, want nil", err)
		}

		// Verify it was deleted
		var count int64
		app.DB.Model(&CustomURL{}).Where("id = ?", url.ID).Count(&count)
		if count != 0 {
			t.Error("DeleteURL() did not delete the URL from database")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		app := setupDBTestApp(t)

		err := app.DeleteURL(99999)
		if err == nil {
			t.Error("DeleteURL() with non-existent ID should return error")
		}
		if !strings.Contains(err.Error(), "99999") {
			t.Errorf("DeleteURL() error = %q, want error containing ID '99999'", err.Error())
		}
	})

	t.Run("returns error for ID 0", func(t *testing.T) {
		app := setupDBTestApp(t)

		err := app.DeleteURL(0)
		if err == nil {
			t.Error("DeleteURL() with ID 0 should return error")
		}
	})
}

func TestGetAllURLs(t *testing.T) {
	t.Run("returns empty slice when no URLs", func(t *testing.T) {
		app := setupDBTestApp(t)

		urls, err := app.GetAllURLs()
		if err != nil {
			t.Fatalf("GetAllURLs() error = %v, want nil", err)
		}
		if len(urls) != 0 {
			t.Errorf("GetAllURLs() returned %d URLs, want 0", len(urls))
		}
	})

	t.Run("returns all URLs ordered by created_at desc", func(t *testing.T) {
		app := setupDBTestApp(t)

		// Create multiple URLs
		app.CreateURL("https://example.com/first")
		app.CreateURL("https://example.com/second")
		app.CreateURL("https://example.com/third")

		urls, err := app.GetAllURLs()
		if err != nil {
			t.Fatalf("GetAllURLs() error = %v, want nil", err)
		}
		if len(urls) != 3 {
			t.Fatalf("GetAllURLs() returned %d URLs, want 3", len(urls))
		}

		// Verify order (most recent first)
		if urls[0].OldURL != "https://example.com/third" {
			t.Errorf("GetAllURLs()[0].OldURL = %q, want %q", urls[0].OldURL, "https://example.com/third")
		}
		if urls[1].OldURL != "https://example.com/second" {
			t.Errorf("GetAllURLs()[1].OldURL = %q, want %q", urls[1].OldURL, "https://example.com/second")
		}
		if urls[2].OldURL != "https://example.com/first" {
			t.Errorf("GetAllURLs()[2].OldURL = %q, want %q", urls[2].OldURL, "https://example.com/first")
		}
	})

	t.Run("returns URLs with all fields populated", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://example.com/test")

		urls, err := app.GetAllURLs()
		if err != nil {
			t.Fatalf("GetAllURLs() error = %v", err)
		}
		if len(urls) != 1 {
			t.Fatalf("GetAllURLs() returned %d URLs, want 1", len(urls))
		}

		url := urls[0]
		if url.ID == 0 {
			t.Error("GetAllURLs() URL has ID 0")
		}
		if url.OldURL == "" {
			t.Error("GetAllURLs() URL has empty OldURL")
		}
		if url.ShortURL == "" {
			t.Error("GetAllURLs() URL has empty ShortURL")
		}
		if url.CreatedAt.IsZero() {
			t.Error("GetAllURLs() URL has zero CreatedAt")
		}
	})
}

func TestSearchURLs(t *testing.T) {
	t.Run("finds matching URLs case-insensitive", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://example.com/test-page")
		app.CreateURL("https://github.com/project")
		app.CreateURL("https://EXAMPLE.org/other")

		// Search for "example" should match both example.com and EXAMPLE.org
		urls, err := app.SearchURLs("example")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v, want nil", err)
		}
		if len(urls) != 2 {
			t.Errorf("SearchURLs('example') returned %d URLs, want 2", len(urls))
		}
	})

	t.Run("finds URLs with partial match", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://example.com/hello-world")
		app.CreateURL("https://example.com/goodbye")

		urls, err := app.SearchURLs("hello")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v", err)
		}
		if len(urls) != 1 {
			t.Errorf("SearchURLs('hello') returned %d URLs, want 1", len(urls))
		}
		if urls[0].OldURL != "https://example.com/hello-world" {
			t.Errorf("SearchURLs('hello')[0].OldURL = %q, want %q", urls[0].OldURL, "https://example.com/hello-world")
		}
	})

	t.Run("returns empty slice for no matches", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://example.com/page")

		urls, err := app.SearchURLs("nonexistent")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v, want nil", err)
		}
		if len(urls) != 0 {
			t.Errorf("SearchURLs('nonexistent') returned %d URLs, want 0", len(urls))
		}
	})

	t.Run("returns empty slice when database is empty", func(t *testing.T) {
		app := setupDBTestApp(t)

		urls, err := app.SearchURLs("anything")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v, want nil", err)
		}
		if len(urls) != 0 {
			t.Errorf("SearchURLs() on empty DB returned %d URLs, want 0", len(urls))
		}
	})

	t.Run("results ordered by created_at desc", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://test.com/first")
		app.CreateURL("https://test.com/second")
		app.CreateURL("https://test.com/third")

		urls, err := app.SearchURLs("test.com")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v", err)
		}
		if len(urls) != 3 {
			t.Fatalf("SearchURLs('test.com') returned %d URLs, want 3", len(urls))
		}

		// Verify order (most recent first)
		if urls[0].OldURL != "https://test.com/third" {
			t.Errorf("SearchURLs()[0].OldURL = %q, want %q", urls[0].OldURL, "https://test.com/third")
		}
	})

	t.Run("handles special characters in search", func(t *testing.T) {
		app := setupDBTestApp(t)

		app.CreateURL("https://example.com/path?query=value")

		urls, err := app.SearchURLs("query=value")
		if err != nil {
			t.Fatalf("SearchURLs() error = %v", err)
		}
		if len(urls) != 1 {
			t.Errorf("SearchURLs('query=value') returned %d URLs, want 1", len(urls))
		}
	})
}
