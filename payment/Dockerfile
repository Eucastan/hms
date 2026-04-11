FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY payment/go.mod payment/go.sum ./
COPY shared/ ../shared/

RUN go mod download

#COPY payment/ ./payment/
COPY . .

# Build from inside the payment module directory
WORKDIR /app/payment
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/payment ./cmd/http/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/payment .
COPY --from=builder /app/shared/pkg/grpcserver/retry_policy.json ./retry_policy.json

EXPOSE 4005
EXPOSE 50055

RUN adduser -D -u 1001 appuser
USER appuser

CMD [ "./payment" ]