package parser

import (
	"testing"
	"time"
)

func TestParseRFC3164(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected SyslogMessage
	}{
		{
			name:  "Valid RFC 3164 with PID",
			input: "<34>Oct 11 22:14:15 mymachine su[1234]: 'su root' failed for user on /dev/pts/8",
			expected: SyslogMessage{
				Facility: 4,
				Severity: 2,
				Hostname: "mymachine",
				Tag:      "su",
				PID:      "1234",
				Message:  "'su root' failed for user on /dev/pts/8",
			},
		},
		{
			name:  "Valid RFC 3164 without PID",
			input: "<13>Feb  5 17:32:18 10.0.0.99 myapp: This is a test message",
			expected: SyslogMessage{
				Facility: 1,
				Severity: 5,
				Hostname: "10.0.0.99",
				Tag:      "myapp",
				Message:  "This is a test message",
			},
		},
		{
			name:  "RFC 3164 with single digit day",
			input: "<86>Dec  1 08:30:00 server01 kernel: Out of memory",
			expected: SyslogMessage{
				Facility: 10,
				Severity: 6,
				Hostname: "server01",
				Tag:      "kernel",
				Message:  "Out of memory",
			},
		},
		{
			name:    "Invalid format - missing priority",
			input:   "Oct 11 22:14:15 mymachine su: message",
			wantErr: true,
		},
		{
			name:    "Invalid format - empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:  "RFC 3164 without colon separator",
			input: "<14>Jan 15 10:20:30 host01 process some message here",
			expected: SyslogMessage{
				Facility: 1,
				Severity: 6,
				Hostname: "host01",
				Tag:      "process",
				Message:  "some message here",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRFC3164(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRFC3164() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Facility != tt.expected.Facility {
				t.Errorf("Facility = %v, want %v", got.Facility, tt.expected.Facility)
			}
			if got.Severity != tt.expected.Severity {
				t.Errorf("Severity = %v, want %v", got.Severity, tt.expected.Severity)
			}
			if got.Hostname != tt.expected.Hostname {
				t.Errorf("Hostname = %v, want %v", got.Hostname, tt.expected.Hostname)
			}
			if got.Tag != tt.expected.Tag {
				t.Errorf("Tag = %v, want %v", got.Tag, tt.expected.Tag)
			}
			if got.PID != tt.expected.PID {
				t.Errorf("PID = %v, want %v", got.PID, tt.expected.PID)
			}
			if got.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", got.Message, tt.expected.Message)
			}
		})
	}
}

func TestParseRFC5424(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected SyslogMessage
	}{
		{
			name:  "Valid RFC 5424 with all fields",
			input: "<34>1 2024-10-11T22:14:15.003Z mymachine su 1234 ID47 - 'su root' failed",
			expected: SyslogMessage{
				Facility: 4,
				Severity: 2,
				Hostname: "mymachine",
				AppName:  "su",
				Tag:      "su",
				ProcID:   "1234",
				PID:      "1234",
				MsgID:    "ID47",
				Message:  "'su root' failed",
			},
		},
		{
			name:  "RFC 5424 with nil values",
			input: "<165>1 2003-10-11T22:14:15.003Z mymachine - - - - An application event log entry",
			expected: SyslogMessage{
				Facility: 20,
				Severity: 5,
				Hostname: "mymachine",
				Message:  "An application event log entry",
			},
		},
		{
			name:  "RFC 5424 with structured data",
			input: "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - [exampleSDID@32473 iut=\"3\" eventSource=\"Application\"] An application event",
			expected: SyslogMessage{
				Facility: 20,
				Severity: 5,
				Hostname: "192.0.2.1",
				AppName:  "myproc",
				Tag:      "myproc",
				ProcID:   "8710",
				PID:      "8710",
				Message:  "An application event",
			},
		},
		{
			name:    "Invalid RFC 5424 - insufficient fields",
			input:   "<34>1 2024-10-11T22:14:15.003Z",
			wantErr: true,
		},
		{
			name:    "Invalid RFC 5424 - bad priority",
			input:   "<abc>1 2024-10-11T22:14:15.003Z host app - - - msg",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRFC5424(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRFC5424() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Facility != tt.expected.Facility {
				t.Errorf("Facility = %v, want %v", got.Facility, tt.expected.Facility)
			}
			if got.Severity != tt.expected.Severity {
				t.Errorf("Severity = %v, want %v", got.Severity, tt.expected.Severity)
			}
			if got.Hostname != tt.expected.Hostname {
				t.Errorf("Hostname = %v, want %v", got.Hostname, tt.expected.Hostname)
			}
			if got.AppName != tt.expected.AppName {
				t.Errorf("AppName = %v, want %v", got.AppName, tt.expected.AppName)
			}
			if got.ProcID != tt.expected.ProcID {
				t.Errorf("ProcID = %v, want %v", got.ProcID, tt.expected.ProcID)
			}
			if got.MsgID != tt.expected.MsgID {
				t.Errorf("MsgID = %v, want %v", got.MsgID, tt.expected.MsgID)
			}
			if got.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", got.Message, tt.expected.Message)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		isRFC5424 bool
	}{
		{
			name:      "Auto-detect RFC 3164",
			input:     "<34>Oct 11 22:14:15 mymachine su: test",
			isRFC5424: false,
		},
		{
			name:      "Auto-detect RFC 5424",
			input:     "<34>1 2024-10-11T22:14:15.003Z mymachine su - - - test",
			isRFC5424: true,
		},
		{
			name:    "Empty message",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Verify the message was parsed
			if got == nil {
				t.Error("Parse() returned nil message")
				return
			}

			// Basic validation
			if got.Raw != tt.input {
				t.Errorf("Raw = %v, want %v", got.Raw, tt.input)
			}
		})
	}
}

func TestIsRFC5424(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "RFC 5424 format",
			input: "<34>1 2024-10-11T22:14:15.003Z mymachine",
			want:  true,
		},
		{
			name:  "RFC 3164 format",
			input: "<34>Oct 11 22:14:15 mymachine",
			want:  false,
		},
		{
			name:  "RFC 5424 with version 2",
			input: "<34>2 2024-10-11T22:14:15.003Z mymachine",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRFC5424(tt.input); got != tt.want {
				t.Errorf("isRFC5424() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityCalculation(t *testing.T) {
	tests := []struct {
		priority int
		facility int
		severity int
	}{
		{34, 4, 2},   // auth.crit
		{13, 1, 5},   // user.notice
		{86, 10, 6},  // authpriv.info
		{165, 20, 5}, // local4.notice
		{0, 0, 0},    // kern.emerg
		{191, 23, 7}, // local7.debug
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			facility := tt.priority / 8
			severity := tt.priority % 8

			if facility != tt.facility {
				t.Errorf("Facility = %v, want %v (priority %v)", facility, tt.facility, tt.priority)
			}
			if severity != tt.severity {
				t.Errorf("Severity = %v, want %v (priority %v)", severity, tt.severity, tt.priority)
			}
		})
	}
}

func TestTimestampParsing(t *testing.T) {
	// Test that RFC 3164 timestamp gets current year added
	input := "<34>Oct 11 22:14:15 mymachine su: test"
	msg, err := ParseRFC3164(input)
	if err != nil {
		t.Fatalf("ParseRFC3164() error = %v", err)
	}

	currentYear := time.Now().Year()
	if msg.Timestamp.Year() != currentYear {
		t.Errorf("Timestamp year = %v, want %v", msg.Timestamp.Year(), currentYear)
	}

	// Verify time components
	if msg.Timestamp.Month() != time.October {
		t.Errorf("Timestamp month = %v, want October", msg.Timestamp.Month())
	}
	if msg.Timestamp.Day() != 11 {
		t.Errorf("Timestamp day = %v, want 11", msg.Timestamp.Day())
	}
	if msg.Timestamp.Hour() != 22 {
		t.Errorf("Timestamp hour = %v, want 22", msg.Timestamp.Hour())
	}
}

func TestHelperMethods(t *testing.T) {
	// Priority <34> = Facility 4 (auth) + Severity 2 (critical)
	input := "<34>Oct 11 22:14:15 mymachine su: test"
	msg, err := ParseRFC3164(input)
	if err != nil {
		t.Fatalf("ParseRFC3164() error = %v", err)
	}

	if msg.FacilityName() != "auth" {
		t.Errorf("FacilityName() = %v, want auth", msg.FacilityName())
	}

	if msg.SeverityName() != "critical" {
		t.Errorf("SeverityName() = %v, want critical", msg.SeverityName())
	}

	if msg.Priority() != 34 {
		t.Errorf("Priority() = %v, want 34", msg.Priority())
	}

	// Test another priority: <165> = Facility 20 (local4) + Severity 5 (notice)
	input2 := "<165>1 2003-10-11T22:14:15.003Z mymachine app - - - test"
	msg2, err := ParseRFC5424(input2)
	if err != nil {
		t.Fatalf("ParseRFC5424() error = %v", err)
	}

	if msg2.FacilityName() != "local4" {
		t.Errorf("FacilityName() = %v, want local4", msg2.FacilityName())
	}

	if msg2.SeverityName() != "notice" {
		t.Errorf("SeverityName() = %v, want notice", msg2.SeverityName())
	}

	if msg2.Priority() != 165 {
		t.Errorf("Priority() = %v, want 165", msg2.Priority())
	}
}
