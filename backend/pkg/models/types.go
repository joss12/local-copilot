package models

import "time"

// File represents a source code file in the workspace
type File struct {
	ID          int64     `json:"id"`
	Path        string    `json:"path"`
	Language    string    `json:"language"` // "javascript", "typescript", "go"
	ContentHash string    `json:"content_hash"`
	LastIndexed time.Time `json:"last_indexed"`
}

// Symbol represents a code symbol (function, variable, class, etc.)
type Symbol struct {
	ID        int64  `json:"id"`
	FileID    int64  `json:"file_id"`
	Type      string `json:"type"` // "function", "variable", "class", "import"
	Name      string `json:"name"`
	Signature string `json:"signature"` // Full signature for functions
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	Context   string `json:"context"` // Surrounding code snippet
}

// Pattern represents a searchable code pattern
type Pattern struct {
	ID          int64  `json:"id"`
	SymbolID    int64  `json:"symbol_id"`
	PatternType string `json:"pattern_type"` // "function_call", "variable_usage", etc.
	PatternText string `json:"pattern_text"` // The actual code snippet
}

// IndexRequest represents a request to index a workspace
type IndexRequest struct {
	WorkspacePath string `json:"workspace_path"`
}

// IndexResponse represents the result of an indexing operation
type IndexResponse struct {
	FilesProcessed int      `json:"files_processed"`
	SymbolsFound   int      `json:"symbols_found"`
	Errors         []string `json:"errors,omitempty"`
	Duration       string   `json:"duration"`
}

// SuggestionRequest represents a request for code suggestions
type SuggestionRequest struct {
	FilePath      string `json:"file_path"`
	Language      string `json:"language"`
	CurrentLine   int    `json:"current_line"`
	CurrentColumn int    `json:"current_column"`
	ContextBefore string `json:"context_before"` // Code before cursor
	ContextAfter  string `json:"context_after"`  // Code after cursor
	PartialSymbol string `json:"partial_symbol"` // What user is typing
}

// Suggestion represents a single code suggestion
type Suggestion struct {
	Text       string  `json:"text"`        // The suggested code
	Type       string  `json:"type"`        // "function", "variable", etc.
	Source     string  `json:"source"`      // Which file it came from
	Confidence float64 `json:"confidence"`  // 0.0 to 1.0
	LineNumber int     `json:"line_number"` // Where it's defined
}

// SuggestionResponse represents the response containing suggestions
type SuggestionResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Duration    string       `json:"duration"`
}
