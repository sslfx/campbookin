FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/gateway-service ./cmd/gateway/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/gateway-service .
EXPOSE 8080
CMD ["./gateway-service"]