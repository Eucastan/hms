FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum ./
COPY shared/ ../shared/

RUN go mod download

COPY . .

# Build from inside the api-gateway module directory
WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/gateway ./cmd/gateway/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/gateway .

EXPOSE 8000

RUN adduser -D -u 1001 appuser
USER appuser

CMD ["./gateway"]