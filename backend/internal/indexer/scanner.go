package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var SupportedExtensions = map[string]string{
	".js":  "javascript",
	".jsx": "javascript",
	".ts":  "typescript",
	"tsx":  "typescript",
	".go":  "go",
}

// FileScaner scans a workspace for source files
type FileScanner struct {
	workspacePath string
}

// create new file scanner
func NewFileScanner(workspacePath string) *FileScanner {
	return &FileScanner{
		workspacePath: workspacePath,
	}
}

// This walks the workspace and returns all supported source files
func (fs *FileScanner) ScanFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(fs.workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		///Skipping directories
		if info.IsDir() {
			// skipping common dir with unneeded index
			dirName := info.Name()
			if shouldSkipDirectory(dirName) {
				return filepath.SkipDir
			}
			return nil
		}

		//ckeck if files has supported extension
		ext := filepath.Ext(path)
		if _, supported := SupportedExtensions[ext]; supported {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return files, nil
}

// shouldSkipDirectory checks if a directory should be skipped during scanning
func shouldSkipDirectory(dirName string) bool {
	skipDirs := []string{
		"node_modules",
		".git",
		".vscode",
		".idea",
		"dist",
		"build",
		"coverage",
		".next",
		".nuxt",
		"vendor",
		"target",
		".data",
		"__pycache__",
	}

	for _, skip := range skipDirs {
		if strings.EqualFold(dirName, skip) {
			return true
		}
	}
	return false
}

// GetLanguageFromPath determines the language from a file path
func GetLanguageFromPath(filePath string) string {
	ext := filepath.Ext(filePath)
	if lang, ok := SupportedExtensions[ext]; ok {
		return lang
	}
	return "unknwon"
}
