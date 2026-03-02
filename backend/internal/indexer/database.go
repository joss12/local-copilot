package indexer

import (
	"database/sql"
	"fmt"
	"github.com/joss12/local-copilot/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	//Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	database := &Database{db: db}

	//Intialize schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("Failed to initalize schema: %w", err)
	}

	return database, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		language TEXT NOT NULL,
		content_hash TEXT NOT NULL,
		last_indexed DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS symbols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		signature TEXT,
		line_start INTEGER,
		line_end INTEGER,
		context TEXT,
		FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol_id INTEGER NOT NULL,
		pattern_type TEXT,
		pattern_text TEXT,
		FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
	CREATE INDEX IF NOT EXISTS idx_symbols_type ON symbols(type);
	CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
	`

	_, err := d.db.Exec(schema)
	return err
}

// InsertFile inserts or updates a file record
func (d *Database) InsertFile(file *models.File) (int64, error) {
	query := `
	INSERT INTO files (path, language, content_hash, last_indexed)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			language = excluded.language,
			content_hash = excluded.content_hash,
			last_indexed = excluded.last_indexed
		RETURNING id
	`
	var id int64
	err := d.db.QueryRow(query, file.Path, file.Language, file.ContentHash, time.Now()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Failed to insert file: %w", err)
	}
	return id, nil
}

func (d *Database) InsertSymbol(symbol *models.Symbol) (int64, error) {
	query := `
	INSERT INTO symbols (file_id, type, name, signature, line_start, line_end, context)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := d.db.Exec(query,
		symbol.FileID,
		symbol.Type,
		symbol.Name,
		symbol.Signature,
		symbol.LineStart,
		symbol.LineEnd,
		symbol.Context,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert symbol: %w", err)
	}
	return result.LastInsertId()
}

func (d *Database) InsertPattern(pattern *models.Pattern) (int64, error) {
	query := `
	INSERT INTO patterns (symbol_id, pattern_type, pattern_text)
		VALUES (?, ?, ?)
	`

	result, err := d.db.Exec(query,
		pattern.SymbolID,
		pattern.PatternType,
		pattern.PatternText,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert pattern: %w", err)
	}
	return result.LastInsertId()
}

func (d *Database) DeleteSymbolsByFileID(fileID int64) error {
	_, err := d.db.Exec("DELETE FROM symbols WHERE file_id = ?", fileID)
	return err
}

// SearchSymbols searches for symbols by name (for suggestions)
func (d *Database) SearchSymbols(partialName string, limit int) ([]models.Symbol, error) {
	query := `
	SELECT id, file_id, type, name, signature, line_start, line_end, context
		FROM symbols
		WHERE name LIKE ?
		ORDER BY name
		LIMIT ?
	`

	rows, err := d.db.Query(query, partialName+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []models.Symbol
	for rows.Next() {
		var s models.Symbol
		err := rows.Scan(&s.ID, &s.FileID, &s.Type, &s.Name, &s.Signature, &s.LineStart, &s.LineEnd, &s.Context)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}
	return symbols, rows.Err()
}

// GetFileByID retrieves a file by ID
func (d *Database) GetFileByID(id int64) (*models.File, error) {
	query := `SELECT id, path, language, content_hash, last_indexed FROM files WHERE id = ?`

	var f models.File
	err := d.db.QueryRow(query, id).Scan(&f.ID, &f.Path, &f.Language, &f.ContentHash, &f.LastIndexed)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
