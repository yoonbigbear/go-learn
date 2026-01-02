package main

import (
	"fmt"
	"net"
)

func main() {
	// 1. IP와 포트 설정 (아래에서 확인한 값으로 바꾸세요!)
	// 예: 192.168.49.2:7345
	serverAddr := "192.168.49.2:7345"

	// 2. UDP 주소 해석
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Println("주소 에러:", err)
		return
	}

	// 3. 서버에 연결 (UDP라 실제 연결은 아님)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("연결 에러:", err)
		return
	}
	defer conn.Close()

	// 4. 메시지 보내기
	msg := []byte("Hello Game Server!")
	_, err = conn.Write(msg)
	if err != nil {
		fmt.Println("전송 에러:", err)
		return
	}
	fmt.Printf("보냄: %s\n", msg)

	// 5. 응답 받기 (대기)
	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("수신 에러:", err)
		return
	}
	fmt.Printf("받음: %s\n", string(buf[:n]))
}
