package main

import (
	"fmt"
	"log"
	"syslog-visualizer/internal/parser"
)

func main() {
	fmt.Println("Syslog Parser Demo")
	fmt.Println("==================")
	fmt.Println()

	// Example RFC 3164 messages
	rfc3164Examples := []string{
		"<34>Oct 11 22:14:15 mymachine su[1234]: 'su root' failed for user on /dev/pts/8",
		"<13>Feb  5 17:32:18 10.0.0.99 myapp: This is a test message",
		"<86>Dec  1 08:30:00 server01 kernel: Out of memory",
	}

	// Example RFC 5424 messages
	rfc5424Examples := []string{
		"<34>1 2024-10-11T22:14:15.003Z mymachine su 1234 ID47 - 'su root' failed",
		"<165>1 2003-10-11T22:14:15.003Z mymachine - - - - An application event log entry",
		"<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - [exampleSDID@32473 iut=\"3\"] An event",
	}

	fmt.Println("RFC 3164 Examples:")
	fmt.Println("------------------")
	for i, raw := range rfc3164Examples {
		msg, err := parser.Parse(raw)
		if err != nil {
			log.Printf("Error parsing message %d: %v\n", i+1, err)
			continue
		}
		printMessage(msg)
	}

	fmt.Println("\nRFC 5424 Examples:")
	fmt.Println("------------------")
	for i, raw := range rfc5424Examples {
		msg, err := parser.Parse(raw)
		if err != nil {
			log.Printf("Error parsing message %d: %v\n", i+1, err)
			continue
		}
		printMessage(msg)
	}
}

func printMessage(msg *parser.SyslogMessage) {
	fmt.Printf("Timestamp:  %s\n", msg.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Hostname:   %s\n", msg.Hostname)
	fmt.Printf("Facility:   %d (%s)\n", msg.Facility, msg.FacilityName())
	fmt.Printf("Severity:   %d (%s)\n", msg.Severity, msg.SeverityName())
	fmt.Printf("Priority:   %d\n", msg.Priority())
	fmt.Printf("Tag:        %s\n", msg.Tag)
	if msg.PID != "" {
		fmt.Printf("PID:        %s\n", msg.PID)
	}
	if msg.MsgID != "" {
		fmt.Printf("MsgID:      %s\n", msg.MsgID)
	}
	fmt.Printf("Message:    %s\n", msg.Message)
	fmt.Println()
}
