FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY auth/go.mod auth/go.sum ./
COPY shared/ ../shared/

RUN go mod download

COPY . .

# Build from inside the auth module directory
WORKDIR /app/auth
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/auth ./cmd/http/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/auth .

EXPOSE 4000

RUN adduser -D -u 1001 appuser
USER appuser

CMD ["./auth"]