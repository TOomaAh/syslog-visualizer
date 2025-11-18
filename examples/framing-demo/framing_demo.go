package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"syslog-visualizer/internal/framing"
)

func main() {
	fmt.Println("TCP Framing Demo")
	fmt.Println("================")
	fmt.Println()

	// Example syslog messages
	messages := []string{
		"<34>Oct 11 22:14:15 mymachine su[1234]: 'su root' failed",
		"<13>Feb  5 17:32:18 10.0.0.99 myapp: Application started",
		"<86>Dec  1 08:30:00 server01 kernel: Out of memory",
	}

	// Demo 1: Non-Transparent Framing (LF-delimited)
	fmt.Println("Demo 1: Non-Transparent Framing (LF-delimited)")
	fmt.Println("-----------------------------------------------")
	demoFraming(messages, framing.NonTransparent)

	time.Sleep(1 * time.Second)

	// Demo 2: Octet Counting Framing
	fmt.Println("\nDemo 2: Octet Counting Framing")
	fmt.Println("-------------------------------")
	demoFraming(messages, framing.OctetCounting)
}

func demoFraming(messages []string, method framing.FramingMethod) {
	// Start a simple TCP server in a goroutine
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	fmt.Printf("Server listening on %s\n", addr)

	// Server goroutine
	serverDone := make(chan bool)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			return
		}
		defer conn.Close()

		fmt.Println("\nServer: Connection accepted, reading messages...")
		reader := framing.NewReader(conn, method)

		for i := 0; i < len(messages); i++ {
			msg, err := reader.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			fmt.Printf("Server received: %q\n", msg)
		}
		serverDone <- true
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Client: Connect and send messages
	fmt.Println("\nClient: Connecting and sending messages...")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	writer := framing.NewWriter(conn, method)

	for _, msg := range messages {
		fmt.Printf("Client sending: %q\n", msg)
		if err := writer.WriteMessage(msg); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}

	// Wait for server to finish
	<-serverDone
	fmt.Println("Demo completed successfully!")
}

func init() {
	log.SetFlags(0)
}
