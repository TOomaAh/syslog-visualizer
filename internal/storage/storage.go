package storage

import (
	"syslog-visualizer/internal/parser"
	"time"
)

// Storage defines the interface for storing syslog messages
type Storage interface {
	Store(msg *parser.SyslogMessage) error
	Query(filters QueryFilters) ([]*parser.SyslogMessage, error)
	QueryWithCount(filters QueryFilters) ([]*parser.SyslogMessage, int64, error)
	GetFilterOptions() (*FilterOptions, error)
	DeleteOlderThan(duration time.Duration) (int64, error)
	Close() error
}

// FilterOptions contains all unique values for filtering
type FilterOptions struct {
	Hostnames  []string `json:"hostnames"`
	Tags       []string `json:"tags"`
	Facilities []int    `json:"facilities"`
	Severities []int    `json:"severities"`
}

// QueryFilters defines filters for querying syslog messages
type QueryFilters struct {
	StartTime  time.Time
	EndTime    time.Time
	Hostname   string
	Hostnames  []string // Multiple hostnames
	Severity   *int     // Deprecated: use Severities
	Severities []int    // Multiple severities
	Facility   *int     // Deprecated: use Facilities
	Facilities []int    // Multiple facilities
	Tag        string   // Filter by tag
	Search     string   // Search term for message content
	Limit      int
	Offset     int
}

// MemoryStorage is an in-memory storage implementation
type MemoryStorage struct {
	messages []*parser.SyslogMessage
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		messages: make([]*parser.SyslogMessage, 0),
	}
}

// Store stores a syslog message in memory
func (s *MemoryStorage) Store(msg *parser.SyslogMessage) error {
	s.messages = append(s.messages, msg)
	return nil
}

// Query retrieves syslog messages based on filters
func (s *MemoryStorage) Query(filters QueryFilters) ([]*parser.SyslogMessage, error) {
	// TODO: Implement filtering logic
	return s.messages, nil
}

// QueryWithCount retrieves syslog messages with total count
func (s *MemoryStorage) QueryWithCount(filters QueryFilters) ([]*parser.SyslogMessage, int64, error) {
	messages, err := s.Query(filters)
	return messages, int64(len(messages)), err
}

// GetFilterOptions returns all unique values for filtering
func (s *MemoryStorage) GetFilterOptions() (*FilterOptions, error) {
	hostnamesMap := make(map[string]bool)
	tagsMap := make(map[string]bool)
	facilitiesMap := make(map[int]bool)
	severitiesMap := make(map[int]bool)

	for _, msg := range s.messages {
		hostnamesMap[msg.Hostname] = true
		if msg.Tag != "" {
			tagsMap[msg.Tag] = true
		}
		facilitiesMap[msg.Facility] = true
		severitiesMap[msg.Severity] = true
	}

	// Convert maps to sorted slices
	hostnames := make([]string, 0, len(hostnamesMap))
	for h := range hostnamesMap {
		hostnames = append(hostnames, h)
	}

	tags := make([]string, 0, len(tagsMap))
	for t := range tagsMap {
		tags = append(tags, t)
	}

	facilities := make([]int, 0, len(facilitiesMap))
	for f := range facilitiesMap {
		facilities = append(facilities, f)
	}

	severities := make([]int, 0, len(severitiesMap))
	for s := range severitiesMap {
		severities = append(severities, s)
	}

	return &FilterOptions{
		Hostnames:  hostnames,
		Tags:       tags,
		Facilities: facilities,
		Severities: severities,
	}, nil
}

// DeleteOlderThan deletes messages older than the specified duration
func (s *MemoryStorage) DeleteOlderThan(duration time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-duration)
	newMessages := make([]*parser.SyslogMessage, 0)
	deleted := int64(0)

	for _, msg := range s.messages {
		if msg.Timestamp.After(cutoffTime) {
			newMessages = append(newMessages, msg)
		} else {
			deleted++
		}
	}

	s.messages = newMessages
	return deleted, nil
}

// Close closes the storage (no-op for memory storage)
func (s *MemoryStorage) Close() error {
	return nil
}
