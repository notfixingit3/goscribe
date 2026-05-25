# GoScribe Dockerfile
#
# Multi-stage build for minimal production image.
#
# Usage:
#   docker build -t goscribe .
#   docker run --rm goscribe version
#   docker run --rm -v ~/.goscribe.yaml:/home/goscribe/.goscribe.yaml goscribe generate /src
#   docker run --rm -v $(pwd):/src goscribe generate /src
#
# CI/CD example:
#   docker run --rm \
#     -e GOSCRIBE_PROVIDER=openai \
#     -e GOSCRIBE_MODEL=gpt-4 \
#     -v $(pwd):/src \
#     goscribe generate /src

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /goscribe \
    ./cmd/goscribe

# Stage 2: Minimal runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates git && \
    adduser -D -u 1000 goscribe

COPY --from=builder /goscribe /usr/local/bin/goscribe

USER goscribe

ENTRYPOINT ["goscribe"]
CMD ["--help"]
