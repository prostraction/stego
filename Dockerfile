# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o stego-web ./cmd/web/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/stego-web .
COPY --from=builder /app/cmd/web/frontend ./frontend

EXPOSE 8080

CMD ["./stego-web"]
