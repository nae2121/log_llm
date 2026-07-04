FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -H appuser

COPY --from=builder /server /app/server
COPY migrations /app/migrations

ENV BACKEND_PORT=8080
ENV MIGRATIONS_DIR=/app/migrations

EXPOSE 8080

USER appuser

CMD ["/app/server"]
