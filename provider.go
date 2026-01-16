package main

// URLProvider defines the interface for URL operations
// Both App (local DB) and APIClient (remote API) implement this
type URLProvider interface {
	GetAllURLs() ([]CustomURL, error)
	CreateURL(originalURL string) (*CustomURL, error)
	DeleteURL(id uint) error
	SearchURLs(query string) ([]CustomURL, error)
}

// Ensure App implements URLProvider
var _ URLProvider = (*App)(nil)

// Ensure APIClient implements URLProvider
var _ URLProvider = (*APIClient)(nil)
