package framing

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FramingMethod represents the TCP framing method used for syslog messages
type FramingMethod int

const (
	// OctetCounting uses message length prefix (RFC 6587)
	// Format: <length> <message>
	// Example: 25 <34>Oct 11 22:14:15 test
	OctetCounting FramingMethod = iota

	// NonTransparent uses delimiter-based framing (RFC 6587)
	// Format: <message><delimiter>
	// Delimiter is typically LF (\n) or NUL (\0)
	NonTransparent
)

// Reader reads syslog messages from a TCP stream with proper framing
type Reader struct {
	reader  *bufio.Reader
	method  FramingMethod
	maxSize int
}

// NewReader creates a new framing reader
func NewReader(r io.Reader, method FramingMethod) *Reader {
	return &Reader{
		reader:  bufio.NewReader(r),
		method:  method,
		maxSize: 8192, // Default max message size (8KB)
	}
}

// SetMaxSize sets the maximum message size in bytes
func (r *Reader) SetMaxSize(size int) {
	r.maxSize = size
}

// ReadMessage reads the next syslog message from the stream
func (r *Reader) ReadMessage() (string, error) {
	switch r.method {
	case OctetCounting:
		return r.readOctetCounting()
	case NonTransparent:
		return r.readNonTransparent()
	default:
		return "", fmt.Errorf("unknown framing method: %d", r.method)
	}
}

// readOctetCounting reads a message using octet counting framing
// Format: <length> <message>
// Example: 25 <34>Oct 11 22:14:15 test
func (r *Reader) readOctetCounting() (string, error) {
	// Read the length prefix (terminated by space)
	lengthStr, err := r.reader.ReadString(' ')
	if err != nil {
		return "", fmt.Errorf("failed to read length prefix: %w", err)
	}

	// Parse the length (remove the trailing space)
	lengthStr = strings.TrimSpace(lengthStr)
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid length prefix '%s': %w", lengthStr, err)
	}

	// Validate length
	if length <= 0 {
		return "", fmt.Errorf("invalid message length: %d", length)
	}
	if length > r.maxSize {
		return "", fmt.Errorf("message length %d exceeds maximum %d", length, r.maxSize)
	}

	// Read exactly 'length' bytes
	message := make([]byte, length)
	n, err := io.ReadFull(r.reader, message)
	if err != nil {
		return "", fmt.Errorf("failed to read message (expected %d bytes, got %d): %w", length, n, err)
	}

	return string(message), nil
}

// readNonTransparent reads a message using non-transparent framing
// Messages are delimited by LF (\n) or NUL (\0)
// Tries LF first, which is more common
func (r *Reader) readNonTransparent() (string, error) {
	// Read until newline (most common delimiter)
	message, err := r.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read message: %w", err)
	}

	// Remove the trailing delimiter
	message = strings.TrimRight(message, "\n\r\x00")

	// Validate size
	if len(message) > r.maxSize {
		return "", fmt.Errorf("message length %d exceeds maximum %d", len(message), r.maxSize)
	}

	return message, nil
}

// AutoDetectFraming attempts to detect the framing method by peeking at the stream
// Octet counting starts with digits followed by a space
// Non-transparent starts with '<' (the priority field)
func AutoDetectFraming(r *bufio.Reader) (FramingMethod, error) {
	// Peek at the first few bytes
	peek, err := r.Peek(10)
	if err != nil && err != io.EOF {
		return NonTransparent, fmt.Errorf("failed to peek at stream: %w", err)
	}

	if len(peek) == 0 {
		return NonTransparent, fmt.Errorf("empty stream")
	}

	// Check if it starts with digits (octet counting)
	// Format: <length> <message>
	for i, b := range peek {
		if b >= '0' && b <= '9' {
			continue
		}
		if b == ' ' && i > 0 {
			// Found digit(s) followed by space - likely octet counting
			return OctetCounting, nil
		}
		// Not a digit or space - break
		break
	}

	// Default to non-transparent framing (LF-delimited)
	return NonTransparent, nil
}

// Writer writes syslog messages to a TCP stream with proper framing
type Writer struct {
	writer io.Writer
	method FramingMethod
}

// NewWriter creates a new framing writer
func NewWriter(w io.Writer, method FramingMethod) *Writer {
	return &Writer{
		writer: w,
		method: method,
	}
}

// WriteMessage writes a syslog message with appropriate framing
func (w *Writer) WriteMessage(message string) error {
	switch w.method {
	case OctetCounting:
		return w.writeOctetCounting(message)
	case NonTransparent:
		return w.writeNonTransparent(message)
	default:
		return fmt.Errorf("unknown framing method: %d", w.method)
	}
}

// writeOctetCounting writes a message using octet counting framing
func (w *Writer) writeOctetCounting(message string) error {
	length := len(message)
	frame := fmt.Sprintf("%d %s", length, message)
	_, err := w.writer.Write([]byte(frame))
	return err
}

// writeNonTransparent writes a message using non-transparent framing (LF delimiter)
func (w *Writer) writeNonTransparent(message string) error {
	frame := message + "\n"
	_, err := w.writer.Write([]byte(frame))
	return err
}
