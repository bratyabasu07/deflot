package targetlist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseTargetList reads a target list file and returns cleaned, valid targets.
// It skips empty lines, comments (starting with #), and trims whitespace.
// Returns an error if the file doesn't exist or contains no valid targets.
func ParseTargetList(filepath string) ([]string, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("target list file not found: %s", filepath)
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open target list: %w", err)
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments
		if isEmpty(line) || isComment(line) {
			continue
		}

		// Sanitize and validate
		target := sanitizeTarget(line)
		if target == "" {
			fmt.Printf("[!] Warning: Skipping invalid target at line %d: %s\n", lineNum, line)
			continue
		}

		targets = append(targets, target)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading target list: %w", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("target list contains no valid targets")
	}

	return targets, nil
}

// sanitizeTarget strips whitespace and performs basic validation.
// Returns empty string if target is invalid.
func sanitizeTarget(target string) string {
	// Trim whitespace
	target = strings.TrimSpace(target)

	// Basic validation: must have at least one dot (domain.tld or *.domain.tld)
	// or be a valid wildcard pattern
	if target == "" {
		return ""
	}

	// Remove protocol if accidentally included
	target = strings.TrimPrefix(target, "http://")
	target = strings.TrimPrefix(target, "https://")

	// Remove trailing slash
	target = strings.TrimSuffix(target, "/")

	// Basic check: should contain at least one character
	// More sophisticated validation happens in the context/domain validator
	if len(target) < 3 {
		return ""
	}

	return target
}

// isComment checks if a line is a comment (starts with #).
func isComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "#")
}

// isEmpty checks if a line is empty or contains only whitespace.
func isEmpty(line string) bool {
	return strings.TrimSpace(line) == ""
}
