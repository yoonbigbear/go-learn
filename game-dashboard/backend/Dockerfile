# 1. 빌드 단계 (Builder)
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0은 알파인 리눅스에서 실행하기 위해 필수

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./game-dashboard/backend/main.go

# 2. 실행 단계 (Runner)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]