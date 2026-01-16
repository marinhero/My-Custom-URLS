package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// APIResponse is the standard response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateURLRequest is the request body for creating a URL
type CreateURLRequest struct {
	URL string `json:"url"`
}

// URLResponse is the response for a single URL
type URLResponse struct {
	ID        uint   `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
	Visits    int    `json:"visits"`
	CreatedAt string `json:"created_at"`
}

// API holds the application dependencies for the API handlers
type API struct {
	app *App
}

// NewAPI creates a new API instance
func NewAPI(app *App) *API {
	return &API{app: app}
}

// SetupRoutes configures the HTTP routes
func (api *API) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/urls", api.handleURLs)
	mux.HandleFunc("/api/urls/", api.handleURLByID)
	mux.HandleFunc("/api/search", api.handleSearch)
	mux.HandleFunc("/health", api.handleHealth)

	// Redirect handler (for short URLs)
	mux.HandleFunc("/r/", api.handleRedirect)

	// CORS middleware
	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth returns server health status
func (api *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: "ok"})
}

// handleURLs handles GET (list) and POST (create) for /api/urls
func (api *API) handleURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		api.listURLs(w, r)
	case http.MethodPost:
		api.createURL(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "method not allowed"})
	}
}

// listURLs returns all URLs
func (api *API) listURLs(w http.ResponseWriter, r *http.Request) {
	urls, err := api.app.GetAllURLs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	response := make([]URLResponse, len(urls))
	for i, u := range urls {
		response[i] = toURLResponse(u)
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: response})
}

// createURL creates a new short URL
func (api *API) createURL(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "invalid request body"})
		return
	}

	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "url is required"})
		return
	}

	url, err := api.app.CreateURL(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: toURLResponse(*url)})
}

// handleURLByID handles DELETE for /api/urls/{id}
func (api *API) handleURLByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/urls/")
	id, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "invalid id"})
		return
	}

	switch r.Method {
	case http.MethodDelete:
		api.deleteURL(w, uint(id))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "method not allowed"})
	}
}

// deleteURL deletes a URL by ID
func (api *API) deleteURL(w http.ResponseWriter, id uint) {
	err := api.app.DeleteURL(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: "deleted"})
}

// handleSearch handles GET /api/search?q=query
func (api *API) handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "query parameter 'q' is required"})
		return
	}

	urls, err := api.app.SearchURLs(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	response := make([]URLResponse, len(urls))
	for i, u := range urls {
		response[i] = toURLResponse(u)
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: response})
}

// handleRedirect handles the short URL redirect
func (api *API) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	// Find URL by short code
	var url CustomURL
	fullShortURL := api.app.Config.RedirectPrefixURL + shortCode
	result := api.app.DB.Where("short_url = ?", fullShortURL).First(&url)
	if result.Error != nil {
		http.NotFound(w, r)
		return
	}

	// Increment visit count
	api.app.DB.Model(&url).Update("visits", url.Visits+1)
	log.Printf("[+] Redirect: %s -> %s (visits: %d)", shortCode, url.OldURL, url.Visits+1)

	http.Redirect(w, r, url.OldURL, http.StatusFound)
}

// toURLResponse converts a CustomURL to URLResponse
func toURLResponse(u CustomURL) URLResponse {
	shortCode := u.ShortURL
	if idx := strings.LastIndex(shortCode, "/"); idx != -1 {
		shortCode = shortCode[idx+1:]
	}

	return URLResponse{
		ID:          u.ID,
		OriginalURL: u.OldURL,
		ShortURL:    u.ShortURL,
		ShortCode:   shortCode,
		Visits:      u.Visits,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
