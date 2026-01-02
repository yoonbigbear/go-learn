package main

import (
	"fmt"
	"log"
	"net"
	"time"

	sdk "agones.dev/agones/sdks/go"
)

func main() {
	s, err := sdk.NewSDK()
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenPacket("udp", ":7654")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = s.Ready()
	if err != nil {
		panic(err)
	}
	log.Println("Game server is ready to accept connections")

	go func() {
		tick := time.Tick(2 * time.Second)
		for range tick {
			s.Health()
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}

		msg := string(buf[:n])
		fmt.Printf("받은 메시지 : %s from %s\n", msg, addr)

		conn.WriteTo([]byte("Server: "+msg), addr)
	}
}
