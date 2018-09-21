package main

import (
	"fmt"
	"github.com/desertbit/orbit/packet"
	"log"
	"net"
	"sync"
	"time"
)

const (
	maxPayloadSize = 10 * 1024 // 10KB
)

// streamPacketRoutine is a showcase of the client side implementation of streaming data
// by using the packet pkg.
func streamPacketRoutine(stream net.Conn, wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("---------------------------")
		stream.Close()
		wg.Done()
	}()

	// Wait a little bit to make output better readable.
	time.Sleep(1 * time.Second)

	// Set a write deadline.
	err := stream.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Printf("error setting write deadline with packet: %v", err)
		return
	}

	// Write the first message to the server.
	err = packet.Write(stream, []byte("hey server, ur mom gay"), maxPayloadSize)
	if err != nil {
		log.Printf("error writing to stream with packet: %v", err)
		return
	}

	// Read the witty response from the server.
	// Notice that we let packet set the read deadline on its own.
	data, err := packet.ReadTimeout(stream, nil, maxPayloadSize, 10*time.Second)
	if err != nil {
		log.Printf("error reading from stream with packet: %v", err)
		return
	}

	// Print the output
	fmt.Println(string(data))
}
