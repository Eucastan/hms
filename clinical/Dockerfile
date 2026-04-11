FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY clinical/go.mod clinical/go.sum ./
COPY shared/ ../shared/

RUN go mod download

COPY . .

# Build from inside the clinical module directory
WORKDIR /app/clinical
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/clinical ./cmd/http/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/clinical .
COPY --from=builder /app/shared/pkg/grpcserver/retry_policy.json ./retry_policy.json

EXPOSE 4002
EXPOSE 50051

# Non root user for security
RUN adduser -D -u 1001 appuser
USER appuser

CMD ["./clinical"]