FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY lab/go.mod lab/go.sum ./
COPY shared/ ../shared/

RUN go mod download

COPY . .

# Build from inside the lab module directory
WORKDIR /app/lab
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/lab ./cmd/http/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/lab .
COPY --from=builder /app/shared/pkg/grpcserver/retry_policy.json ./retry_policy.json

EXPOSE 4003
EXPOSE 50053

RUN adduser -D -u 1001 appuser
USER appuser

CMD ["./lab"]