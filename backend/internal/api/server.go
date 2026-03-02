package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joss12/local-copilot/internal/indexer"
	"github.com/joss12/local-copilot/internal/matcher"
	"github.com/joss12/local-copilot/pkg/models"
)

type Server struct {
	indexer *indexer.Indexer
	matcher *matcher.Matcher
	port    string
}

// New API server
func NewServer(idx *indexer.Indexer, mtch *matcher.Matcher, port string) *Server {
	return &Server{
		indexer: idx,
		matcher: mtch,
		port:    port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("POST /index", s.handleIndex)
	mux.HandleFunc("POST /suggest", s.handleSuggest)

	//CORS middleware
	handler := s.corsMiddleware(mux)

	addr := ":" + s.port
	log.Printf("Server starting on http://localhost%s", addr)
	return http.ListenAndServe(addr, handler)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	start := time.Now()

	//Perform indexing
	response, err := s.indexer.IndexWorkspace()
	if err != nil {
		http.Error(w, fmt.Sprintf("Indexing failed: %v", err), http.StatusInternalServerError)
		return
	}

	response.Duration = time.Since(start).String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %V", err), http.StatusBadRequest)
		return
	}

	start := time.Now()

	//Get suggestions
	suggestions, err := s.matcher.GetSuggestions(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.SuggestionResponse{
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		//Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
