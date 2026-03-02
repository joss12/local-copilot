package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/joss12/local-copilot/internal/api"
	"github.com/joss12/local-copilot/internal/indexer"
	"github.com/joss12/local-copilot/internal/matcher"
)

func main() {
	workspacePath := flag.String("workspace", "", "Path to workspace to index")
	dbPath := flag.String("db", "", "Path to SQLite database")
	port := flag.String("port", "8089", "Server port")
	useLLM := flag.Bool("llm", false, "Enable LLM-powered suggestions (slower)")
	flag.Parse()

	//Set default database path if not provided
	if *dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get home directory:", err)
		}
		*dbPath = filepath.Join(homeDir, ".local-copilot", "copilot.db")
	}

	//Creates database directory if it doesn't exist

	dbDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal("Failed to create database directory:", err)
	}

	log.Printf("Using database: %s", *dbPath)

	//Initialize database
	db, err := indexer.NewDatabase(*dbPath)
	if err != nil {
		log.Fatal("Failed to Initialize database:", err)
	}
	defer db.Close()

	log.Println("Database Initialize successfully")

	//Set default workspace path if not provided
	if *workspacePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current directory:", err)
		}
		*workspacePath = cwd
	}

	log.Printf("Workspace path: %s", *workspacePath)

	//Create indexer
	idx := indexer.NewIndexer(db, *workspacePath)

	//create matcher
	// Create matcher
	mtch := matcher.NewMatcher(db, *useLLM)

	//Create and start API Server
	server := api.NewServer(idx, mtch, *port)

	log.Printf("Starting server on port %s...", *port)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
