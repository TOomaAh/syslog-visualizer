package framing

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestOctetCountingReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "Valid octet counting message",
			input: "24 <34>Oct 11 22:14:15 test",
			want:  "<34>Oct 11 22:14:15 test",
		},
		{
			name:  "Multiple messages",
			input: "11 message one15 message two abc",
			want:  "message one",
		},
		{
			name:    "Invalid length - not a number",
			input:   "abc <34>test",
			wantErr: true,
		},
		{
			name:    "Invalid length - negative",
			input:   "-5 test",
			wantErr: true,
		},
		{
			name:    "Length mismatch - too short",
			input:   "100 short",
			wantErr: true,
		},
		{
			name:  "Zero-length after valid message",
			input: "5 hello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(strings.NewReader(tt.input), OctetCounting)
			got, err := reader.ReadMessage()

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ReadMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNonTransparentReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "LF-delimited message",
			input: "<34>Oct 11 22:14:15 test\n",
			want:  "<34>Oct 11 22:14:15 test",
		},
		{
			name:  "Multiple LF-delimited messages",
			input: "message one\nmessage two\n",
			want:  "message one",
		},
		{
			name:  "CRLF-delimited message",
			input: "<34>Oct 11 22:14:15 test\r\n",
			want:  "<34>Oct 11 22:14:15 test",
		},
		{
			name:  "NUL-delimited message",
			input: "<34>Oct 11 22:14:15 test\x00",
			want:  "<34>Oct 11 22:14:15 test",
		},
		{
			name:  "Message without delimiter at EOF",
			input: "<34>Oct 11 22:14:15 test",
			want:  "<34>Oct 11 22:14:15 test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(strings.NewReader(tt.input), NonTransparent)
			got, err := reader.ReadMessage()

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ReadMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMultipleMessages(t *testing.T) {
	t.Run("Octet counting - multiple messages", func(t *testing.T) {
		input := "5 hello10 world test15 another message"
		reader := NewReader(strings.NewReader(input), OctetCounting)

		// First message
		msg1, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("First ReadMessage() error = %v", err)
		}
		if msg1 != "hello" {
			t.Errorf("First message = %q, want %q", msg1, "hello")
		}

		// Second message
		msg2, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("Second ReadMessage() error = %v", err)
		}
		if msg2 != "world test" {
			t.Errorf("Second message = %q, want %q", msg2, "world test")
		}

		// Third message
		msg3, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("Third ReadMessage() error = %v", err)
		}
		if msg3 != "another message" {
			t.Errorf("Third message = %q, want %q", msg3, "another message")
		}
	})

	t.Run("Non-transparent - multiple messages", func(t *testing.T) {
		input := "hello\nworld\nanother message\n"
		reader := NewReader(strings.NewReader(input), NonTransparent)

		// First message
		msg1, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("First ReadMessage() error = %v", err)
		}
		if msg1 != "hello" {
			t.Errorf("First message = %q, want %q", msg1, "hello")
		}

		// Second message
		msg2, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("Second ReadMessage() error = %v", err)
		}
		if msg2 != "world" {
			t.Errorf("Second message = %q, want %q", msg2, "world")
		}

		// Third message
		msg3, err := reader.ReadMessage()
		if err != nil {
			t.Fatalf("Third ReadMessage() error = %v", err)
		}
		if msg3 != "another message" {
			t.Errorf("Third message = %q, want %q", msg3, "another message")
		}
	})
}

func TestAutoDetectFraming(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  FramingMethod
	}{
		{
			name:  "Detect octet counting",
			input: "24 <34>Oct 11 22:14:15 test",
			want:  OctetCounting,
		},
		{
			name:  "Detect octet counting - large number",
			input: "1024 <34>message...",
			want:  OctetCounting,
		},
		{
			name:  "Detect non-transparent - starts with <",
			input: "<34>Oct 11 22:14:15 test\n",
			want:  NonTransparent,
		},
		{
			name:  "Detect non-transparent - no digits",
			input: "some message\n",
			want:  NonTransparent,
		},
		{
			name:  "Edge case - digit but no space",
			input: "123abc",
			want:  NonTransparent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got, err := AutoDetectFraming(reader)
			if err != nil {
				t.Fatalf("AutoDetectFraming() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("AutoDetectFraming() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxSize(t *testing.T) {
	t.Run("Octet counting - exceeds max size", func(t *testing.T) {
		input := "100 " + strings.Repeat("x", 100)
		reader := NewReader(strings.NewReader(input), OctetCounting)
		reader.SetMaxSize(50)

		_, err := reader.ReadMessage()
		if err == nil {
			t.Error("Expected error for message exceeding max size")
		}
	})

	t.Run("Non-transparent - exceeds max size", func(t *testing.T) {
		input := strings.Repeat("x", 100) + "\n"
		reader := NewReader(strings.NewReader(input), NonTransparent)
		reader.SetMaxSize(50)

		_, err := reader.ReadMessage()
		if err == nil {
			t.Error("Expected error for message exceeding max size")
		}
	})
}

func TestOctetCountingWriter(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "Simple message",
			message: "hello",
			want:    "5 hello",
		},
		{
			name:    "Syslog message",
			message: "<34>Oct 11 22:14:15 test",
			want:    "24 <34>Oct 11 22:14:15 test",
		},
		{
			name:    "Empty message",
			message: "",
			want:    "0 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf, OctetCounting)

			err := writer.WriteMessage(tt.message)
			if err != nil {
				t.Fatalf("WriteMessage() error = %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("WriteMessage() wrote %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNonTransparentWriter(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "Simple message",
			message: "hello",
			want:    "hello\n",
		},
		{
			name:    "Syslog message",
			message: "<34>Oct 11 22:14:15 test",
			want:    "<34>Oct 11 22:14:15 test\n",
		},
		{
			name:    "Empty message",
			message: "",
			want:    "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf, NonTransparent)

			err := writer.WriteMessage(tt.message)
			if err != nil {
				t.Fatalf("WriteMessage() error = %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("WriteMessage() wrote %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	messages := []string{
		"<34>Oct 11 22:14:15 mymachine su: test",
		"<13>Feb  5 17:32:18 10.0.0.99 myapp: message",
		"hello world",
	}

	t.Run("Octet counting round trip", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewWriter(&buf, OctetCounting)

		// Write messages
		for _, msg := range messages {
			if err := writer.WriteMessage(msg); err != nil {
				t.Fatalf("WriteMessage() error = %v", err)
			}
		}

		// Read messages back
		reader := NewReader(&buf, OctetCounting)
		for i, want := range messages {
			got, err := reader.ReadMessage()
			if err != nil {
				t.Fatalf("ReadMessage() [%d] error = %v", i, err)
			}
			if got != want {
				t.Errorf("ReadMessage() [%d] = %q, want %q", i, got, want)
			}
		}
	})

	t.Run("Non-transparent round trip", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewWriter(&buf, NonTransparent)

		// Write messages
		for _, msg := range messages {
			if err := writer.WriteMessage(msg); err != nil {
				t.Fatalf("WriteMessage() error = %v", err)
			}
		}

		// Read messages back
		reader := NewReader(&buf, NonTransparent)
		for i, want := range messages {
			got, err := reader.ReadMessage()
			if err != nil {
				t.Fatalf("ReadMessage() [%d] error = %v", i, err)
			}
			if got != want {
				t.Errorf("ReadMessage() [%d] = %q, want %q", i, got, want)
			}
		}
	})
}
