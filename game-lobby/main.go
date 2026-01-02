package main

import (
	"context"
	"fmt"
	"game-lobby/pb" // go.mod에 적인 모듈 이름을 기준으로 시작해야 함
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	om "open-match.dev/open-match/pkg/pb"
)

type lobbyServer struct {
	pb.UnimplementedLobbyServiceServer
	omFrontend om.FrontendServiceClient // open-matchd와 통신할 클라이언트
}

func (s *lobbyServer) JoinMatch(req *pb.JoinRequest, stream pb.LobbyService_JoinMatchServer) error {
	ctx := context.Background()
	log.Printf("JoinMatch called: user_id=%s", req.GetUserId())

	ticketReq := &om.CreateTicketRequest{
		Ticket: &om.Ticket{
			SearchFields: &om.SearchFields{
				// MMF가 매칭에 사용할 속성들
				Tags: []string{"mode.demo"},
			},
		},
	}

	// Open Match에 티켓 생성 요청
	ticket, err := s.omFrontend.CreateTicket(ctx, ticketReq)
	if err != nil {
		log.Printf("CreateTicket error: %v", err)
		return err
	}
	log.Printf("Ticket created: id=%s", ticket.Id)

	// 클라이언트에게 티켓 생성 완료 응답 전송
	stream.Send(&pb.JoinResponse{Status: "waiting - ticket created"})

	// 매칭 할당 감시 시작
	// Open Match는 매칭이 되면 티켓에 Assignment 정보를 업데이트 해줌
	watchStream, err := s.omFrontend.WatchAssignments(ctx, &om.WatchAssignmentsRequest{
		TicketId: ticket.Id,
	})
	if err != nil {
		return fmt.Errorf("WatchAssignments error: %v", err)
	}

	for {
		watchResp, err := watchStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("WatchAssignments Recv error: %v", err)
			break
		}

		if watchResp.Assignment != nil {
			log.Printf("매칭 성공. 연결 정보: %s", watchResp.Assignment.Connection)

			// 할당 정보는 보통 게임 서버의 IP:Port 형태로 제공됨
			// 실제 구현에선 Agones가 준 IP/Port를 분리해서 넣어야 함
			stream.Send(&pb.JoinResponse{
				Status:       "matched",
				GameServerIp: watchResp.Assignment.Connection,
			})
			return nil
		}
	}
	return nil
}

func main() {

	// 쿠버네티스 내부 주소를 사용해서 Open Match 프런트엔드에 연결
	omAddr := "open-match-frontend.open-match.svc.cluster.local:50504"
	conn, err := grpc.NewClient(omAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Open Match frontend: %v", err)
	}
	defer conn.Close()

	// lobby서버 열기
	lis, err := net.Listen("tcp", ":5151")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLobbyServiceServer(s, &lobbyServer{
		omFrontend: om.NewFrontendServiceClient(conn),
	})

	log.Printf("Lobby server listening at %v", lis.Addr())
	s.Serve(lis)
}
