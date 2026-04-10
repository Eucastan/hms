FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY pharmacy/go.mod pharmacy/go.sum ./
COPY shared/ ../shared/

RUN go mod download

COPY . .

# Build from inside the pharmacy module directory
WORKDIR /app/pharmacy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/pharmacy ./cmd/http/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/pharmacy .
COPY --from=builder /app/shared/pkg/grpcserver/retry_policy.json ./retry_policy.json

EXPOSE 4004
EXPOSE 50052

RUN adduser -D -u 1001 appuser
USER appuser

CMD [ "./pharmacy" ]