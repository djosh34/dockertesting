# Alternative Dockerfile for testing custom Dockerfile support
# This creates an environment marker file that tests can check for
ARG GO_VERSION=1.25.6

FROM golang:${GO_VERSION}

WORKDIR /app

# Copy the entire package (build context)
COPY . .

# Create a marker file to verify this Dockerfile was used
RUN echo "alternative-dockerfile-marker" > /tmp/dockerfile_marker.txt

# Download dependencies
RUN go mod download

# Keep container alive for exec commands
ENTRYPOINT ["/bin/sh", "-c", "trap 'exit 0' TERM; while :; do sleep 0.1; done"]
