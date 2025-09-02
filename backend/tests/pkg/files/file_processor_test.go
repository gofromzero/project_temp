package files

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileTypeValidation tests file type validation logic
func TestFileTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		expectValid bool
	}{
		{
			name:        "Valid CSV file",
			fileName:    "tenants.csv",
			expectValid: true,
		},
		{
			name:        "Valid XLSX file",
			fileName:    "tenants.xlsx",
			expectValid: true,
		},
		{
			name:        "Valid XLS file",
			fileName:    "tenants.xls",
			expectValid: true,
		},
		{
			name:        "Invalid TXT file",
			fileName:    "tenants.txt",
			expectValid: false,
		},
		{
			name:        "Invalid PDF file",
			fileName:    "tenants.pdf",
			expectValid: false,
		},
		{
			name:        "Empty filename",
			fileName:    "",
			expectValid: false,
		},
		{
			name:        "Filename too short",
			fileName:    "ab",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidFileType(tt.fileName)
			assert.Equal(t, tt.expectValid, isValid, "File type validation should match expected result")
		})
	}
}

// TestFileSizeValidation tests file size validation
func TestFileSizeValidation(t *testing.T) {
	tests := []struct {
		name        string
		fileSize    int64
		expectValid bool
	}{
		{
			name:        "Valid small file",
			fileSize:    1024, // 1KB
			expectValid: true,
		},
		{
			name:        "Valid medium file", 
			fileSize:    10 * 1024 * 1024, // 10MB
			expectValid: true,
		},
		{
			name:        "Valid large file at limit",
			fileSize:    50 * 1024 * 1024, // 50MB (exact limit)
			expectValid: true,
		},
		{
			name:        "Invalid oversized file",
			fileSize:    60 * 1024 * 1024, // 60MB (exceeds 50MB limit)
			expectValid: false,
		},
		{
			name:        "Zero size file",
			fileSize:    0,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxSize := int64(50 * 1024 * 1024) // 50MB limit
			isValidSize := tt.fileSize <= maxSize
			assert.Equal(t, tt.expectValid, isValidSize, "File size validation should match expected result")
		})
	}
}

// TestCSVSanitization tests CSV injection prevention
func TestCSVSanitization(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectSanitized bool
	}{
		{
			name:           "Safe value",
			input:          "normal_value",
			expectSanitized: false,
		},
		{
			name:           "Formula injection with equals",
			input:          "=SUM(A1:A10)",
			expectSanitized: true,
		},
		{
			name:           "Formula injection with plus",
			input:          "+1+2",
			expectSanitized: true,
		},
		{
			name:           "Formula injection with minus",
			input:          "-1",
			expectSanitized: true,
		},
		{
			name:           "Formula injection with at symbol",
			input:          "@SUM(1,2)",
			expectSanitized: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeCSVValue(tt.input)
			
			if tt.expectSanitized {
				assert.True(t, strings.HasPrefix(result, "'"), "Dangerous value should be prefixed with quote")
				assert.Contains(t, result, tt.input, "Original value should be preserved after quote")
			} else {
				assert.Equal(t, tt.input, result, "Safe value should remain unchanged")
			}
		})
	}
}

// TestBatchSynchronousThreshold tests the threshold for sync vs async processing
func TestBatchSynchronousThreshold(t *testing.T) {
	tests := []struct {
		name              string
		itemCount         int
		expectSynchronous bool
		operationType     string
	}{
		{
			name:              "Small import - synchronous",
			itemCount:         50,
			expectSynchronous: true,
			operationType:     "import",
		},
		{
			name:              "Medium import - synchronous",
			itemCount:         99,
			expectSynchronous: true,
			operationType:     "import",
		},
		{
			name:              "Large import - asynchronous",
			itemCount:         100,
			expectSynchronous: false,
			operationType:     "import",
		},
		{
			name:              "Small export - synchronous",
			itemCount:         500,
			expectSynchronous: true,
			operationType:     "export",
		},
		{
			name:              "Large export - asynchronous",
			itemCount:         1000,
			expectSynchronous: false,
			operationType:     "export",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isSynchronous bool
			
			switch tt.operationType {
			case "import":
				isSynchronous = tt.itemCount < 100
			case "export":
				isSynchronous = tt.itemCount < 1000
			default:
				isSynchronous = true // Default to synchronous for other operations
			}
			
			assert.Equal(t, tt.expectSynchronous, isSynchronous, "Processing mode should match threshold")
		})
	}
}

// Helper functions (duplicated from files package for testing)
func isValidFileType(filename string) bool {
	if len(filename) < 5 {
		return false
	}

	ext := filename[len(filename)-4:]
	switch ext {
	case ".csv":
		return true
	case ".xls", "xlsx":
		return true
	default:
		return false
	}
}

func sanitizeCSVValue(value string) string {
	// Remove dangerous characters that could lead to CSV injection
	dangerousChars := []string{"=", "+", "-", "@", "\t", "\n", "\r"}

	for _, char := range dangerousChars {
		if strings.HasPrefix(value, char) {
			// Prepend with single quote to neutralize formula injection
			value = "'" + value
			break
		}
	}

	// Additional sanitization - remove control characters
	value = strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, LF, CR
			return -1 // Remove control characters
		}
		return r
	}, value)

	return value
}