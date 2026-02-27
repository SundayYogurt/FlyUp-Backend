# ---------- build stage ----------
FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/mail-svc

# ---------- run stage ----------
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /app/main /app/main
COPY --from=builder /app/internal/templates /app/internal/templates

ENV ENV=prod

CMD ["/app/main"]