package main

import (
	"context"
	"game-lobby/pb"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	lobbyAddress := "127.0.0.1:5151"

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		runClient(lobbyAddress, "User-123")
	}()

	time.Sleep(500 * time.Millisecond) // 서버가 먼저 시작되도록 잠시 대기

	go func() {
		defer wg.Done()
		runClient(lobbyAddress, "User-456")
	}()

	wg.Wait()
	log.Println("테스트 종료")
}

func runClient(addr string, userID string) {
	// 로비서버 연결
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] 서버 연결 실패: %v", userID, err)
		return
	}
	defer conn.Close()
	c := pb.NewLobbyServiceClient(conn)

	// 매칭 요청
	log.Printf("[%s] JoinMatch 호출", userID)
	stream, err := c.JoinMatch(context.Background(), &pb.JoinRequest{UserId: userID})
	if err != nil {
		log.Printf("[%s] JoinMatch 호출 실패: %v", userID, err)
		return
	}

	// 응답 대기
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[%s] 응답 수신 실패: %v", userID, err)
			break
		}

		if resp.Status == "matched" {
			log.Printf("[%s] 매칭 완료! 서버 정보: %s", userID, resp.GameServerIp)
			newIP := "127.0.0.1"
			_, port, _ := net.SplitHostPort(resp.GameServerIp)
			finalAddr := net.JoinHostPort(newIP, port)
			// 여기서는 테스트를 위해 미리 정의된 gameServerAddr를 사용합니다.
			// 실제 코드: sendUDPMessage(resp.GameServerIp, userID)
			sendUDPMessage(finalAddr, userID)
			return
		}

		log.Printf("[%s] 상태 대기 중: %s", userID, resp.Status)
	}

}

func sendUDPMessage(gameServerAddr string, userId string) {
	log.Printf("[%s] 게임 서버에 UDP 메시지 전송 시도: %s", userId, gameServerAddr)

	conn, err := net.Dial("udp", gameServerAddr)
	if err != nil {
		log.Printf("게임 서버 연결 실패: %v", err)
		return
	}
	defer conn.Close()

	message := []byte("shutdown")

	_, err = conn.Write(message)
	if err != nil {
		log.Printf("메시지 전송 실패: %v", err)
		return
	}
	log.Printf("메시지 전송 성공: %s", message)

	// 3. 서버로부터 응답 대기
	// 타임아웃을 설정하여 서버가 응답이 없을 경우 무한정 기다리는 것을 방지합니다.
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // 5초 타임아웃

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer) // conn.ReadFrom(buffer) 대신 conn.Read(buffer) 사용
	if err != nil {
		// net.Error의 Timeout() 메서드로 타임아웃 에러인지 확인할 수 있습니다.
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("서버 응답 시간 초과.")
		}
		log.Printf("응답 수신 실패: %v", err)
	}

	// 4. 응답 처리
	responseMsg := string(buffer[:n])
	log.Printf("서버로부터 응답 수신: %s", responseMsg)
}
