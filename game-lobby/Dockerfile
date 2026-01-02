FROM golang:1.25 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lobby .

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/lobby .
ENTRYPOINT ["./lobby"]