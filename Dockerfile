FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o video-manager .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/video-manager .

EXPOSE 25009

CMD ["./video-manager"]
