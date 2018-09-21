package main

import (
	"fmt"
	"github.com/desertbit/orbit/sample/api"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// streamRawRoutine is a showcase of the client side implementation of streaming on a
// raw net.Conn without using any helpers.
func streamRawRoutine(stream net.Conn, wg *sync.WaitGroup) {
	defer func() {
		// For better output readability.
		fmt.Println("---------------------------")
		stream.Close()
		wg.Done()
	}()

	var (
		n   int
		err error
		buf = make([]byte, 256)
	)

	for {
		// Set a read timeout.
		err = stream.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Printf("error reading from stream '%s': could not set read deadline", api.ChannelIDRaw)
			return
		}

		// Read from the stream.
		n, err = stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("error reading from stream '%s': %v", api.ChannelIDRaw, err)
			}
			return
		}
		if n == 0 {
			log.Printf("error reading from stream '%s': no data read", api.ChannelIDRaw)
			return
		}

		// Write the data to the terminal.
		fmt.Println(string(buf[:n]))
	}
}
