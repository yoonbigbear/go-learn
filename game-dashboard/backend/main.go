package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"

	// 1. Agones & Kubernetes 라이브러리 추가

	agonesclient "agones.dev/agones/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Open Match 라이브러리 필요
	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

// 전역 변수 (Open Match 연결용)
var queryServiceClient pb.QueryServiceClient

// 전역 변수로 Agones 클라이언트 선언 (재사용을 위해)
var agonesClient *agonesclient.Clientset

func main() {
	// --- 1. Kubernetes/Agones 클라이언트 초기화 ---
	// (A) 클러스터 내부(In-Cluster)에서 실행될 때 (Pod 안에서)
	config, err := rest.InClusterConfig()
	if err != nil {
		// (B) 로컬(Outside-Cluster)에서 실행될 때 (kubeconfig 사용 - 개발용)
		// ~/.kube/config 파일을 찾음
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("쿠버네티스 설정 로드 실패: %v", err)
		}
	}

	// Agones 클라이언트 생성
	agonesClient, err = agonesclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("Agones 클라이언트 생성 실패: %v", err)
	}

	// --- 2. HTTP 핸들러 등록 ---
	http.HandleFunc("/api/tickets", getTickets)         // 기존 티켓 조회
	http.HandleFunc("/api/gameservers", getGameServers) // [New] 게임서버 조회

	// ... 기존 코드 ...
	log.Println("Frontend Server started at :5150")
	if err := http.ListenAndServe(":5150", nil); err != nil {
		log.Fatal(err)
	}
}

// [New] 게임 서버 목록을 가져오는 핸들러
func getGameServers(w http.ResponseWriter, r *http.Request) {
	// CORS 헤더 (필요시)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Agones가 설치된 네임스페이스를 지정해야 함 (보통 "agones-system" 또는 "default")
	// 사용자님 환경에 맞는 네임스페이스를 적으세요.
	namespace := "agones-system"

	// Agones API 호출: GameServer 목록 조회
	// ListOptions{} 로 모든 게임 서버를 가져옵니다.
	gsList, err := agonesClient.AgonesV1().GameServers(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("게임 서버 조회 에러: %v", err)
		http.Error(w, "Failed to fetch game servers", http.StatusInternalServerError)
		return
	}

	// 프론트엔드로 JSON 응답
	// gsList.Items 안에 []GameServer 배열이 들어있습니다.
	if err := json.NewEncoder(w).Encode(gsList.Items); err != nil {
		log.Printf("JSON 인코딩 에러: %v", err)
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
	}
}

func initOpenMatchClient() {
	// Open Match Query Service 주소 (쿠버네티스 내부 주소)
	// 포트 50503은 QueryService 기본 포트입니다.
	conn, err := grpc.Dial("open-match-query.open-match.svc.cluster.local:50503", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Open Match 연결 실패: %v", err)
	}
	queryServiceClient = pb.NewQueryServiceClient(conn)
}

func getTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// QueryService에게 "모든 티켓 내놔" 라고 요청
	// Pool 필터 없이 요청하면 전체 스캔 (주의: 실제 운영에선 비효율적일 수 있음)
	stream, err := queryServiceClient.QueryTickets(context.Background(), &pb.QueryTicketsRequest{
		Pool: &pb.Pool{
			// 모든 티켓을 가져오기 위한 빈 필터 (또는 특정 태그 필터)
			// 필터가 없으면 에러가 날 수도 있으니, 보통은 "mode.demo" 같은 태그를 겁니다.
		},
	})
	if err != nil {
		http.Error(w, "QueryTickets 실패: "+err.Error(), 500)
		return
	}

	var allTickets []*pb.Ticket
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("스트림 에러: %v", err)
			break
		}
		allTickets = append(allTickets, resp.Tickets...)
	}

	json.NewEncoder(w).Encode(allTickets)
}
