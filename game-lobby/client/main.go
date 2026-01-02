package main

import (
	"context"
	"game-lobby/pb"
	"io"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	lobbyAddress := "127.0.0.1:13862"

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
			return
		}
		log.Printf("[%s] 상태 대기 중: %s", userID, resp.Status)
	}

}
