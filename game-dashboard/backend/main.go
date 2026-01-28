package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "open-match.dev/open-match/pkg/pb" // Open Match Protobuf
)

func main() {
	r := gin.Default()

	// 1. CORS 설정 (React 포트 허용)
	r.Use(cors.Default())

	// 2. Open Match QueryService 연결 (K8s 내부 주소 or 포트포워딩 주소)
	// 로컬 Minikube 개발 시엔 kubectl port-forward 후 localhost:51503 사용
	conn, err := grpc.Dial("open-match-query.open-match.svc.cluster.local:50503", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// gRPC 클라이언트 생성
	queryClient := pb.NewQueryServiceClient(conn)

	// 3. API 만들기: 대기 중인 티켓 조회
	r.GET("/api/tickets", func(c *gin.Context) {
		// Open Match에 QueryTickets 요청 (Pool 필터링 없이 전체 조회 예시)
		req := &pb.QueryTicketsRequest{
			Pool: &pb.Pool{Name: "all-tickets"},
		}

		// gRPC 호출 스트림 받기
		stream, err := queryClient.QueryTickets(context.Background(), req)
		if err != nil {
			log.Printf("QueryTickets Error: %v", err)
			c.JSON(500, gin.H{"error": err.Error(), "msg": "Open Match Query 실패"})
			return
		}

		var tickets []*pb.Ticket
		for {
			resp, err := stream.Recv()
			if err != nil {
				break // 스트림 끝
			}
			tickets = append(tickets, resp.Tickets...)
		}

		c.JSON(http.StatusOK, tickets)
	})

	r.Run(":8080")
}
