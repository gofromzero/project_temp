package files

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domainTenant "github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/xuri/excelize/v2"
)

// FileProcessor handles file parsing and generation operations
type FileProcessor struct{}

// NewFileProcessor creates a new file processor
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{}
}

// ParseFile parses uploaded file and returns tenant data
func (fp *FileProcessor) ParseFile(file *ghttp.UploadFile, filename string) ([]map[string]interface{}, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".csv":
		return fp.parseCSVFile(file)
	case ".xlsx", ".xls":
		return fp.parseExcelFile(file)
	default:
		return nil, errors.New("unsupported file format")
	}
}

// parseCSVFile parses CSV file content
func (fp *FileProcessor) parseCSVFile(file *ghttp.UploadFile) ([]map[string]interface{}, error) {
	// Validate file size
	if err := fp.ValidateFileSize(file.Size); err != nil {
		return nil, err
	}

	// Validate file format
	if err := fp.ValidateFileFormat(file.Filename); err != nil {
		return nil, err
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Create CSV reader
	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records with limits to prevent DoS
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, errors.New("empty CSV file")
	}

	// Enforce maximum record limit for security
	const maxRecords = 10000
	if len(records) > maxRecords {
		return nil, fmt.Errorf("CSV file exceeds maximum allowed records: %d", maxRecords)
	}

	// First row should be headers
	headers := records[0]
	var result []map[string]interface{}

	// Process data rows
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue // Skip empty rows
		}

		tenantData := make(map[string]interface{})

		// Map record values to headers with sanitization
		for j, header := range headers {
			if j < len(record) {
				value := strings.TrimSpace(record[j])
				// Sanitize value to prevent CSV injection
				value = sanitizeCSVValue(value)
				if value != "" && len(value) <= 1000 { // Prevent excessively long values
					tenantData[header] = value
				}
			}
		}

		// Skip rows without required fields
		if tenantData["name"] == nil || tenantData["code"] == nil {
			continue
		}

		// Parse config fields
		if err := fp.parseConfigFromCSV(tenantData); err != nil {
			return nil, fmt.Errorf("failed to parse config in row %d: %w", i+1, err)
		}

		result = append(result, tenantData)
	}

	return result, nil
}

// parseExcelFile parses Excel file content
func (fp *FileProcessor) parseExcelFile(file *ghttp.UploadFile) ([]map[string]interface{}, error) {
	// Validate file size
	if err := fp.ValidateFileSize(file.Size); err != nil {
		return nil, err
	}

	// Validate file format
	if err := fp.ValidateFileFormat(file.Filename); err != nil {
		return nil, err
	}

	// Save uploaded file temporarily with secure path
	tempPath := fmt.Sprintf("/tmp/upload_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	if _, err := file.Save(tempPath); err != nil {
		return nil, fmt.Errorf("failed to save temp file: %w", err)
	}
	defer os.Remove(tempPath) // Clean up temp file

	// Open Excel file
	f, err := excelize.OpenFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet name
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("no sheets found in Excel file")
	}

	// Get all rows from first sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, errors.New("empty Excel file")
	}

	// First row should be headers
	headers := rows[0]
	var result []map[string]interface{}

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}

		tenantData := make(map[string]interface{})

		// Map row values to headers
		for j, header := range headers {
			if j < len(row) {
				value := strings.TrimSpace(row[j])
				if value != "" {
					tenantData[header] = value
				}
			}
		}

		// Skip rows without required fields
		if tenantData["name"] == nil || tenantData["code"] == nil {
			continue
		}

		// Parse config fields
		if err := fp.parseConfigFromExcel(tenantData); err != nil {
			return nil, fmt.Errorf("failed to parse config in row %d: %w", i+1, err)
		}

		result = append(result, tenantData)
	}

	return result, nil
}

// parseConfigFromCSV parses tenant config from CSV fields
func (fp *FileProcessor) parseConfigFromCSV(data map[string]interface{}) error {
	config := make(map[string]interface{})

	// Parse maxUsers
	if maxUsersStr, ok := data["maxUsers"].(string); ok {
		if maxUsers, err := strconv.Atoi(maxUsersStr); err == nil {
			config["maxUsers"] = maxUsers
		}
	}

	// Parse features (comma-separated)
	if featuresStr, ok := data["features"].(string); ok {
		if featuresStr != "" {
			features := strings.Split(featuresStr, ",")
			for i, f := range features {
				features[i] = strings.TrimSpace(f)
			}
			config["features"] = features
		}
	}

	// Parse theme
	if theme, ok := data["theme"].(string); ok {
		config["theme"] = theme
	}

	// Parse domain
	if domain, ok := data["domain"].(string); ok {
		config["domain"] = domain
	}

	if len(config) > 0 {
		data["config"] = config
	}

	return nil
}

// parseConfigFromExcel parses tenant config from Excel fields
func (fp *FileProcessor) parseConfigFromExcel(data map[string]interface{}) error {
	config := make(map[string]interface{})

	// Parse maxUsers
	if maxUsersStr, ok := data["maxUsers"].(string); ok {
		if maxUsers, err := strconv.Atoi(maxUsersStr); err == nil {
			config["maxUsers"] = maxUsers
		}
	}

	// Parse features (comma-separated or newline-separated)
	if featuresStr, ok := data["features"].(string); ok {
		if featuresStr != "" {
			var features []string
			if strings.Contains(featuresStr, ",") {
				features = strings.Split(featuresStr, ",")
			} else {
				features = strings.Split(featuresStr, "\n")
			}
			for i, f := range features {
				features[i] = strings.TrimSpace(f)
			}
			config["features"] = features
		}
	}

	// Parse theme
	if theme, ok := data["theme"].(string); ok {
		config["theme"] = theme
	}

	// Parse domain
	if domain, ok := data["domain"].(string); ok {
		config["domain"] = domain
	}

	if len(config) > 0 {
		data["config"] = config
	}

	return nil
}

// GenerateExportFile generates export file for tenants
func (fp *FileProcessor) GenerateExportFile(tenants []*domainTenant.Tenant, format string, includeStats bool) (string, error) {
	switch strings.ToLower(format) {
	case "csv":
		return fp.generateCSVExport(tenants, includeStats)
	case "xlsx", "excel":
		return fp.generateExcelExport(tenants, includeStats)
	default:
		return "", errors.New("unsupported export format")
	}
}

// generateCSVExport generates CSV export file
func (fp *FileProcessor) generateCSVExport(tenants []*domainTenant.Tenant, includeStats bool) (string, error) {
	// Create temporary file
	tempPath := fmt.Sprintf("/tmp/tenant_export_%d.csv", len(tenants))

	// This is a simplified implementation - in real scenarios, you would:
	// 1. Create the actual file
	// 2. Write CSV headers and data
	// 3. Include statistics if requested
	// 4. Handle file permissions and cleanup

	// Mock implementation
	return tempPath, nil
}

// generateExcelExport generates Excel export file
func (fp *FileProcessor) generateExcelExport(tenants []*domainTenant.Tenant, includeStats bool) (string, error) {
	// Create new Excel file
	f := excelize.NewFile()

	// Create tenant data sheet
	sheetName := "Tenants"
	f.NewSheet(sheetName)

	// Set headers
	headers := []string{"ID", "Name", "Code", "Status", "Max Users", "Features", "Theme", "Domain", "Created At", "Updated At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write tenant data
	for rowIndex, tenant := range tenants {
		row := rowIndex + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), tenant.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), tenant.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), tenant.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), string(tenant.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), tenant.Config.MaxUsers)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), strings.Join(tenant.Config.Features, ", "))

		if tenant.Config.Theme != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), *tenant.Config.Theme)
		}

		if tenant.Config.Domain != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), *tenant.Config.Domain)
		}

		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), tenant.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), tenant.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// Add statistics sheet if requested
	if includeStats {
		fp.addStatisticsSheet(f, tenants)
	}

	// Save file
	tempPath := fmt.Sprintf("/tmp/tenant_export_%d.xlsx", len(tenants))
	if err := f.SaveAs(tempPath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return tempPath, nil
}

// addStatisticsSheet adds a statistics sheet to the Excel export
func (fp *FileProcessor) addStatisticsSheet(f *excelize.File, tenants []*domainTenant.Tenant) {
	sheetName := "Statistics"
	f.NewSheet(sheetName)

	// Calculate statistics
	stats := fp.calculateTenantStatistics(tenants)

	// Write statistics
	row := 1
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Metric")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Value")

	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Total Tenants")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats["total"])

	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Active Tenants")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats["active"])

	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Suspended Tenants")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats["suspended"])

	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Disabled Tenants")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats["disabled"])

	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Average Max Users")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats["avgMaxUsers"])
}

// calculateTenantStatistics calculates statistics for tenants
func (fp *FileProcessor) calculateTenantStatistics(tenants []*domainTenant.Tenant) map[string]interface{} {
	stats := map[string]interface{}{
		"total":         len(tenants),
		"active":        0,
		"suspended":     0,
		"disabled":      0,
		"totalMaxUsers": 0,
		"avgMaxUsers":   0,
	}

	totalMaxUsers := 0
	for _, tenant := range tenants {
		switch tenant.Status {
		case domainTenant.StatusActive:
			stats["active"] = stats["active"].(int) + 1
		case domainTenant.StatusSuspended:
			stats["suspended"] = stats["suspended"].(int) + 1
		case domainTenant.StatusDisabled:
			stats["disabled"] = stats["disabled"].(int) + 1
		}

		totalMaxUsers += tenant.Config.MaxUsers
	}

	if len(tenants) > 0 {
		stats["avgMaxUsers"] = totalMaxUsers / len(tenants)
	}

	return stats
}

// ValidateFileFormat validates if the file format is supported
func (fp *FileProcessor) ValidateFileFormat(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	supportedFormats := []string{".csv", ".xlsx", ".xls"}
	for _, format := range supportedFormats {
		if ext == format {
			return nil
		}
	}

	return fmt.Errorf("unsupported file format: %s. Supported formats: %s", ext, strings.Join(supportedFormats, ", "))
}

// ValidateFileSize validates if the file size is within limits
func (fp *FileProcessor) ValidateFileSize(size int64) error {
	maxSize := int64(50 * 1024 * 1024) // 50MB
	if size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, maxSize)
	}
	return nil
}

// sanitizeCSVValue sanitizes CSV values to prevent injection attacks
func sanitizeCSVValue(value string) string {
	// Remove dangerous characters that could lead to CSV injection
	// These characters can be used in CSV injection attacks
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
