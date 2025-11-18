package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"syslog-visualizer/internal/collector"
	"syslog-visualizer/internal/framing"
	"syslog-visualizer/internal/parser"
	"syslog-visualizer/internal/storage"
)

func main() {
	fmt.Println("Syslog Collector starting...")

	// Initialize storage
	store := storage.NewMemoryStorage()
	defer store.Close()

	// Create message handler that stores messages
	handler := func(msg *parser.SyslogMessage) error {
		log.Printf("[%s] %s %s[%s]: %s",
			msg.SeverityName(),
			msg.Hostname,
			msg.Tag,
			msg.PID,
			msg.Message,
		)
		return store.Store(msg)
	}

	// Configure collector
	cfg := collector.Config{
		Address:       ":514",
		Protocol:      "udp", // Use "tcp" or "both" for TCP support
		FramingMethod: framing.NonTransparent,
		Handler:       handler,
	}

	// Create collector
	col, err := collector.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start collector in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := col.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("Shutdown signal received")
	case err := <-errChan:
		log.Printf("Collector error: %v", err)
	}

	// Stop collector
	if err := col.Stop(); err != nil {
		log.Printf("Error stopping collector: %v", err)
	}

	log.Println("Collector stopped successfully")
}
