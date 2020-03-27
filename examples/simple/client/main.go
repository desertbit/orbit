package main

import (
	"context"
	"crypto/tls"
	"log"
	"sync"

	"github.com/desertbit/orbit/examples/simple/hello"
	"github.com/desertbit/orbit/pkg/client"
	olog "github.com/desertbit/orbit/pkg/hook/log"
	"github.com/desertbit/orbit/pkg/transport/quic"
)

func main() {
	tr, err := quic.NewTransport(&quic.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"orbit-simple-example"},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	c, err := hello.NewClient(&client.Options{
		Host:      "127.0.0.1:1122",
		Transport: tr,
		Hooks: client.Hooks{
			olog.ClientHook(),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := c.Test(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}()

	err = c.SayHi(context.Background(), &hello.SayHiArg{Name: "Wastl"})
	if err != nil {
		log.Fatalln(err)
	}

	wg.Wait()
}
