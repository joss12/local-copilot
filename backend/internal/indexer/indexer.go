package indexer

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"os"
	"strings"

	"github.com/joss12/local-copilot/pkg/models"
)

// Indexer handles code indexing operations
type Indexer struct {
	db      *Database
	scanner *FileScanner
}

// NewIndexer creates a new indexer
func NewIndexer(db *Database, workspacePath string) *Indexer {
	return &Indexer{
		db:      db,
		scanner: NewFileScanner(workspacePath),
	}
}

// IndexWorkspace indexes all files in the workspace
func (idx *Indexer) IndexWorkspace() (*models.IndexResponse, error) {
	files, err := idx.scanner.ScanFiles()
	if err != nil {
		return nil, err
	}

	response := &models.IndexResponse{
		FilesProcessed: 0,
		SymbolsFound:   0,
		Errors:         []string{},
	}

	for _, filePath := range files {
		symbols, err := idx.indexFile(filePath)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("%s: %v", filePath, err))
			continue
		}

		response.FilesProcessed++
		response.SymbolsFound += len(symbols)
	}

	return response, nil
}

// indexFile indexes a single file
func (idx *Indexer) indexFile(filePath string) ([]models.Symbol, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate content hash
	hash := sha256.Sum256(content)
	contentHash := fmt.Sprintf("%x", hash)

	// Determine language
	language := GetLanguageFromPath(filePath)

	// Insert/update file record
	file := &models.File{
		Path:        filePath,
		Language:    language,
		ContentHash: contentHash,
	}

	fileID, err := idx.db.InsertFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to insert file: %w", err)
	}

	// Delete old symbols for this file
	if err := idx.db.DeleteSymbolsByFileID(fileID); err != nil {
		return nil, fmt.Errorf("failed to delete old symbols: %w", err)
	}

	// Parse based on language
	var symbols []models.Symbol
	switch language {
	case "go":
		symbols, err = idx.parseGoFile(filePath, fileID, content)
	case "javascript", "typescript":
		// We'll implement this in the next step
		symbols = []models.Symbol{}
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	if err != nil {
		return nil, err
	}

	// Insert symbols into database
	for _, symbol := range symbols {
		_, err := idx.db.InsertSymbol(&symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to insert symbol: %w", err)
		}
	}

	return symbols, nil
}

// parseGoFile parses a Go source file and extracts symbols
func (idx *Indexer) parseGoFile(filePath string, fileID int64, content []byte) ([]models.Symbol, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	var symbols []models.Symbol

	// Extract top-level declarations
	ast.Inspect(file, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			// Extract function
			symbol := extractGoFunction(decl, fset, fileID, content)
			symbols = append(symbols, symbol)

		case *ast.GenDecl:
			// Extract variables, constants, types
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					// Variable or constant
					for _, name := range s.Names {
						symbol := extractGoVariable(name, s, fset, fileID)
						symbols = append(symbols, symbol)
					}
				case *ast.TypeSpec:
					// Type declaration
					symbol := extractGoType(s, fset, fileID)
					symbols = append(symbols, symbol)
				}
			}
		}
		return true
	})

	return symbols, nil
}

// extractGoFunction extracts a function symbol
func extractGoFunction(decl *ast.FuncDecl, fset *token.FileSet, fileID int64, content []byte) models.Symbol {
	start := fset.Position(decl.Pos())
	end := fset.Position(decl.End())

	// Build function signature
	var signature strings.Builder
	signature.WriteString("func ")
	if decl.Recv != nil {
		signature.WriteString("(")
		// Method receiver
		for i, field := range decl.Recv.List {
			if i > 0 {
				signature.WriteString(", ")
			}
			if len(field.Names) > 0 {
				signature.WriteString(field.Names[0].Name)
				signature.WriteString(" ")
			}
			signature.WriteString(formatType(field.Type))
		}
		signature.WriteString(") ")
	}
	signature.WriteString(decl.Name.Name)
	signature.WriteString("(")
	if decl.Type.Params != nil {
		for i, field := range decl.Type.Params.List {
			if i > 0 {
				signature.WriteString(", ")
			}
			// Parameter names
			for j, name := range field.Names {
				if j > 0 {
					signature.WriteString(", ")
				}
				signature.WriteString(name.Name)
			}
			if len(field.Names) > 0 {
				signature.WriteString(" ")
			}
			signature.WriteString(formatType(field.Type))
		}
	}
	signature.WriteString(")")

	// Return type
	if decl.Type.Results != nil {
		results := decl.Type.Results.List
		if len(results) > 0 {
			signature.WriteString(" ")
			if len(results) > 1 || len(results[0].Names) > 1 {
				signature.WriteString("(")
			}
			for i, field := range results {
				if i > 0 {
					signature.WriteString(", ")
				}
				signature.WriteString(formatType(field.Type))
			}
			if len(results) > 1 || len(results[0].Names) > 1 {
				signature.WriteString(")")
			}
		}
	}

	// Extract context (surrounding code)
	context := extractContext(content, start.Offset, end.Offset)

	return models.Symbol{
		FileID:    fileID,
		Type:      "function",
		Name:      decl.Name.Name,
		Signature: signature.String(),
		LineStart: start.Line,
		LineEnd:   end.Line,
		Context:   context,
	}
}

// extractGoVariable extracts a variable/constant symbol
func extractGoVariable(name *ast.Ident, spec *ast.ValueSpec, fset *token.FileSet, fileID int64) models.Symbol {
	pos := fset.Position(name.Pos())

	varType := "var"
	if spec.Type != nil {
		varType = fmt.Sprintf("%s", spec.Type)
	}

	return models.Symbol{
		FileID:    fileID,
		Type:      "variable",
		Name:      name.Name,
		Signature: varType,
		LineStart: pos.Line,
		LineEnd:   pos.Line,
		Context:   "",
	}
}

// extractGoType extracts a type symbol
func extractGoType(spec *ast.TypeSpec, fset *token.FileSet, fileID int64) models.Symbol {
	pos := fset.Position(spec.Pos())

	return models.Symbol{
		FileID:    fileID,
		Type:      "type",
		Name:      spec.Name.Name,
		Signature: fmt.Sprintf("type %s", spec.Name.Name),
		LineStart: pos.Line,
		LineEnd:   pos.Line,
		Context:   "",
	}
}

// formatGoField formats a function parameter field
func formatGoField(field *ast.Field) string {
	var names []string
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	typeStr := formatGoFieldType(field)
	if len(names) > 0 {
		return strings.Join(names, ", ") + " " + typeStr
	}
	return typeStr
}

// formatGoFieldType formats the type of a field
func formatGoFieldType(field *ast.Field) string {
	return fmt.Sprintf("%s", field.Type)
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name

	case *ast.StarExpr:
		return "*" + formatType(t.X)

	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatType(t.Elt)
		}
		return "[" + formatExpr(t.Len) + "]" + formatType(t.Elt)

	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)

	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return "interface{}"
		}
		return "interface{...}"

	case *ast.StructType:
		return "struct{...}"

	case *ast.FuncType:
		return "func(...)"

	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name

	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan <- " + formatType(t.Value)
		case ast.RECV:
			return "<- chan" + formatType(t.Value)
		default:
			return "chan " + formatType(t.Value)
		}

	case *ast.Ellipsis:
		return "..." + formatType(t.Elt)

	default:
		return "unknown"
	}
}

func formatExpr(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Value
	}
	return ""
}

// extractContext extracts surrounding code as context
func extractContext(content []byte, start, end int) string {
	if start < 0 || end > len(content) || start >= end {
		return ""
	}
	context := string(content[start:end])
	if len(context) > 500 {
		context = context[:500] + "..."
	}
	return context
}
