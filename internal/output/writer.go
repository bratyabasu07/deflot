package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	appCtx "github.com/bratyabasu07/deflot/internal/context"
	"github.com/bratyabasu07/deflot/internal/filters"
)

// Writer handles buffered output to files and stdout.
type Writer struct {
	appCtx *appCtx.AppContext

	// File handles
	mu              sync.Mutex
	mainFile        *os.File
	mainWriter      *bufio.Writer
	categoryFiles   map[string]*os.File
	categoryWriters map[string]*bufio.Writer
}

// New creates a new Writer instance.
func New(ctx *appCtx.AppContext) (*Writer, error) {
	w := &Writer{
		appCtx:          ctx,
		categoryFiles:   make(map[string]*os.File),
		categoryWriters: make(map[string]*bufio.Writer),
	}

	if ctx.OutputDir != "" {
		// Create output directories
		if err := os.MkdirAll(ctx.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}

		sensitiveDir := filepath.Join(ctx.OutputDir, "sensitiveurls")
		if err := os.MkdirAll(sensitiveDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create sensitive urls directory: %w", err)
		}

		// Create main output file
		mainPath := filepath.Join(ctx.OutputDir, "wayback_urls.txt")
		f, err := os.Create(mainPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create main output file: %w", err)
		}
		w.mainFile = f
		w.mainWriter = bufio.NewWriter(f)
	}

	return w, nil
}

// Write outputs a record to the appropriate file(s).
func (w *Writer) Write(record appCtx.ScanRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Format output line
	var line string
	if w.appCtx.JSON {
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		line = string(data)
	} else {
		line = record.URL
	}

	// Write to stdout if enabled
	if w.appCtx.Stdout {
		fmt.Println(line)
	}

	// Write to main file
	if w.mainWriter != nil {
		w.mainWriter.WriteString(line + "\n")
	}

	// Write to category file if applicable
	if record.Category != "" && record.Category != "none" && w.appCtx.OutputDir != "" {
		if err := w.writeCategoryFile(record.Category, line); err != nil {
			return err
		}
	}

	return nil
}

// writeCategoryFile writes to the appropriate category file.
func (w *Writer) writeCategoryFile(category, line string) error {
	writer, ok := w.categoryWriters[category]
	if !ok {
		// Create the file for this category
		filename := getCategoryFilename(category)
		sensitiveDir := filepath.Join(w.appCtx.OutputDir, "sensitiveurls")
		filePath := filepath.Join(sensitiveDir, filename)

		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create category file %s: %w", filename, err)
		}
		w.categoryFiles[category] = f
		writer = bufio.NewWriter(f)
		w.categoryWriters[category] = writer
	}

	writer.WriteString(line + "\n")
	return nil
}

// getCategoryFilename maps category to filename.
func getCategoryFilename(category string) string {
	switch category {
	case filters.CatSecret:
		return "secret_urls.txt"
	case filters.CatConfig:
		return "config_urls.txt"
	case filters.CatBackup:
		return "backup_exposure_urls.txt"
	case filters.CatDatabase:
		return "database_backup_urls.txt"
	case filters.CatAPI:
		return "api_specs_urls.txt"
	case filters.CatParam:
		return "parameter_urls.txt"
	case filters.CatJS:
		return "js_urls.txt"
	case filters.CatPDF:
		return "pdf_urls.txt"
	case filters.CatLog:
		return "log_urls.txt"
	case filters.CatVCS:
		return "vcs_exposure_urls.txt"
	case filters.CatCloud:
		return "cloud_urls.txt"
	case filters.CatArchive:
		return "archive_urls.txt"
	case filters.CatDoc:
		return "doc_urls.txt"
	case filters.CatSheet:
		return "sheet_urls.txt"
	default:
		return "other_urls.txt"
	}
}

// Close flushes all buffers and closes all files.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush and close main file
	if w.mainWriter != nil {
		w.mainWriter.Flush()
	}
	if w.mainFile != nil {
		w.mainFile.Close()
	}

	// Flush and close category files
	for cat, writer := range w.categoryWriters {
		writer.Flush()
		if f, ok := w.categoryFiles[cat]; ok {
			f.Close()
		}
	}

	return nil
}
