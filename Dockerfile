FROM golang:1.24-bookworm AS builder
LABEL stage=builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o video-manager .

FROM debian:bookworm-slim
# Build argument to optionally include ffmpeg in the final image.
# Set to "true" to include ffmpeg for video transcoding workloads.
# Defaults to "false" for a lightweight base image without ffmpeg.
# Usage: docker build --build-arg INCLUDE_FFMPEG=true -t video-manager:ffmpeg .
ARG INCLUDE_FFMPEG=false

# Conditionally install ffmpeg when INCLUDE_FFMPEG=true.
# This adds ~200-300MB to the image size but enables video processing capabilities.
RUN if [ "$INCLUDE_FFMPEG" = "true" ]; then \
    apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/*; \
    fi
WORKDIR /app

COPY --from=builder /app/video-manager .

EXPOSE 25009

CMD ["./video-manager"]
