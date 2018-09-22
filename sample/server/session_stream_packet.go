package main

import (
	"fmt"
	"github.com/desertbit/orbit/packet"
	"net"
	"time"
)

const (
	maxPayloadSize = 10 * 1024 // KB
)

// handleStreamPacket is a showcase of the server side implementation of streaming data
// by using the packet pkg.
func handleStreamPacket(stream net.Conn) error {
	// Set the read deadline manually
	err := stream.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return fmt.Errorf("error setting read deadline with packet: %v", err)
	}

	// Read data from stream.
	// We give nil for the buffer, therefore packet will allocate a buffer that fits the
	// payload of the message.
	data, err := packet.Read(stream, nil, maxPayloadSize)
	if err != nil {
		return fmt.Errorf("error reading from stream with packet: %v", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("error reading from stream with packet: no data read")
	}

	// Output the data to the console.
	fmt.Println(string(data))

	// For better output readability.
	time.Sleep(1 * time.Second)

	// Write a witty response.
	// This time, we let packet set the timeout.
	err = packet.WriteTimeout(stream, []byte("no u!"), maxPayloadSize, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error writing to stream with packet: %v", err)
	}

	return nil
}
