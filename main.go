package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	serverMode := flag.Bool("server", false, "Run as API server instead of TUI")
	flag.Parse()

	config := LoadConfig()

	if *serverMode {
		runServer(config)
	} else {
		runTUI(config)
	}
}

// runServer starts the HTTP API server
func runServer(config Config) {
	db, err := connectToDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to %s database", config.DBEngine)

	app := NewApp(db, config)
	api := NewAPI(app)
	handler := api.SetupRoutes()

	// Get port from environment (Render sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = config.ServerPort
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting API server on port %s", port)
	log.Printf("Redirect prefix: %s", config.RedirectPrefixURL)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health        - Health check")
	log.Printf("  GET  /api/urls      - List all URLs")
	log.Printf("  POST /api/urls      - Create short URL")
	log.Printf("  DELETE /api/urls/ID - Delete URL")
	log.Printf("  GET  /api/search?q= - Search URLs")
	log.Printf("  GET  /r/{code}      - Redirect")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// runTUI starts the terminal user interface
func runTUI(config Config) {
	var provider URLProvider
	var isRemote bool

	if config.APIURL != "" {
		// Remote mode: connect to API
		client := NewAPIClient(config.APIURL)
		if err := client.CheckHealth(); err != nil {
			log.Fatalf("Failed to connect to API: %v", err)
		}
		provider = client
		isRemote = true
		log.Printf("Connected to remote API: %s", config.APIURL)
	} else {
		// Local mode: connect to database
		db, err := connectToDB(config)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		provider = NewApp(db, config)
		isRemote = false
		log.Printf("Connected to local %s database", config.DBEngine)
	}

	m := initialModel(provider, isRemote)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
