package targetlist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTargetList(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    []string
		shouldError bool
	}{
		{
			name: "valid targets",
			content: `example.com
google.com
github.com`,
			expected:    []string{"example.com", "google.com", "github.com"},
			shouldError: false,
		},
		{
			name: "targets with comments",
			content: `# Main targets
example.com
# Bug bounty program
*.bugcrowd.com
github.com`,
			expected:    []string{"example.com", "*.bugcrowd.com", "github.com"},
			shouldError: false,
		},
		{
			name: "targets with empty lines",
			content: `example.com

google.com

`,
			expected:    []string{"example.com", "google.com"},
			shouldError: false,
		},
		{
			name: "targets with whitespace",
			content: `  example.com  
  google.com
github.com  `,
			expected:    []string{"example.com", "google.com", "github.com"},
			shouldError: false,
		},
		{
			name: "targets with protocols",
			content: `https://example.com
http://google.com
github.com/`,
			expected:    []string{"example.com", "google.com", "github.com"},
			shouldError: false,
		},
		{
			name: "empty file",
			content: `


`,
			expected:    nil,
			shouldError: true,
		},
		{
			name: "only comments",
			content: `# Comment 1
# Comment 2`,
			expected:    nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), "targets.txt")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse targets
			targets, err := ParseTargetList(tmpFile)

			// Check error expectation
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check results
			if len(targets) != len(tt.expected) {
				t.Errorf("Expected %d targets, got %d", len(tt.expected), len(targets))
			}

			for i, expected := range tt.expected {
				if i >= len(targets) {
					break
				}
				if targets[i] != expected {
					t.Errorf("Target %d: expected %s, got %s", i, expected, targets[i])
				}
			}
		})
	}
}

func TestParseTargetList_NonExistentFile(t *testing.T) {
	_, err := ParseTargetList("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestSanitizeTarget(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"  example.com  ", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"example.com/", "example.com"},
		{"*.example.com", "*.example.com"},
		{"", ""},
		{"  ", ""},
		{"ab", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeTarget(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsComment(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"# This is a comment", true},
		{"  # This is also a comment", true},
		{"Not a comment", false},
		{"", false},
		{"  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isComment(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"  ", true},
		{"\t", true},
		{"not empty", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
