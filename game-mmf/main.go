package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	om "open-match.dev/open-match/pkg/pb"
)

type matchFunctionService struct {
	om.UnimplementedMatchFunctionServer
	queryClient om.QueryServiceClient
}

// Director로부터 매칭 요청을 받으면 호출되는 함수
func (s *matchFunctionService) Run(req *om.RunRequest, stream om.MatchFunction_RunServer) error {
	log.Printf("매칭 요청 받음 (프로필: %s)", req.Profile.Name)

	// QueryService에서 티켓 가져오기
	// Director가 보내준 Pool("all-users") 정보를 그대로 사용해서 조회한다.
	pool := req.Profile.Pools[0]

	queryStream, err := s.queryClient.QueryTickets(stream.Context(), &om.QueryTicketsRequest{
		Pool: pool,
	})
	if err != nil {
		return fmt.Errorf("티켓 조회 실패: %v", err)
	}

	var tickets []*om.Ticket
	for {
		resp, err := queryStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		tickets = append(tickets, resp.Tickets...)
	}

	log.Printf("조회된 티켓 수: %d", len(tickets))

	// 매칭 로직 (2명씩 짝짓기)
	for i := 0; i < len(tickets)-1; i += 2 {
		ticketA := tickets[i]
		ticketB := tickets[i+1]

		match := &om.Match{
			MatchId:       fmt.Sprintf("match-%d", i/2),
			MatchProfile:  req.Profile.Name,
			MatchFunction: "simple-2p-function",
			Tickets:       []*om.Ticket{ticketA, ticketB}, // 이 2명을 매칭
		}

		if err := stream.Send(&om.RunResponse{Proposal: match}); err != nil {
			log.Printf("전송 실패: %v", err)
		}
		log.Printf("매칭 제안 생성됨: %s 와 %s", ticketA.Id, ticketB.Id)
	}
	return nil
}

func main() {
	qsAddr := "open-match-query.open-match.svc.cluster.local:50503"
	conn, err := grpc.NewClient(qsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("QueryService 연결 실패: %v", err)
	}
	defer conn.Close()

	lis, err := net.Listen("tcp", ":50502")
	if err != nil {
		log.Fatalf("리스너 생성 실패: %v", err)
	}

	s := grpc.NewServer()
	om.RegisterMatchFunctionServer(s, &matchFunctionService{queryClient: om.NewQueryServiceClient(conn)})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
	log.Println("MMF 서버 시작( Port: 50502)...")
}
