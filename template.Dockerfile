# Dockerfile for running Go tests inside a container
ARG GO_VERSION=1.25.6

FROM golang:${GO_VERSION}

WORKDIR /app

# Copy the entire package (build context)
COPY . .

# Download dependencies
RUN go mod download

# Keep container alive for exec commands
ENTRYPOINT ["/bin/sh", "-c", "trap 'exit 0' TERM; while :; do sleep 0.1; done"]