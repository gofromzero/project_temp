package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
)

// TestDatabaseConnection tests MySQL database connectivity
// Returns error if connection fails or configuration is missing
func TestDatabaseConnection() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Safely handle GoFrame DB configuration panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("database configuration error: %v", r)
		}
	}()

	// Try to get DB connection
	db := g.DB()
	if db == nil {
		return fmt.Errorf("database not configured - ensure GoFrame config.yaml is accessible")
	}

	_, dbErr := db.Query(ctx, "SELECT 1")
	if dbErr != nil {
		return fmt.Errorf("database connection failed: %w", dbErr)
	}
	return nil
}

// TestRedisConnection tests Redis connectivity
// Uses configuration from GoFrame config or falls back to localhost:6379
func TestRedisConnection() (err error) {
	// Default Redis address
	addr := "localhost:6379"

	// Safely try to get Redis address from GoFrame config
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Config loading panicked, keep default addr
			}
		}()

		if redisConfig := g.Cfg().MustGet(context.Background(), "redis.default.address"); !redisConfig.IsEmpty() {
			addr = redisConfig.String()
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0, // Default DB
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis connection failed to %s: %w", addr, err)
	}
	return nil
}
