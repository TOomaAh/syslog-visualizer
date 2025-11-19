package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"syslog-visualizer/internal/auth"
	"syslog-visualizer/internal/collector"
	"syslog-visualizer/internal/framing"
	"syslog-visualizer/internal/parser"
	"syslog-visualizer/internal/storage"
)

type RetentionConfig struct {
	RetentionPeriod time.Duration
	CleanupInterval time.Duration
	Enabled         bool
}

func main() {
	retentionPeriod := flag.String("retention", getEnv("RETENTION_PERIOD", "7d"), "Data retention period (e.g., 24h, 7d, 30d)")
	cleanupInterval := flag.String("cleanup-interval", getEnv("CLEANUP_INTERVAL", "1h"), "Cleanup interval (e.g., 30m, 1h, 6h)")
	enableRetention := flag.Bool("enable-retention", getEnvBool("ENABLE_RETENTION", true), "Enable automatic data cleanup")
	enableAuth := flag.Bool("enable-auth", getEnvBool("ENABLE_AUTH", false), "Enable authentication")
	authUsers := flag.String("auth-users", getEnv("AUTH_USERS", ""), "Comma-separated list of username:password pairs (e.g., admin:password123,user:pass456)")
	flag.Parse()

	fmt.Println("Syslog Visualizer starting...")

	retentionCfg, err := parseRetentionConfig(*retentionPeriod, *cleanupInterval, *enableRetention)
	if err != nil {
		log.Fatalf("Failed to parse retention configuration: %v", err)
	}

	if retentionCfg.Enabled {
		log.Printf("Data retention enabled: keeping logs for %v, cleanup every %v",
			retentionCfg.RetentionPeriod, retentionCfg.CleanupInterval)
	} else {
		log.Println("WARNING: Data retention disabled: logs will be kept indefinitely")
	}

	authManager := auth.NewAuthManager(*enableAuth)

	if *enableAuth {
		if *authUsers == "" {
			log.Fatal("ERROR: Authentication enabled but no users configured. Use -auth-users flag or AUTH_USERS env variable")
		}

		userPairs := strings.Split(*authUsers, ",")
		for _, pair := range userPairs {
			parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(parts) != 2 {
				log.Fatalf("ERROR: Invalid user format: %s (expected username:password)", pair)
			}

			username := strings.TrimSpace(parts[0])
			password := strings.TrimSpace(parts[1])

			if err := authManager.AddUser(username, password); err != nil {
				log.Fatalf("ERROR: Failed to add user %s: %v", username, err)
			}

			apiToken, _ := authManager.GetAPIToken(username)
			log.Printf("User created: %s (API Token: %s)", username, apiToken)
		}

		log.Println("Authentication enabled")
		go startSessionCleanup(authManager)
	} else {
		log.Println("WARNING: Authentication disabled: API is publicly accessible")
	}

	dbPath := getEnv("DB_PATH", "/data/syslog.db")
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()
	log.Printf("Database initialized: %s", dbPath)

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

	collectorCfg := collector.Config{
		Address:       ":514",
		Protocol:      "udp",
		FramingMethod: framing.NonTransparent,
		Handler:       handler,
	}

	col, err := collector.New(collectorCfg)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/auth/login", handleLogin(authManager))
	mux.HandleFunc("/api/auth/logout", handleLogout(authManager))

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/syslogs", handleGetSyslogs(store))
	protectedMux.HandleFunc("/api/filter-options", handleGetFilterOptions(store))
	protectedMux.HandleFunc("/api/timeline", handleGetTimeline(store))

	mux.Handle("/api/syslogs", authManager.Middleware(protectedMux))
	mux.Handle("/api/filter-options", authManager.Middleware(protectedMux))
	mux.Handle("/api/timeline", authManager.Middleware(protectedMux))

	apiHandler := enableCORS(mux)

	apiPort := ":8080"
	apiServer := &http.Server{
		Addr:    apiPort,
		Handler: apiHandler,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	cleanupDoneChan := make(chan struct{})
	if retentionCfg.Enabled {
		go startDataRetentionCleanup(store, retentionCfg, cleanupDoneChan)
	}

	collectorErrChan := make(chan error, 1)
	go func() {
		log.Println("Starting syslog collector...")
		if err := col.Start(); err != nil {
			collectorErrChan <- fmt.Errorf("collector error: %w", err)
		}
	}()

	apiErrChan := make(chan error, 1)
	go func() {
		log.Printf("Starting API server on %s", apiPort)
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			apiErrChan <- fmt.Errorf("API server error: %w", err)
		}
	}()

	log.Println("Syslog Visualizer is running")
	log.Printf("  - Collector listening on :514 (UDP)")
	log.Printf("  - API server listening on %s", apiPort)
	log.Println("Press Ctrl+C to stop")

	select {
	case <-sigChan:
		log.Println("Shutdown signal received")
	case err := <-collectorErrChan:
		log.Printf("Collector error: %v", err)
	case err := <-apiErrChan:
		log.Printf("API server error: %v", err)
	}

	log.Println("Shutting down...")

	if retentionCfg.Enabled {
		close(cleanupDoneChan)
	}

	if err := col.Stop(); err != nil {
		log.Printf("Error stopping collector: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}

	log.Println("Shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func parseRetentionConfig(retentionStr, cleanupStr string, enabled bool) (*RetentionConfig, error) {
	retention, err := parseDuration(retentionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid retention period: %w", err)
	}

	cleanup, err := parseDuration(cleanupStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cleanup interval: %w", err)
	}

	return &RetentionConfig{
		RetentionPeriod: retention,
		CleanupInterval: cleanup,
		Enabled:         enabled,
	}, nil
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func startDataRetentionCleanup(store storage.Storage, cfg *RetentionConfig, done <-chan struct{}) {
	ticker := time.NewTicker(cfg.CleanupInterval)
	defer ticker.Stop()

	runCleanup(store, cfg.RetentionPeriod)

	for {
		select {
		case <-ticker.C:
			runCleanup(store, cfg.RetentionPeriod)
		case <-done:
			log.Println("Data retention cleanup stopped")
			return
		}
	}
}

func runCleanup(store storage.Storage, retentionPeriod time.Duration) {
	deleted, err := store.DeleteOlderThan(retentionPeriod)
	if err != nil {
		log.Printf("Error during cleanup: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("Cleaned up %d old messages (older than %v)", deleted, retentionPeriod)
	}
}

func startSessionCleanup(authManager *auth.AuthManager) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		authManager.CleanupExpiredSessions()
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleGetSyslogs(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		queryParams := r.URL.Query()
		filters := storage.QueryFilters{
			Limit: 100,
		}

		if limitStr := queryParams.Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				filters.Limit = limit
			}
		}

		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				filters.Offset = offset
			}
		}

		if severitiesStr := queryParams.Get("severities"); severitiesStr != "" {
			severities := parseIntSlice(severitiesStr)
			if len(severities) > 0 {
				filters.Severities = severities
			}
		}

		if facilitiesStr := queryParams.Get("facilities"); facilitiesStr != "" {
			facilities := parseIntSlice(facilitiesStr)
			if len(facilities) > 0 {
				filters.Facilities = facilities
			}
		}

		if hostname := queryParams.Get("hostname"); hostname != "" {
			filters.Hostname = hostname
		}

		if hostnamesStr := queryParams.Get("hostnames"); hostnamesStr != "" {
			hostnames := parseStringSlice(hostnamesStr)
			if len(hostnames) > 0 {
				filters.Hostnames = hostnames
			}
		}

		if tag := queryParams.Get("tag"); tag != "" {
			filters.Tag = tag
		}

		if search := queryParams.Get("search"); search != "" {
			filters.Search = search
		}

		if startTimeStr := queryParams.Get("start_time"); startTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				filters.StartTime = startTime
			}
		}

		if endTimeStr := queryParams.Get("end_time"); endTimeStr != "" {
			if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				filters.EndTime = endTime
			}
		}

		messages, totalCount, err := store.QueryWithCount(filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"data":  messages,
			"total": totalCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func parseIntSlice(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if i, err := strconv.Atoi(part); err == nil {
			result = append(result, i)
		}
	}
	return result
}

func parseStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func handleGetFilterOptions(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		options, err := store.GetFilterOptions()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(options)
	}
}

type TimeSlot struct {
	Timestamp      time.Time         `json:"timestamp"`
	SeverityCounts map[int]int       `json:"severity_counts"`
	Total          int               `json:"total"`
}

func handleGetTimeline(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		queryParams := r.URL.Query()

		// Get all messages (no pagination)
		filters := storage.QueryFilters{
			Limit: 100000, // Large limit to get all messages
		}

		// Apply optional filters
		if severitiesStr := queryParams.Get("severities"); severitiesStr != "" {
			severities := parseIntSlice(severitiesStr)
			if len(severities) > 0 {
				filters.Severities = severities
			}
		}

		if facilitiesStr := queryParams.Get("facilities"); facilitiesStr != "" {
			facilities := parseIntSlice(facilitiesStr)
			if len(facilities) > 0 {
				filters.Facilities = facilities
			}
		}

		if hostnamesStr := queryParams.Get("hostnames"); hostnamesStr != "" {
			hostnames := parseStringSlice(hostnamesStr)
			if len(hostnames) > 0 {
				filters.Hostnames = hostnames
			}
		}

		messages, err := store.Query(filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(messages) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]TimeSlot{})
			return
		}

		// Find time range
		var oldestTime, newestTime time.Time
		for i, msg := range messages {
			if i == 0 {
				oldestTime = msg.Timestamp
				newestTime = msg.Timestamp
			} else {
				if msg.Timestamp.Before(oldestTime) {
					oldestTime = msg.Timestamp
				}
				if msg.Timestamp.After(newestTime) {
					newestTime = msg.Timestamp
				}
			}
		}

		// Use now if newer than newest message
		now := time.Now()
		if now.After(newestTime) {
			newestTime = now
		}

		totalDuration := newestTime.Sub(oldestTime)

		// Calculate slot duration based on time range
		var slotDuration time.Duration
		var numSlots int

		if totalDuration <= 10*time.Minute {
			slotDuration = 30 * time.Second
		} else if totalDuration <= time.Hour {
			slotDuration = 2 * time.Minute
		} else if totalDuration <= 24*time.Hour {
			slotDuration = 30 * time.Minute
		} else {
			slotDuration = 2 * time.Hour
		}

		numSlots = int(totalDuration / slotDuration)
		if numSlots > 60 {
			numSlots = 60
			slotDuration = totalDuration / 60
		}
		if numSlots == 0 {
			numSlots = 1
		}

		// Create time slots
		slots := make([]TimeSlot, numSlots)
		for i := 0; i < numSlots; i++ {
			slotStart := oldestTime.Add(time.Duration(i) * slotDuration)
			slots[i] = TimeSlot{
				Timestamp:      slotStart,
				SeverityCounts: make(map[int]int),
				Total:          0,
			}
		}

		// Count messages in each slot
		for _, msg := range messages {
			slotIndex := int(msg.Timestamp.Sub(oldestTime) / slotDuration)
			if slotIndex >= numSlots {
				slotIndex = numSlots - 1
			}
			if slotIndex < 0 {
				slotIndex = 0
			}

			slots[slotIndex].SeverityCounts[msg.Severity]++
			slots[slotIndex].Total++
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slots)
	}
}

func handleLogin(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// If auth is disabled, return error
		if !authManager.IsEnabled() {
			http.Error(w, "Authentication is disabled", http.StatusNotImplemented)
			return
		}

		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify credentials
		if !authManager.VerifyPassword(credentials.Username, credentials.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create session
		sessionToken, err := authManager.CreateSession(credentials.Username)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Get API token
		apiToken, _ := authManager.GetAPIToken(credentials.Username)

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionToken,
			Path:     "/",
			MaxAge:   86400, // 24 hours
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Return success with API token
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "success",
			"username":  credentials.Username,
			"apiToken":  apiToken,
			"message":   "Login successful",
		})
	}
}

func handleLogout(authManager *auth.AuthManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get session cookie
		cookie, err := r.Cookie("session")
		if err == nil {
			authManager.DeleteSession(cookie.Value)
		}

		// Clear session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Logout successful",
		})
	}
}
