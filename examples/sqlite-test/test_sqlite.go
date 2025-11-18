package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"syslog-visualizer/internal/parser"
	"syslog-visualizer/internal/storage"
)

func main() {
	// Create test database
	dbPath := "test_syslog.db"

	// Remove existing test database
	os.Remove(dbPath)

	// Initialize SQLite storage
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	defer os.Remove(dbPath) // Clean up after test

	fmt.Println("SQLite storage initialized with GORM")

	// Create test messages
	testMessages := []*parser.SyslogMessage{
		{
			Timestamp: time.Now(),
			Hostname:  "test-server-1",
			Facility:  1, // user
			Severity:  6, // info
			Tag:       "test-app",
			Message:   "This is a test message",
			Raw:       "<14>Oct 11 22:14:15 test-server-1 test-app: This is a test message",
			PID:       "1234",
			AppName:   "test-app",
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Hostname:  "test-server-2",
			Facility:  3, // daemon
			Severity:  3, // error
			Tag:       "error-service",
			Message:   "An error occurred",
			Raw:       "<27>Oct 11 21:14:15 test-server-2 error-service: An error occurred",
			PID:       "5678",
			AppName:   "error-service",
		},
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Hostname:  "test-server-1",
			Facility:  0, // kern
			Severity:  2, // critical
			Tag:       "kernel",
			Message:   "Critical kernel message",
			Raw:       "<2>Oct 11 20:14:15 test-server-1 kernel: Critical kernel message",
			AppName:   "kernel",
		},
	}

	// Store messages
	fmt.Println("\nStoring test messages...")
	for i, msg := range testMessages {
		if err := store.Store(msg); err != nil {
			log.Fatalf("Failed to store message %d: %v", i+1, err)
		}
		fmt.Printf("  Stored message %d: [%s] %s - %s\n",
			i+1, msg.SeverityName(), msg.Hostname, msg.Message)
	}

	// Query all messages
	fmt.Println("\nQuerying all messages...")
	filters := storage.QueryFilters{Limit: 100}
	messages, err := store.Query(filters)
	if err != nil {
		log.Fatalf("Failed to query messages: %v", err)
	}
	fmt.Printf("  Retrieved %d messages\n", len(messages))

	// Verify count
	if len(messages) != len(testMessages) {
		log.Fatalf("Expected %d messages, got %d", len(testMessages), len(messages))
	}

	// Query with filters
	fmt.Println("\nQuerying with filters (severity=3)...")
	severity := 3
	filteredQuery := storage.QueryFilters{
		Severity: &severity,
		Limit:    100,
	}
	filtered, err := store.Query(filteredQuery)
	if err != nil {
		log.Fatalf("Failed to query filtered messages: %v", err)
	}
	fmt.Printf("  Retrieved %d messages with severity=error\n", len(filtered))
	if len(filtered) != 1 {
		log.Fatalf("Expected 1 error message, got %d", len(filtered))
	}

	// Test search
	fmt.Println("\nSearching for 'error'...")
	searchResults, err := store.SearchMessages("error", 10)
	if err != nil {
		log.Fatalf("Failed to search messages: %v", err)
	}
	fmt.Printf("  Found %d messages containing 'error'\n", len(searchResults))

	// Get stats
	fmt.Println("\nGetting database statistics...")
	stats, err := store.Stats()
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("  Total messages: %v\n", stats["total_messages"])
	fmt.Printf("  By severity: %v\n", stats["by_severity"])
	if dbSize, ok := stats["db_size_bytes"]; ok {
		fmt.Printf("  Database size: %v bytes\n", dbSize)
	}

	// Test deletion
	fmt.Println("\nTesting deletion of old messages...")
	deleted, err := store.DeleteOlderThan(90 * time.Minute)
	if err != nil {
		log.Fatalf("Failed to delete old messages: %v", err)
	}
	fmt.Printf("  Deleted %d old messages\n", deleted)

	// Verify remaining
	remaining, err := store.Query(storage.QueryFilters{Limit: 100})
	if err != nil {
		log.Fatalf("Failed to query remaining messages: %v", err)
	}
	fmt.Printf("  %d messages remain\n", len(remaining))

	fmt.Println("\nAll SQLite storage tests passed!")
}
