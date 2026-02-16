package handlers

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
)

// OutputsHandler handles outputs endpoints
type OutputsHandler struct {
	cfg *config.Config
}

// NewOutputsHandler creates a new outputs handler
func NewOutputsHandler(cfg *config.Config) *OutputsHandler {
	return &OutputsHandler{cfg: cfg}
}

// OutputItem represents a file or directory in outputs
type OutputItem struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Path      string  `json:"path"`
	FileCount int     `json:"file_count,omitempty"`
	SizeKB    float64 `json:"size_kb,omitempty"`
	Modified  string  `json:"modified,omitempty"`
	Category  string  `json:"category,omitempty"`
}

// ListOutputs handles GET /outputs
func (h *OutputsHandler) ListOutputs(w http.ResponseWriter, r *http.Request) {
	basePath := h.cfg.OutputsDir

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Outputs directory not found",
			"path":  basePath,
		})
		return
	}

	contents := h.scanDirectory(basePath, basePath)

	// Calculate totals
	var totalSizeKB float64
	var dirs, files int
	for _, item := range contents {
		if item.Type == "directory" {
			dirs++
		} else {
			files++
			totalSizeKB += item.SizeKB
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path":          basePath,
		"total_items":   len(contents),
		"directories":   dirs,
		"files":         files,
		"total_size_kb": totalSizeKB,
		"contents":      contents,
		"note":          "Use GET /outputs/{ticker} to see detailed contents of a specific ticker directory",
	})
}

// ListTickerOutputs handles GET /outputs/{ticker}
func (h *OutputsHandler) ListTickerOutputs(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	ticker = strings.ToLower(ticker)

	tickerPath := filepath.Join(h.cfg.OutputsDir, ticker)

	if _, err := os.Stat(tickerPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "No outputs found for ticker '" + ticker + "'",
		})
		return
	}

	// Scan recursively
	files := h.scanRecursive(tickerPath, tickerPath)

	// Calculate total size
	var totalSizeKB float64
	for _, f := range files {
		totalSizeKB += f.SizeKB
	}

	// Group by category
	categories := make(map[string][]OutputItem)
	categoryList := []string{}

	for _, f := range files {
		cat := f.Category
		if _, exists := categories[cat]; !exists {
			categoryList = append(categoryList, cat)
		}
		categories[cat] = append(categories[cat], f)
	}

	sort.Strings(categoryList)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ticker":            strings.ToUpper(ticker),
		"path":              tickerPath,
		"total_files":       len(files),
		"total_size_kb":     totalSizeKB,
		"categories":        categoryList,
		"files_by_category": categories,
		"all_files":         files,
	})
}

// scanDirectory scans a directory and returns items
func (h *OutputsHandler) scanDirectory(path, relativeTo string) []OutputItem {
	items := []OutputItem{}

	entries, err := os.ReadDir(path)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		relPath, _ := filepath.Rel(relativeTo, fullPath)

		if entry.IsDir() {
			// Count files in directory
			fileCount := 0
			filepath.WalkDir(fullPath, func(p string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					fileCount++
				}
				return nil
			})

			items = append(items, OutputItem{
				Name:      entry.Name(),
				Type:      "directory",
				Path:      relPath,
				FileCount: fileCount,
			})
		} else {
			info, err := entry.Info()
			var sizeKB float64
			var modified string

			if err == nil {
				sizeKB = float64(info.Size()) / 1024.0
				modified = info.ModTime().Format("2006-01-02 15:04:05")
			}

			items = append(items, OutputItem{
				Name:     entry.Name(),
				Type:     "file",
				Path:     relPath,
				SizeKB:   sizeKB,
				Modified: modified,
			})
		}
	}

	// Sort: directories first, then by name
	sort.Slice(items, func(i, j int) bool {
		if (items[i].Type == "file") != (items[j].Type == "file") {
			return items[i].Type == "directory"
		}
		return items[i].Name < items[j].Name
	})

	return items
}

// scanRecursive scans a directory recursively and returns all files
func (h *OutputsHandler) scanRecursive(path, relativeTo string) []OutputItem {
	files := []OutputItem{}

	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(relativeTo, p)
		info, _ := d.Info()

		var sizeKB float64
		var modified string
		if info != nil {
			sizeKB = float64(info.Size()) / 1024.0
			modified = info.ModTime().Format("2006-01-02 15:04:05")
		} else {
			modified = time.Now().Format("2006-01-02 15:04:05")
		}

		category := filepath.Base(filepath.Dir(p))

		files = append(files, OutputItem{
			Name:     d.Name(),
			Path:     relPath,
			SizeKB:   sizeKB,
			Modified: modified,
			Category: category,
		})

		return nil
	})

	// Sort by path
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files
}
