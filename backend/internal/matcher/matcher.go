package matcher

import (
	"strings"

	"github.com/joss12/local-copilot/internal/indexer"
	"github.com/joss12/local-copilot/internal/llm"
	"github.com/joss12/local-copilot/pkg/models"
)

type Matcher struct {
	db        *indexer.Database
	llmClient *llm.OllamaClient
	useLLM    bool
}

// Create a new matcher
func NewMatcher(db *indexer.Database, useLLM bool) *Matcher {
	var client *llm.OllamaClient
	if useLLM {
		client = llm.NewOllamaClient("", "")
	}

	return &Matcher{
		db:        db,
		llmClient: client,
		useLLM:    useLLM,
	}
}

// GetSuggestions returns code suggestions based on the current context
func (m *Matcher) GetSuggestions(req *models.SuggestionRequest) ([]models.Suggestion, error) {
	var allSuggestions []models.Suggestion

	// Strategy 1: Pattern matching (fast)
	if req.PartialSymbol != "" {
		symbols, err := m.db.SearchSymbols(req.PartialSymbol, 10)
		if err == nil {
			for _, symbol := range symbols {
				file, err := m.db.GetFileByID(symbol.FileID)
				if err != nil {
					continue
				}

				suggestion := models.Suggestion{
					Text:       m.formatSuggestion(&symbol),
					Type:       symbol.Type,
					Source:     file.Path,
					Confidence: m.calculateConfidence(&symbol, req),
					LineNumber: symbol.LineStart,
				}
				allSuggestions = append(allSuggestions, suggestion)
			}
		}
	}

	// Strategy 2: LLM completion (slower but smarter)
	if m.useLLM && m.llmClient != nil && m.llmClient.IsAvailable() {
		llmCompletion, err := m.llmClient.GenerateCompletion(
			req.Language,
			req.ContextBefore,
			req.ContextAfter,
		)

		if err == nil && llmCompletion != "" {
			// Add LLM suggestion with high confidence
			llmSuggestion := models.Suggestion{
				Text:       llmCompletion,
				Type:       "llm",
				Source:     "DeepSeek Coder",
				Confidence: 0.95,
				LineNumber: 0,
			}
			allSuggestions = append(allSuggestions, llmSuggestion)
		}
	}

	// Sort by confidence (highest first)
	m.sortByConfidence(allSuggestions)

	return allSuggestions, nil
}
func (m *Matcher) formatSuggestion(symbol *models.Symbol) string {
	switch symbol.Type {
	case "function":
		if symbol.Signature != "" {
			return symbol.Signature
		}
		return symbol.Name + "()"
	case "variable":
		return symbol.Name
	case "type":
		return symbol.Name
	default:
		return symbol.Name
	}
}

func (m *Matcher) calculateConfidence(symbol *models.Symbol, req *models.SuggestionRequest) float64 {
	confidence := 0.5

	if strings.HasPrefix(symbol.Name, req.PartialSymbol) {
		confidence += 0.3
	}

	file, err := m.db.GetFileByID(symbol.FileID)
	if err == nil && file.Path == req.FilePath {
		confidence += 0.2
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// sortByConfidence sorts suggestions by confidence (descending)
func (m *Matcher) sortByConfidence(suggestions []models.Suggestion) {

	n := len(suggestions)
	if n <= 1 {
		return // Nothing to sort
	}

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if suggestions[j].Confidence < suggestions[j+1].Confidence {
				suggestions[j], suggestions[j+1] = suggestions[j+1], suggestions[j]
			}
		}
	}
}
