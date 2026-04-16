FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/unconf-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
	addgroup -S unconf && adduser -S -G unconf -h /home/unconf unconf && \
	mkdir -p /data && chown -R unconf:unconf /data /home/unconf
WORKDIR /home/unconf

COPY --from=builder /out/unconf-server /usr/local/bin/unconf-server

EXPOSE 8080
VOLUME ["/data"]
ENV UNCONF_DB_PATH=/data/unconf.db
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1

USER unconf:unconf
ENTRYPOINT ["/usr/local/bin/unconf-server"]
