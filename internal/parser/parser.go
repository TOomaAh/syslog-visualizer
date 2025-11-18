package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syslog-visualizer/pkg/syslog"
	"time"
)

// SyslogMessage represents a parsed syslog message
type SyslogMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	Facility  int       `json:"facility"`
	Severity  int       `json:"severity"`
	Tag       string    `json:"tag"`
	Message   string    `json:"message"`
	Raw       string    `json:"raw"`
	PID       string    `json:"pid,omitempty"`       // Process ID (RFC 3164)
	AppName   string    `json:"appName,omitempty"`   // Application name (RFC 5424)
	ProcID    string    `json:"procID,omitempty"`    // Process ID (RFC 5424)
	MsgID     string    `json:"msgID,omitempty"`     // Message ID (RFC 5424)
}

// FacilityName returns the human-readable name for the facility
func (m *SyslogMessage) FacilityName() string {
	return syslog.FacilityName(m.Facility)
}

// SeverityName returns the human-readable name for the severity
func (m *SyslogMessage) SeverityName() string {
	return syslog.SeverityName(m.Severity)
}

// Priority returns the calculated priority value (Facility * 8 + Severity)
func (m *SyslogMessage) Priority() int {
	return m.Facility*8 + m.Severity
}

// RFC 3164 timestamp formats (BSD syslog)
var rfc3164TimeFormats = []string{
	"Jan _2 15:04:05",
	"Jan 02 15:04:05",
}

// Parse parses a raw syslog message according to RFC 3164 or RFC 5424
// Auto-detects the format based on the message structure
func Parse(raw string) (*SyslogMessage, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty syslog message")
	}

	// Try to detect format
	// RFC 5424 has format: <PRI>VERSION where VERSION is a digit
	// RFC 3164 has format: <PRI>TIMESTAMP
	if isRFC5424(raw) {
		return ParseRFC5424(raw)
	}

	return ParseRFC3164(raw)
}

// isRFC5424 detects if the message is in RFC 5424 format
func isRFC5424(raw string) bool {
	// RFC 5424 starts with <PRI>VERSION where VERSION is typically "1"
	// Pattern: <digits>digits (after closing >)
	re := regexp.MustCompile(`^<\d+>(\d+)\s`)
	return re.MatchString(raw)
}

// ParseRFC3164 parses a syslog message in RFC 3164 format
// Format: <PRI>TIMESTAMP HOSTNAME TAG[PID]: MESSAGE
// Example: <34>Oct 11 22:14:15 mymachine su[1234]: 'su root' failed
func ParseRFC3164(raw string) (*SyslogMessage, error) {
	msg := &SyslogMessage{Raw: raw}

	priEnd := strings.Index(raw, ">")
	if priEnd == -1 || raw[0] != '<' {
		return nil, fmt.Errorf("invalid RFC 3164 format: missing priority")
	}

	priStr := raw[1:priEnd]
	pri, err := strconv.Atoi(priStr)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}

	msg.Facility = pri / 8
	msg.Severity = pri % 8

	rest := strings.TrimSpace(raw[priEnd+1:])

	// Extract timestamp (e.g., "Oct 11 22:14:05")
	// BSD syslog timestamp is 15 or 16 characters
	if len(rest) < 16 {
		return nil, fmt.Errorf("invalid RFC 3164 format: message too short")
	}

	timestampStr := rest[:15]
	var timestamp time.Time
	var parseErr error

	for _, format := range rfc3164TimeFormats {
		// Parse in local timezone since RFC 3164 doesn't include timezone info
		// and most syslog servers send local time
		timestamp, parseErr = time.ParseInLocation(format, timestampStr, time.Local)
		if parseErr == nil {
			// Add current year since BSD syslog doesn't include it
			now := time.Now()
			timestamp = time.Date(
				now.Year(),
				timestamp.Month(),
				timestamp.Day(),
				timestamp.Hour(),
				timestamp.Minute(),
				timestamp.Second(),
				timestamp.Nanosecond(),
				time.Local,
			)

			// If timestamp is more than 24 hours in the future, it's probably from last year
			// This handles year boundary (e.g., receiving Jan logs in December)
			if timestamp.After(now.Add(24 * time.Hour)) {
				timestamp = timestamp.AddDate(-1, 0, 0)
			}

			// Convert to UTC for consistent storage
			// This ensures the timestamp is stored as the actual moment in time
			timestamp = timestamp.UTC()

			break
		}
	}

	if parseErr != nil {
		timestamp = time.Now()
	}
	msg.Timestamp = timestamp

	rest = strings.TrimSpace(rest[15:])
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid RFC 3164 format: missing hostname or message")
	}

	msg.Hostname = parts[0]
	rest = parts[1]

	// Extract TAG[PID]: MESSAGE
	// TAG can be followed by [PID] and then : or just :
	tagRe := regexp.MustCompile(`^([^\s\[:]+)(?:\[(\d+)\])?:\s*(.*)$`)
	matches := tagRe.FindStringSubmatch(rest)

	if matches != nil {
		msg.Tag = matches[1]
		msg.PID = matches[2]
		msg.Message = matches[3]
	} else {
		// Fallback: split by first space if no colon found
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) >= 1 {
			msg.Tag = parts[0]
		}
		if len(parts) >= 2 {
			msg.Message = parts[1]
		}
	}

	return msg, nil
}

// ParseRFC5424 parses a syslog message in RFC 5424 format
// Format: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG
// Example: <34>1 2024-10-11T22:14:15.003Z mymachine su 1234 ID47 - 'su root' failed
func ParseRFC5424(raw string) (*SyslogMessage, error) {
	msg := &SyslogMessage{Raw: raw}

	priEnd := strings.Index(raw, ">")
	if priEnd == -1 || raw[0] != '<' {
		return nil, fmt.Errorf("invalid RFC 5424 format: missing priority")
	}

	priStr := raw[1:priEnd]
	pri, err := strconv.Atoi(priStr)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}

	msg.Facility = pri / 8
	msg.Severity = pri % 8

	rest := strings.TrimSpace(raw[priEnd+1:])
	fields := strings.SplitN(rest, " ", 7)

	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid RFC 5424 format: insufficient fields (got %d, need 7)", len(fields))
	}

	// VERSION (field 0) - skip it, we just validate it's a number
	if _, err := strconv.Atoi(fields[0]); err != nil {
		return nil, fmt.Errorf("invalid version field: %w", err)
	}

	if fields[1] != "-" {
		timestamp, err := time.Parse(time.RFC3339, fields[1])
		if err != nil {
			timestamp, err = time.Parse("2006-01-02T15:04:05.999999Z07:00", fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp: %w", err)
			}
		}
		msg.Timestamp = timestamp
	} else {
		msg.Timestamp = time.Now()
	}

	if fields[2] != "-" {
		msg.Hostname = fields[2]
	}

	if fields[3] != "-" {
		msg.AppName = fields[3]
		msg.Tag = fields[3]
	}

	if fields[4] != "-" {
		msg.ProcID = fields[4]
		msg.PID = fields[4]
	}

	if fields[5] != "-" {
		msg.MsgID = fields[5]
	}

	// STRUCTURED-DATA and MSG (fields 6)
	// For now, we'll treat everything after MSGID as the message
	// A full implementation would parse structured data
	remainder := fields[6]

	if strings.HasPrefix(remainder, "[") {
		sdEnd := findStructuredDataEnd(remainder)
		if sdEnd > 0 {
			if sdEnd < len(remainder) {
				msg.Message = strings.TrimSpace(remainder[sdEnd:])
			}
		} else {
			msg.Message = remainder
		}
	} else if remainder == "-" {
		msg.Message = ""
	} else {
		if strings.HasPrefix(remainder, "- ") {
			msg.Message = remainder[2:]
		} else {
			msg.Message = remainder
		}
	}

	return msg, nil
}

// findStructuredDataEnd finds the end of structured data section
// Structured data format: [id key="value" ...] or multiple [...]
func findStructuredDataEnd(s string) int {
	depth := 0
	escaped := false

	for i, ch := range s {
		if escaped {
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			escaped = true
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}

	return -1
}
