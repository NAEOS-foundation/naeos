FROM golang:1.26.10-alpine@sha256:a2012b4a1d1f9f7e9f44e0d39b805a901e856b1dd80b4e6ee4338db547050308 AS builder

WORKDIR /app

# This module currently has no external dependencies, so do not require a
# go.sum file to exist. go mod download remains ready for future dependencies.
COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /naeos-outreach-server ./cmd/naeos-outreach-server/

FROM alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /naeos-outreach-server /usr/local/bin/naeos-outreach-server

# Container networking requires a non-loopback listener. This explicit opt-in
# is scoped to the image; local execution remains loopback-only by default.
ENV NAEOS_OUTREACH_ADDR=0.0.0.0:8088
ENV NAEOS_OUTREACH_ALLOW_NON_LOOPBACK=true

RUN adduser -D -u 1000 naeos
USER naeos

EXPOSE 8088

HEALTHCHECK --interval=30s --timeout=5s CMD wget -q -O /dev/null http://127.0.0.1:8088/health || exit 1

ENTRYPOINT ["naeos-outreach-server"]
