package storage

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"syslog-visualizer/internal/parser"
)

// SyslogMessageModel is the GORM model for syslog messages
type SyslogMessageModel struct {
	ID        uint      `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"index;not null"`
	Hostname  string    `gorm:"index;not null"`
	Facility  int       `gorm:"index;not null"`
	Severity  int       `gorm:"index;not null"`
	Tag       string    `gorm:"type:text"`
	Message   string    `gorm:"type:text;not null"`
	Raw       string    `gorm:"type:text;not null"`
	PID       string    `gorm:"type:text"`
	AppName   string    `gorm:"type:text"`
	ProcID    string    `gorm:"type:text"`
	MsgID     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"index;autoCreateTime"`
}

// TableName overrides the table name
func (SyslogMessageModel) TableName() string {
	return "syslog_messages"
}

// SQLiteStorage is a SQLite-based storage implementation using GORM
type SQLiteStorage struct {
	db *gorm.DB
}

// NewSQLiteStorage creates a new SQLite storage with GORM
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	storage := &SQLiteStorage{db: db}

	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return storage, nil
}

// migrate runs GORM auto-migration
func (s *SQLiteStorage) migrate() error {
	return s.db.AutoMigrate(&SyslogMessageModel{})
}

// Store stores a syslog message in the database
func (s *SQLiteStorage) Store(msg *parser.SyslogMessage) error {
	model := &SyslogMessageModel{
		Timestamp: msg.Timestamp,
		Hostname:  msg.Hostname,
		Facility:  msg.Facility,
		Severity:  msg.Severity,
		Tag:       msg.Tag,
		Message:   msg.Message,
		Raw:       msg.Raw,
		PID:       msg.PID,
		AppName:   msg.AppName,
		ProcID:    msg.ProcID,
		MsgID:     msg.MsgID,
	}

	if err := s.db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	return nil
}

// Query retrieves syslog messages based on filters
func (s *SQLiteStorage) Query(filters QueryFilters) ([]*parser.SyslogMessage, error) {
	query := s.db.Model(&SyslogMessageModel{})

	if !filters.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", filters.StartTime)
	}

	if !filters.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", filters.EndTime)
	}

	// Hostname filters (support both single and multiple)
	if filters.Hostname != "" {
		query = query.Where("hostname = ?", filters.Hostname)
	}
	if len(filters.Hostnames) > 0 {
		query = query.Where("hostname IN ?", filters.Hostnames)
	}

	// Severity filters (support both single and multiple)
	if filters.Severity != nil {
		query = query.Where("severity = ?", *filters.Severity)
	}
	if len(filters.Severities) > 0 {
		query = query.Where("severity IN ?", filters.Severities)
	}

	// Facility filters (support both single and multiple)
	if filters.Facility != nil {
		query = query.Where("facility = ?", *filters.Facility)
	}
	if len(filters.Facilities) > 0 {
		query = query.Where("facility IN ?", filters.Facilities)
	}

	if filters.Tag != "" {
		query = query.Where("tag = ?", filters.Tag)
	}

	// Search filter (search in message, tag, and hostname)
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(message) LIKE ? OR LOWER(tag) LIKE ? OR LOWER(hostname) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	query = query.Order("timestamp DESC")

	limit := filters.Limit
	if limit <= 0 {
		limit = 1000 // Default limit to prevent huge result sets
	}
	query = query.Limit(limit)

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var models []SyslogMessageModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	messages := make([]*parser.SyslogMessage, len(models))
	for i, model := range models {
		messages[i] = &parser.SyslogMessage{
			ID:        model.ID,
			Timestamp: model.Timestamp,
			Hostname:  model.Hostname,
			Facility:  model.Facility,
			Severity:  model.Severity,
			Tag:       model.Tag,
			Message:   model.Message,
			Raw:       model.Raw,
			PID:       model.PID,
			AppName:   model.AppName,
			ProcID:    model.ProcID,
			MsgID:     model.MsgID,
		}
	}

	return messages, nil
}

// QueryWithCount retrieves syslog messages with total count based on filters
func (s *SQLiteStorage) QueryWithCount(filters QueryFilters) ([]*parser.SyslogMessage, int64, error) {
	countQuery := s.db.Model(&SyslogMessageModel{})
	dataQuery := s.db.Model(&SyslogMessageModel{})

	// Apply the same filters to both queries
	applyFilters := func(query *gorm.DB) *gorm.DB {
		if !filters.StartTime.IsZero() {
			query = query.Where("timestamp >= ?", filters.StartTime)
		}

		if !filters.EndTime.IsZero() {
			query = query.Where("timestamp <= ?", filters.EndTime)
		}

		// Hostname filters (support both single and multiple)
		if filters.Hostname != "" {
			query = query.Where("hostname = ?", filters.Hostname)
		}
		if len(filters.Hostnames) > 0 {
			query = query.Where("hostname IN ?", filters.Hostnames)
		}

		// Severity filters (support both single and multiple)
		if filters.Severity != nil {
			query = query.Where("severity = ?", *filters.Severity)
		}
		if len(filters.Severities) > 0 {
			query = query.Where("severity IN ?", filters.Severities)
		}

		// Facility filters (support both single and multiple)
		if filters.Facility != nil {
			query = query.Where("facility = ?", *filters.Facility)
		}
		if len(filters.Facilities) > 0 {
			query = query.Where("facility IN ?", filters.Facilities)
		}

		// Tag filter
		if filters.Tag != "" {
			query = query.Where("tag = ?", filters.Tag)
		}

		// Search filter (search in message, tag, and hostname)
		if filters.Search != "" {
			searchPattern := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(message) LIKE ? OR LOWER(tag) LIKE ? OR LOWER(hostname) LIKE ?",
				searchPattern, searchPattern, searchPattern)
		}

		return query
	}

	countQuery = applyFilters(countQuery)
	dataQuery = applyFilters(dataQuery)

	var totalCount int64
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	dataQuery = dataQuery.Order("timestamp DESC")

	limit := filters.Limit
	if limit <= 0 {
		limit = 1000 // Default limit to prevent huge result sets
	}
	dataQuery = dataQuery.Limit(limit)

	if filters.Offset > 0 {
		dataQuery = dataQuery.Offset(filters.Offset)
	}

	var models []SyslogMessageModel
	if err := dataQuery.Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query messages: %w", err)
	}

	messages := make([]*parser.SyslogMessage, len(models))
	for i, model := range models {
		messages[i] = &parser.SyslogMessage{
			ID:        model.ID,
			Timestamp: model.Timestamp,
			Hostname:  model.Hostname,
			Facility:  model.Facility,
			Severity:  model.Severity,
			Tag:       model.Tag,
			Message:   model.Message,
			Raw:       model.Raw,
			PID:       model.PID,
			AppName:   model.AppName,
			ProcID:    model.ProcID,
			MsgID:     model.MsgID,
		}
	}

	return messages, totalCount, nil
}

// GetFilterOptions returns all unique values for filtering
func (s *SQLiteStorage) GetFilterOptions() (*FilterOptions, error) {
	options := &FilterOptions{
		Hostnames:  make([]string, 0),
		Tags:       make([]string, 0),
		Facilities: make([]int, 0),
		Severities: make([]int, 0),
	}

	var hostnames []string
	if err := s.db.Model(&SyslogMessageModel{}).
		Distinct("hostname").
		Order("hostname ASC").
		Pluck("hostname", &hostnames).Error; err != nil {
		return nil, fmt.Errorf("failed to get hostnames: %w", err)
	}
	options.Hostnames = hostnames

	var tags []string
	if err := s.db.Model(&SyslogMessageModel{}).
		Distinct("tag").
		Where("tag != ?", "").
		Order("tag ASC").
		Pluck("tag", &tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	options.Tags = tags

	var facilities []int
	if err := s.db.Model(&SyslogMessageModel{}).
		Distinct("facility").
		Order("facility ASC").
		Pluck("facility", &facilities).Error; err != nil {
		return nil, fmt.Errorf("failed to get facilities: %w", err)
	}
	options.Facilities = facilities

	var severities []int
	if err := s.db.Model(&SyslogMessageModel{}).
		Distinct("severity").
		Order("severity ASC").
		Pluck("severity", &severities).Error; err != nil {
		return nil, fmt.Errorf("failed to get severities: %w", err)
	}
	options.Severities = severities

	return options, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Stats returns database statistics
func (s *SQLiteStorage) Stats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total message count
	var count int64
	if err := s.db.Model(&SyslogMessageModel{}).Count(&count).Error; err != nil {
		return nil, err
	}
	stats["total_messages"] = count

	type SeverityCount struct {
		Severity int
		Count    int64
	}
	var severityCounts []SeverityCount
	if err := s.db.Model(&SyslogMessageModel{}).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Order("severity").
		Scan(&severityCounts).Error; err != nil {
		return nil, err
	}

	severityMap := make(map[int]int64)
	for _, sc := range severityCounts {
		severityMap[sc.Severity] = sc.Count
	}
	stats["by_severity"] = severityMap

	// Database file size using raw SQL for PRAGMA
	sqlDB, err := s.db.DB()
	if err == nil {
		var pageCount, pageSize int64
		row := sqlDB.QueryRow("PRAGMA page_count")
		if row.Scan(&pageCount) == nil {
			row = sqlDB.QueryRow("PRAGMA page_size")
			if row.Scan(&pageSize) == nil {
				stats["db_size_bytes"] = pageCount * pageSize
			}
		}
	}

	return stats, nil
}

// DeleteOlderThan deletes messages older than the specified duration
func (s *SQLiteStorage) DeleteOlderThan(duration time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-duration)

	result := s.db.Where("timestamp < ?", cutoffTime).Delete(&SyslogMessageModel{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old messages: %w", result.Error)
	}

	rowsAffected := result.RowsAffected

	// Vacuum to reclaim space
	if rowsAffected > 0 {
		sqlDB, err := s.db.DB()
		if err != nil {
			return rowsAffected, fmt.Errorf("deleted %d rows but failed to get DB: %w", rowsAffected, err)
		}
		if _, err := sqlDB.Exec("VACUUM"); err != nil {
			return rowsAffected, fmt.Errorf("deleted %d rows but vacuum failed: %w", rowsAffected, err)
		}
	}

	return rowsAffected, nil
}

// SearchMessages searches for messages containing the search term
func (s *SQLiteStorage) SearchMessages(searchTerm string, limit int) ([]*parser.SyslogMessage, error) {
	if limit <= 0 {
		limit = 100
	}

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	var models []SyslogMessageModel
	err := s.db.Where("message LIKE ? OR tag LIKE ? OR hostname LIKE ?",
		searchPattern, searchPattern, searchPattern).
		Order("timestamp DESC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	messages := make([]*parser.SyslogMessage, len(models))
	for i, model := range models {
		messages[i] = &parser.SyslogMessage{
			ID:        model.ID,
			Timestamp: model.Timestamp,
			Hostname:  model.Hostname,
			Facility:  model.Facility,
			Severity:  model.Severity,
			Tag:       model.Tag,
			Message:   model.Message,
			Raw:       model.Raw,
			PID:       model.PID,
			AppName:   model.AppName,
			ProcID:    model.ProcID,
			MsgID:     model.MsgID,
		}
	}

	return messages, nil
}
