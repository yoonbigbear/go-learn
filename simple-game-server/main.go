package main

import (
	"fmt"
	"log"
	"net"
	"time"

	sdk "agones.dev/agones/sdks/go"
)

func main() {
	log.Print("Creating SDK instance")
	agonesSdk, err := sdk.NewSDK()
	if err != nil {
		panic(err)
	}

	log.Print("Listening on UDP :7654")
	conn, err := net.ListenPacket("udp", ":7654")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Print("SDK Get Ready")
	err = agonesSdk.Ready()
	if err != nil {
		panic(err)
	}
	log.Println("Game server is ready to accept connections")

	go func() {
		tick := time.Tick(2 * time.Second)
		for range tick {
			agonesSdk.Health()
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

		if msg == "shutdown" {
			time.Sleep(5 * time.Second)
			log.Println("Shutdown 명령어 수신, 서버 종료 중...")
			break
		}
	}
	agonesSdk.Shutdown() // Agones에게 서버를 종료하라고 알리는 코드
}
