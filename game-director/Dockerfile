FROM golang:1.25 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o director .

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/director .
ENTRYPOINT ["./director"]
