package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// open-match 패키지
	om "open-match.dev/open-match/pkg/pb"

	// Agones & Kubernetes
	allocationv1 "agones.dev/agones/pkg/apis/allocation/v1"
	agonesclient "agones.dev/agones/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func main() {
	// Open Match 백엔드 서비스 클라이언트 생성
	omConn, _ := grpc.NewClient("open-match-backend.open-match.svc.cluster.local:50505", grpc.WithTransportCredentials(insecure.NewCredentials()))
	omClient := om.NewBackendServiceClient(omConn)

	log.Println("Director 시작: 매칭 감시 중")

	for {
		ctx := context.Background()
		tracer := otel.Tracer("director-service")

		// 새로운 주기(Cycle)마다 Span 생성
		ctx, span := tracer.Start(ctx, "FetchMatches")
		defer span.End()

		// 2. FetchMatches 요청 (매칭 만들어와!)
		req := &om.FetchMatchesRequest{
			Config: &om.FunctionConfig{
				Host: "my-mmf.default.svc.cluster.local", // 우리가 만들 MMF 주소
				Port: 50502,
				Type: om.FunctionConfig_GRPC,
			},
			Profile: &om.MatchProfile{
				Name: "simple-match-profile",
				Pools: []*om.Pool{
					{Name: "all-users"}, // 조건 없이 다 가져와
				},
			},
		}

		stream, err := omClient.FetchMatches(context.Background(), req)
		if err != nil {
			log.Printf("Fetch 실패: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// 3. 매칭 결과 수신 및 티켓 할당
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("매칭 수신 실패: %v", err)
				break
			}

			match := resp.GetMatch()
			matchId := match.GetMatchId()
			playerCount := len(match.Tickets)
			log.Printf("매칭 완료: 매치 ID %s, 플레이어 수 %d", matchId, playerCount)

			tracer := otel.Tracer("director-service")
			// 상위 Span(FetchMatches)의 자식 Span 생성
			ctx, span := tracer.Start(ctx, "AllocateGameServer")
			defer span.End()
			// Span에 태그 추가 (검색 용이)
			span.SetAttributes(attribute.String("match_id", matchId))
			slog.InfoContext(ctx, "Agones 게임 서버 할당 시도", "match_id", matchId)

			// 4. 게임 서버 할당 (Agones)
			serverIP, err := allocateGameServer()
			if err != nil {
				log.Printf("Aggones로부터 게임 서버 할당 실패: %v", err)
				continue // 서버가 없으면 매칭을 버림.
			}

			// 5. 티켓 할당 (AssignTickets 호출)
			var assignments []*om.AssignmentGroup
			assignments = append(assignments,
				&om.AssignmentGroup{
					TicketIds: getTicketIds(match.Tickets),
					Assignment: &om.Assignment{
						Connection: serverIP,
					},
				})
			assignReq := &om.AssignTicketsRequest{
				Assignments: assignments,
			}

			// Ticket과 ServerIP 정보를 OpenMatch에 기록
			_, err = omClient.AssignTickets(context.Background(), assignReq)
			if err != nil {
				log.Printf("티켓 할당 실패: %v", err)
				continue
			}
			log.Printf("티켓 할당 성공: 매치 ID %s", match.GetMatchId())
		}

		time.Sleep(1 * time.Second)
	}
}

func getTicketIds(tickets []*om.Ticket) []string {
	var ids []string
	for _, ticket := range tickets {
		ids = append(ids, ticket.Id)
	}
	return ids
}

func allocateGameServer() (string, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("Kubernetes 클러스터 구성 로드 실패: %v", err)
	}

	agonesClient, err := agonesclient.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("Agones 클라이언트 생성 실패: %v", err)
	}

	// 할당 요청서 작성
	req := &allocationv1.GameServerAllocation{
		Spec: allocationv1.GameServerAllocationSpec{
			Required: allocationv1.GameServerSelector{
				LabelSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"agones.dev/fleet": "simple-udp-fleet",
					},
				},
			},
		},
	}

	// 요청 전송
	log.Println("게임 서버 할당 요청 중...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alloc, err := agonesClient.AllocationV1().GameServerAllocations("default").Create(ctx, req, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("게임 서버 할당 실패: %v", err)
	}

	if alloc.Status.State != allocationv1.GameServerAllocationAllocated {
		return "", fmt.Errorf("게임 서버 할당 실패: 상태 %s", alloc.Status.State)
	}

	ip := alloc.Status.Address
	port := alloc.Status.Ports[0].Port
	address := fmt.Sprintf("%s:%d", ip, port)
	log.Printf("게임 서버 할당 완료: %s", address)

	return address, nil
}
