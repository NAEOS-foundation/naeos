FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build ./cmd/app

FROM alpine:3.19
COPY --from=build /app/app /app/app
CMD ["/app/app"]
