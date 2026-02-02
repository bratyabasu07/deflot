package jssecrethunter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scanner wraps the JSSecretHunter python script.
type Scanner struct {
	toolPath   string
	pythonPath string
	enabled    bool
}

// New creates a new scanner instance.
func New(enabled bool) *Scanner {
	home, _ := os.UserHomeDir()
	toolPath := filepath.Join(home, "tools", "JSSecretHunter", "scanner_pro.py")
	pythonPath := filepath.Join(home, "tools", "JSSecretHunter", "venv", "bin", "python3")

	return &Scanner{
		toolPath:   toolPath,
		pythonPath: pythonPath,
		enabled:    enabled,
	}
}

// Scan runs the scanner on a given URL.
func (s *Scanner) Scan(ctx context.Context, url, outputDir string) error {
	if !s.enabled {
		return nil
	}

	args := []string{s.toolPath, "-u", url}

	// Handle Output Directory
	if outputDir != "" {
		// Create a specific subdirectory for JS secrets
		secretsDir := filepath.Join(outputDir, "js_secrets")
		if err := os.MkdirAll(secretsDir, 0755); err != nil {
			return fmt.Errorf("failed to create secrets dir: %v", err)
		}

		// Sanitize URL for filename
		safeName := filepath.Base(url)
		if strings.Contains(safeName, "?") {
			parts := strings.Split(safeName, "?")
			safeName = parts[0]
		}
		// If safeName is empty or just domain, we might need better naming.
		// Use a timestamp or hash?
		// For now simple basename is fine.
		reportPath := filepath.Join(secretsDir, fmt.Sprintf("%s_secrets.txt", safeName))
		args = append(args, "-o", reportPath)
	}

	// Stream output to user so they see what's happening
	cmd := exec.CommandContext(ctx, s.pythonPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("scanner failed: %v", err)
	}

	// Post-Scan: Analyze Report
	if outputDir != "" && len(args) > 2 {
		// reportPath is the last argument after -o
		var reportPath string
		for i, arg := range args {
			if arg == "-o" && i+1 < len(args) {
				reportPath = args[i+1]
				break
			}
		}

		if reportPath != "" {
			content, err := os.ReadFile(reportPath)
			if err == nil {
				sContent := string(content)
				if strings.Contains(sContent, "Total instances found: 0") {
					// Empty report, delete it to reduce noise
					_ = os.Remove(reportPath)
				} else {
					// Findings! Append to summary.
					// We use a simple file-based append. Concurrent writes might be an issue?
					// The Pipeline runs concurrent "Scan" calls.
					// We should probably rely on the OS atomic append or a lock.
					// For simplicity and specialized tool use, we append to SUMMARY.txt in the secretsDir.
					secretsDir := filepath.Dir(reportPath)
					summaryPath := filepath.Join(secretsDir, "SUMMARY.txt")

					// Prepare summary entry
					entry := fmt.Sprintf("\n\n[%s] Found Secrets in %s\n%s\n", time.Now().Format(time.RFC3339), url, sContent)

					f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						defer f.Close()
						if _, err := f.WriteString(entry); err != nil {
							fmt.Printf("[!] Failed to write to summary: %v\n", err)
						}
					}

					// Option: Delete individual file if added to summary?
					// User said "output should be saved on a dedicated dir".
					// Keeping summary + individual significant findings seems best.
					// So specific significant files remain, but NO empty files.
				}
			}
		}
	}

	return nil
}

// IsAvailable checks if the tool is installed.
func (s *Scanner) IsAvailable() bool {
	_, err := os.Stat(s.toolPath)
	return err == nil
}
