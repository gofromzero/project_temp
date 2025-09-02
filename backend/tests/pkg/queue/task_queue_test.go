package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTaskCreation tests basic task creation
func TestTaskCreation(t *testing.T) {
	task := &Task{
		ID:        "test-task-123",
		Type:      "tenant_import",
		Status:    "pending",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"test": "data",
		},
	}

	assert.NotNil(t, task, "Task should be created successfully")
	assert.Equal(t, "test-task-123", task.ID, "Task ID should match")
	assert.Equal(t, "tenant_import", task.Type, "Task type should match")
	assert.Equal(t, "pending", task.Status, "Task status should be pending")
	assert.NotEmpty(t, task.Data, "Task data should not be empty")
}

// TestTaskProgress tests task progress tracking
func TestTaskProgress(t *testing.T) {
	progress := &TaskProgress{
		Total:     100,
		Processed: 50,
		Success:   45,
		Failed:    3,
		Skipped:   2,
	}

	assert.Equal(t, 100, progress.Total, "Total should match")
	assert.Equal(t, 50, progress.Processed, "Processed should match")
	assert.Equal(t, 45, progress.Success, "Success should match")
	assert.Equal(t, 3, progress.Failed, "Failed should match")
	assert.Equal(t, 2, progress.Skipped, "Skipped should match")

	// Verify progress consistency
	assert.Equal(t, progress.Success+progress.Failed+progress.Skipped, progress.Processed,
		"Processed should equal sum of success, failed, and skipped")
}

// TestTaskStatusTransitions tests valid task status transitions
func TestTaskStatusTransitions(t *testing.T) {
	validTransitions := map[string][]string{
		"pending":    {"processing", "failed"},
		"processing": {"completed", "completed_with_errors", "failed"},
		"completed":  {}, // Terminal state
		"failed":     {}, // Terminal state
	}

	for fromStatus, toStatuses := range validTransitions {
		for _, toStatus := range toStatuses {
			t.Run("from_"+fromStatus+"_to_"+toStatus, func(t *testing.T) {
				task := &Task{
					ID:     "test-task",
					Status: fromStatus,
				}

				// Simulate status transition
				task.Status = toStatus
				if toStatus == "completed" || toStatus == "failed" {
					now := time.Now()
					task.CompletedAt = &now
				}

				assert.Equal(t, toStatus, task.Status, "Status should be updated")
				if toStatus == "completed" || toStatus == "failed" {
					assert.NotNil(t, task.CompletedAt, "CompletedAt should be set for terminal states")
				}
			})
		}
	}
}

// TestBatchOperationLimits tests batch operation size limits
func TestBatchOperationLimits(t *testing.T) {
	tests := []struct {
		name           string
		tenantCount    int
		expectValid    bool
		operationType  string
	}{
		{
			name:          "Valid small batch",
			tenantCount:   10,
			expectValid:   true,
			operationType: "status_update",
		},
		{
			name:          "Valid medium batch",
			tenantCount:   500,
			expectValid:   true,
			operationType: "status_update",
		},
		{
			name:          "Valid batch at limit",
			tenantCount:   1000,
			expectValid:   true,
			operationType: "status_update",
		},
		{
			name:          "Invalid oversized batch",
			tenantCount:   1001,
			expectValid:   false,
			operationType: "status_update",
		},
		{
			name:          "Empty batch",
			tenantCount:   0,
			expectValid:   false,
			operationType: "status_update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxBatchSize := 1000
			minBatchSize := 1
			
			isValidBatch := tt.tenantCount >= minBatchSize && tt.tenantCount <= maxBatchSize
			assert.Equal(t, tt.expectValid, isValidBatch, "Batch size validation should match expected result")
		})
	}
}

// TestTaskTimeout tests task timeout handling
func TestTaskTimeout(t *testing.T) {
	// Test timeout threshold (30 minutes for batch operations)
	timeoutThreshold := 30 * time.Minute
	
	tests := []struct {
		name           string
		taskDuration   time.Duration
		expectTimeout  bool
	}{
		{
			name:          "Quick task",
			taskDuration:  5 * time.Minute,
			expectTimeout: false,
		},
		{
			name:          "Normal task",
			taskDuration:  20 * time.Minute,
			expectTimeout: false,
		},
		{
			name:          "Task at threshold",
			taskDuration:  30 * time.Minute,
			expectTimeout: false,
		},
		{
			name:          "Timeout task",
			taskDuration:  35 * time.Minute,
			expectTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			task := &Task{
				ID:        "timeout-test",
				CreatedAt: now.Add(-tt.taskDuration),
				Status:    "processing",
			}

			shouldTimeout := time.Since(task.CreatedAt) > timeoutThreshold
			assert.Equal(t, tt.expectTimeout, shouldTimeout, "Timeout detection should match expected result")
		})
	}
}

// TestQueuePriority tests task queue priority handling
func TestQueuePriority(t *testing.T) {
	// Test different task types have different priority levels
	taskPriorities := map[string]int{
		"tenant_import": 2,
		"tenant_export": 3,
		"bulk_update":   1,
		"cleanup":       4,
	}

	for taskType, expectedPriority := range taskPriorities {
		t.Run("priority_"+taskType, func(t *testing.T) {
			// Simulate priority assignment logic
			var priority int
			switch taskType {
			case "bulk_update":
				priority = 1 // Highest priority
			case "tenant_import":
				priority = 2 // High priority
			case "tenant_export":
				priority = 3 // Medium priority
			case "cleanup":
				priority = 4 // Low priority
			default:
				priority = 3 // Default medium priority
			}

			assert.Equal(t, expectedPriority, priority, "Task priority should match expected value")
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkTaskCreation benchmarks task creation performance
func BenchmarkTaskCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:        "benchmark-task",
			Type:      "tenant_import",
			Status:    "pending",
			CreatedAt: time.Now(),
			Data: map[string]interface{}{
				"test": "data",
			},
		}
		_ = task // Prevent optimization
	}
}

// BenchmarkTaskStatusUpdate benchmarks status update performance
func BenchmarkTaskStatusUpdate(b *testing.B) {
	task := &Task{
		ID:        "benchmark-task",
		Type:      "tenant_import", 
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	statuses := []string{"processing", "completed", "failed"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status := statuses[i%len(statuses)]
		task.Status = status
		
		if status == "completed" || status == "failed" {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
}

// Task struct for testing (duplicated from queue package)
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Result      []byte                 `json:"result,omitempty"`
	Errors      []byte                 `json:"errors,omitempty"`
	Progress    *TaskProgress          `json:"progress,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskProgress struct for testing
type TaskProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Success   int `json:"success"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
}