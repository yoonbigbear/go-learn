# 빌드 단계
FROM golang:1.25 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o game-server ./simple-game-server

# 실행 단계 (가벼운 Alpine 리눅스 사용)
FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/game-server .
# user를 추가해서 실행하는게 보안상 좋지만 실습을 위해 생략
ENTRYPOINT ["./game-server"]
