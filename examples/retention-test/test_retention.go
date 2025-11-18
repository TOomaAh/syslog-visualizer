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
	dbPath := "test_retention.db"

	// Remove existing test database
	os.Remove(dbPath)

	// Initialize SQLite storage
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	defer os.Remove(dbPath) // Clean up after test

	fmt.Println("Testing data retention feature\n")

	// Create test messages with different timestamps
	now := time.Now()
	testMessages := []*parser.SyslogMessage{
		{
			Timestamp: now,
			Hostname:  "recent-host",
			Facility:  1,
			Severity:  6,
			Tag:       "recent",
			Message:   "Recent message (should be kept)",
			Raw:       "<14>Recent message",
		},
		{
			Timestamp: now.Add(-10 * time.Minute),
			Hostname:  "old-host-1",
			Facility:  1,
			Severity:  6,
			Tag:       "old",
			Message:   "10 minutes old (should be deleted)",
			Raw:       "<14>10 minutes old",
		},
		{
			Timestamp: now.Add(-20 * time.Minute),
			Hostname:  "old-host-2",
			Facility:  1,
			Severity:  6,
			Tag:       "old",
			Message:   "20 minutes old (should be deleted)",
			Raw:       "<14>20 minutes old",
		},
		{
			Timestamp: now.Add(-3 * time.Minute),
			Hostname:  "medium-host",
			Facility:  1,
			Severity:  6,
			Tag:       "medium",
			Message:   "3 minutes old (should be kept)",
			Raw:       "<14>3 minutes old",
		},
	}

	// Store messages
	fmt.Println("Storing test messages with different timestamps:")
	for i, msg := range testMessages {
		if err := store.Store(msg); err != nil {
			log.Fatalf("Failed to store message %d: %v", i+1, err)
		}
		age := now.Sub(msg.Timestamp)
		fmt.Printf("  [%2d] %s (age: %v)\n", i+1, msg.Message, age.Round(time.Minute))
	}

	// Count initial messages
	initial, err := store.Query(storage.QueryFilters{Limit: 100})
	if err != nil {
		log.Fatalf("Failed to query initial messages: %v", err)
	}
	fmt.Printf("\nInitial message count: %d\n", len(initial))

	// Test retention with 5-minute retention period
	retentionPeriod := 5 * time.Minute
	fmt.Printf("\nApplying retention policy: delete messages older than %v\n", retentionPeriod)

	deleted, err := store.DeleteOlderThan(retentionPeriod)
	if err != nil {
		log.Fatalf("Failed to delete old messages: %v", err)
	}

	fmt.Printf("Deleted %d old messages\n", deleted)

	// Verify remaining messages
	remaining, err := store.Query(storage.QueryFilters{Limit: 100})
	if err != nil {
		log.Fatalf("Failed to query remaining messages: %v", err)
	}

	fmt.Printf("Remaining message count: %d\n\n", len(remaining))

	// Display remaining messages
	fmt.Println("Remaining messages:")
	for i, msg := range remaining {
		age := now.Sub(msg.Timestamp)
		fmt.Printf("  [%d] %s (age: %v)\n", i+1, msg.Message, age.Round(time.Minute))
	}

	// Verify expectations
	expectedRemaining := 2 // Only recent and 3-minute old messages should remain
	if len(remaining) != expectedRemaining {
		log.Fatalf("ERROR: Expected %d remaining messages, got %d", expectedRemaining, len(remaining))
	}

	fmt.Println("\nData retention test passed!")
	fmt.Printf("  - Retention period: %v\n", retentionPeriod)
	fmt.Printf("  - Messages deleted: %d\n", deleted)
	fmt.Printf("  - Messages retained: %d\n", len(remaining))
}
