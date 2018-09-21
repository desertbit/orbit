package main

import (
	"fmt"
	"github.com/desertbit/orbit/sample/api"
	"io"
	"log"
	"net"
	"time"
)

func streamOrbitRoutine(stream net.Conn) {
	defer stream.Close()

	var (
		n int
		err error
		buf = make([]byte, 256)
	)

	for {
		err = stream.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Printf("error reading from stream '%s': could not set read deadline", api.ChannelIDOrbit)
			return
		}

		// Read from the stream
		n, err = stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("error reading from stream '%s': %v", api.ChannelIDOrbit, err)
			}
			return
		}
		if n == 0 {
			log.Printf("error reading from stream '%s': no data read", api.ChannelIDOrbit)
			return
		}

		// Write the data to the terminal
		fmt.Println(string(buf[:n]))
	}
}
