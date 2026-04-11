FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY internal/services/admin/go.mod internal/services/admin/go.sum ./internal/services/admin/
COPY pkg/proto/go.mod pkg/proto/go.sum ./pkg/proto/

RUN go mod download

COPY internal/services/admin/ ./internal/services/admin/
COPY pkg/proto/ ./pkg/proto/


RUN CGO_ENABLED=0 GOOS=linux go build -o /admin-srv ./internal/services/admin/cmd/main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /admin-srv .

EXPOSE 50051
CMD ["./admin-srv"]