FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o unconf-server ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/unconf-server .

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./unconf-server"]
