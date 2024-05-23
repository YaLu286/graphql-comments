FROM golang:1.18-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/gQLserver .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/gQLserver .
COPY --from=builder /app/migrations/* ./


CMD ["./gQLserver"]