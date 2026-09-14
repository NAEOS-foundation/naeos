FROM golang:1.26.6-alpine@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /naeos ./cmd/naeos/

FROM alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1

RUN apk --no-cache add ca-certificates git

WORKDIR /app

COPY --from=builder /naeos /usr/local/bin/naeos
COPY --from=builder /app/LICENSE /app/NOTICE /usr/local/share/naeos/

RUN adduser -D -u 1000 naeos
USER naeos

HEALTHCHECK --interval=30s --timeout=5s CMD naeos health || exit 1

ENTRYPOINT ["naeos"]
CMD ["--help"]
