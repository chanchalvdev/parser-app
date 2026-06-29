FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /out/api ./api
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=8 \
  CMD curl -f http://127.0.0.1:8080/health || exit 1
CMD ["./api"]
