FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /naeos ./cmd/naeos/

FROM alpine:3.19

RUN apk --no-cache add ca-certificates git

WORKDIR /app

COPY --from=builder /naeos /usr/local/bin/naeos

RUN adduser -D -u 1000 naeos
USER naeos

ENTRYPOINT ["naeos"]
CMD ["--help"]
