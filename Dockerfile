# Step 1: Modules caching
FROM golang:1.26-alpine3.23 AS modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

# Step 2: Builder
FROM golang:1.26-alpine3.23 AS builder

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

# Step 3: Final
FROM alpine:3.23 AS app

RUN apk add --no-cache ca-certificates ffmpeg nodejs python3 py3-pip && \
    python3 -m venv /opt/yt-dlp && \
    /opt/yt-dlp/bin/pip install --no-cache-dir 'yt-dlp[default]==2026.07.04' 'bgutil-ytdlp-pot-provider==2.0.0'

ENV PATH="/opt/yt-dlp/bin:${PATH}"

COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
CMD ["/app"]
