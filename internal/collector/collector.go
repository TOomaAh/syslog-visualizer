package collector

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"syslog-visualizer/internal/framing"
	"syslog-visualizer/internal/parser"
)

// MessageHandler is called for each received syslog message
type MessageHandler func(*parser.SyslogMessage) error

// Collector represents a syslog collector that listens for incoming messages
type Collector struct {
	address        string
	protocol       string
	framingMethod  framing.FramingMethod
	handler        MessageHandler
	udpConn        *net.UDPConn
	tcpListener    net.Listener
	ctx            context.Context
	cancel         context.CancelFunc
	maxMessageSize int
}

// Config holds the collector configuration
type Config struct {
	Address        string                // Listen address (e.g., "0.0.0.0:514" or ":514")
	Protocol       string                // "udp", "tcp", or "both"
	FramingMethod  framing.FramingMethod // For TCP: OctetCounting or NonTransparent
	Handler        MessageHandler        // Callback for each message
	MaxMessageSize int                   // Maximum message size in bytes (default 8192)
}

// New creates a new Collector instance
func New(cfg Config) (*Collector, error) {
	if cfg.Address == "" {
		cfg.Address = ":514"
	}
	if cfg.Protocol == "" {
		cfg.Protocol = "udp"
	}
	if cfg.MaxMessageSize == 0 {
		cfg.MaxMessageSize = 8192
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		address:        cfg.Address,
		protocol:       strings.ToLower(cfg.Protocol),
		framingMethod:  cfg.FramingMethod,
		handler:        cfg.Handler,
		ctx:            ctx,
		cancel:         cancel,
		maxMessageSize: cfg.MaxMessageSize,
	}, nil
}

// Start begins listening for syslog messages
func (c *Collector) Start() error {
	switch c.protocol {
	case "udp":
		return c.startUDP()
	case "tcp":
		return c.startTCP()
	case "both":
		// Start both UDP and TCP in separate goroutines
		errChan := make(chan error, 2)

		go func() {
			if err := c.startUDP(); err != nil {
				errChan <- fmt.Errorf("UDP error: %w", err)
			}
		}()

		go func() {
			if err := c.startTCP(); err != nil {
				errChan <- fmt.Errorf("TCP error: %w", err)
			}
		}()

		// Wait for either goroutine to fail or context cancellation
		select {
		case err := <-errChan:
			return err
		case <-c.ctx.Done():
			return nil
		}
	default:
		return fmt.Errorf("unsupported protocol: %s (use 'udp', 'tcp', or 'both')", c.protocol)
	}
}

// startUDP starts the UDP listener
func (c *Collector) startUDP() error {
	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	c.udpConn = conn

	log.Printf("UDP syslog collector listening on %s", c.address)

	// Read messages in a loop
	buffer := make([]byte, c.maxMessageSize)
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if c.ctx.Err() != nil {
					// Collector is stopping
					return nil
				}
				log.Printf("UDP read error: %v", err)
				continue
			}

			// Process the message
			raw := string(buffer[:n])
			c.processMessage(raw, remoteAddr.String())
		}
	}
}

// startTCP starts the TCP listener
func (c *Collector) startTCP() error {
	listener, err := net.Listen("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	c.tcpListener = listener

	log.Printf("TCP syslog collector listening on %s (framing: %v)", c.address, c.framingMethod)

	// Accept connections in a loop
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				if c.ctx.Err() != nil {
					// Collector is stopping
					return nil
				}
				log.Printf("TCP accept error: %v", err)
				continue
			}

			// Handle connection in a goroutine
			go c.handleTCPConnection(conn)
		}
	}
}

// handleTCPConnection handles a single TCP connection
func (c *Collector) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New TCP connection from %s", remoteAddr)

	reader := framing.NewReader(conn, c.framingMethod)
	reader.SetMaxSize(c.maxMessageSize)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			raw, err := reader.ReadMessage()
			if err != nil {
				if c.ctx.Err() != nil {
					// Collector is stopping
					return
				}
				log.Printf("TCP read error from %s: %v", remoteAddr, err)
				return
			}

			c.processMessage(raw, remoteAddr)
		}
	}
}

// processMessage parses and handles a raw syslog message
func (c *Collector) processMessage(raw string, remoteAddr string) {
	// Parse the message
	msg, err := parser.Parse(raw)
	if err != nil {
		log.Printf("Failed to parse message from %s: %v (raw: %q)", remoteAddr, err, raw)
		return
	}

	// Call the handler if one is configured
	if c.handler != nil {
		if err := c.handler(msg); err != nil {
			log.Printf("Handler error for message from %s: %v", remoteAddr, err)
		}
	}
}

// Stop gracefully stops the collector
func (c *Collector) Stop() error {
	log.Println("Stopping syslog collector...")

	// Cancel context to stop all goroutines
	c.cancel()

	// Close UDP connection
	if c.udpConn != nil {
		if err := c.udpConn.Close(); err != nil {
			return fmt.Errorf("failed to close UDP connection: %w", err)
		}
	}

	// Close TCP listener
	if c.tcpListener != nil {
		if err := c.tcpListener.Close(); err != nil {
			return fmt.Errorf("failed to close TCP listener: %w", err)
		}
	}

	log.Println("Syslog collector stopped")
	return nil
}
