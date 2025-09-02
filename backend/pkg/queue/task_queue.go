package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskQueue manages asynchronous batch operations
type TaskQueue struct {
	redis *gredis.Redis
}

// NewTaskQueue creates a new task queue
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		redis: g.Redis(),
	}
}

// Task represents a batch operation task
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

// TaskProgress represents task execution progress
type TaskProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Success   int `json:"success"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
}

// EnqueueTask adds a new task to the queue
func (q *TaskQueue) EnqueueTask(task *Task) error {
	// Set initial status
	if task.Status == "" {
		task.Status = "pending"
	}

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Serialize task
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	// Store task in Redis hash
	taskKey := fmt.Sprintf("batch_task:%s", task.ID)
	ctx := context.Background()
	if _, err := q.redis.HSet(ctx, taskKey, map[string]interface{}{"data": string(taskData)}); err != nil {
		return fmt.Errorf("failed to store task: %w", err)
	}

	// Set task expiration (7 days)
	expireSeconds := int64(7 * 24 * 60 * 60) // 7 days in seconds
	if _, err := q.redis.Expire(ctx, taskKey, expireSeconds); err != nil {
		return fmt.Errorf("failed to set task expiration: %w", err)
	}

	// Add to processing queue
	queueKey := fmt.Sprintf("batch_queue:%s", task.Type)
	if _, err := q.redis.LPush(ctx, queueKey, task.ID); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// GetTask retrieves a task by ID
func (q *TaskQueue) GetTask(taskID string) (*Task, error) {
	taskKey := fmt.Sprintf("batch_task:%s", taskID)

	// Get task data from Redis
	ctx := context.Background()
	taskDataStr, err := q.redis.HGet(ctx, taskKey, "data")
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if taskDataStr.IsEmpty() {
		return nil, errors.New("task not found")
	}

	// Deserialize task
	var task Task
	if err := json.Unmarshal([]byte(taskDataStr.String()), &task); err != nil {
		return nil, fmt.Errorf("failed to deserialize task: %w", err)
	}

	return &task, nil
}

// UpdateTask updates a task's status and data
func (q *TaskQueue) UpdateTask(task *Task) error {
	now := time.Now()
	task.UpdatedAt = &now

	if task.Status == "completed" || task.Status == "failed" {
		task.CompletedAt = &now
	}

	// Serialize task
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	// Update task in Redis
	taskKey := fmt.Sprintf("batch_task:%s", task.ID)
	ctx := context.Background()
	if _, err := q.redis.HSet(ctx, taskKey, map[string]interface{}{"data": string(taskData)}); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// DequeueTask gets the next task from a specific queue
func (q *TaskQueue) DequeueTask(taskType string) (*Task, error) {
	queueKey := fmt.Sprintf("batch_queue:%s", taskType)

	// Get task ID from queue
	ctx := context.Background()
	taskIDVar, err := q.redis.RPop(ctx, queueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}

	if taskIDVar.IsEmpty() {
		return nil, errors.New("no tasks in queue")
	}

	taskID := taskIDVar.String()

	// Get task data
	return q.GetTask(taskID)
}

// UpdateTaskProgress updates a task's progress information
func (q *TaskQueue) UpdateTaskProgress(taskID string, progress *TaskProgress) error {
	task, err := q.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Progress = progress
	task.Status = "processing"

	return q.UpdateTask(task)
}

// UpdateTaskResult updates a task's result data
func (q *TaskQueue) UpdateTaskResult(taskID string, result interface{}, errors []interface{}) error {
	task, err := q.GetTask(taskID)
	if err != nil {
		return err
	}

	// Serialize result
	if result != nil {
		resultData, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to serialize result: %w", err)
		}
		task.Result = resultData
	}

	// Serialize errors
	if errors != nil {
		errorData, err := json.Marshal(errors)
		if err != nil {
			return fmt.Errorf("failed to serialize errors: %w", err)
		}
		task.Errors = errorData
	}

	// Update status
	if len(errors) > 0 {
		task.Status = "completed_with_errors"
	} else {
		task.Status = "completed"
	}

	return q.UpdateTask(task)
}

// FailTask marks a task as failed
func (q *TaskQueue) FailTask(taskID string, errorMessage string) error {
	task, err := q.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = "failed"

	// Add error to task
	errorData := map[string]interface{}{
		"message": errorMessage,
		"code":    "TASK_EXECUTION_ERROR",
		"time":    time.Now(),
	}

	if errorBytes, err := json.Marshal([]interface{}{errorData}); err == nil {
		task.Errors = errorBytes
	}

	return q.UpdateTask(task)
}

// GetQueueSize returns the number of tasks in a specific queue
func (q *TaskQueue) GetQueueSize(taskType string) (int, error) {
	queueKey := fmt.Sprintf("batch_queue:%s", taskType)

	ctx := context.Background()
	size, err := q.redis.LLen(ctx, queueKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return int(size), nil
}

// GetPendingTasks returns all pending tasks of a specific type
func (q *TaskQueue) GetPendingTasks(taskType string, limit int) ([]*Task, error) {
	queueKey := fmt.Sprintf("batch_queue:%s", taskType)

	// Get task IDs from queue
	ctx := context.Background()
	taskIDs, err := q.redis.LRange(ctx, queueKey, 0, int64(limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	var tasks []*Task
	for _, taskIDVar := range taskIDs {
		taskID := taskIDVar.String()
		task, err := q.GetTask(taskID)
		if err != nil {
			continue // Skip tasks that can't be loaded
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CleanupExpiredTasks removes expired task data
func (q *TaskQueue) CleanupExpiredTasks() error {
	// This would typically be run by a background job
	// to clean up expired tasks and their associated data

	// Get all task keys
	pattern := "batch_task:*"
	ctx := context.Background()
	keys, err := q.redis.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get task keys: %w", err)
	}

	expiredCount := 0
	for _, key := range keys {
		// Check if key exists (not expired)
		exists, err := q.redis.Exists(ctx, key)
		if err != nil {
			continue
		}

		if exists == 0 {
			expiredCount++
		}
	}

	g.Log().Infof(context.Background(), "Cleaned up %d expired batch tasks", expiredCount)
	return nil
}

// GetTaskStatistics returns statistics about batch operations
func (q *TaskQueue) GetTaskStatistics() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"total_tasks":      0,
		"pending_tasks":    0,
		"processing_tasks": 0,
		"completed_tasks":  0,
		"failed_tasks":     0,
	}

	// Get all task keys
	pattern := "batch_task:*"
	ctx := context.Background()
	keys, err := q.redis.Keys(ctx, pattern)
	if err != nil {
		return stats, fmt.Errorf("failed to get task keys: %w", err)
	}

	stats["total_tasks"] = len(keys)

	// Count tasks by status
	for _, key := range keys {
		taskDataStr, err := q.redis.HGet(ctx, key, "data")
		if err != nil {
			continue
		}

		if taskDataStr.IsEmpty() {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(taskDataStr.String()), &task); err != nil {
			continue
		}

		switch task.Status {
		case "pending":
			stats["pending_tasks"] = stats["pending_tasks"].(int) + 1
		case "processing":
			stats["processing_tasks"] = stats["processing_tasks"].(int) + 1
		case "completed", "completed_with_errors":
			stats["completed_tasks"] = stats["completed_tasks"].(int) + 1
		case "failed":
			stats["failed_tasks"] = stats["failed_tasks"].(int) + 1
		}
	}

	return stats, nil
}
