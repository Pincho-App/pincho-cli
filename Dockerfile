FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X 'gitlab.com/wirepusher/cli/cmd.version=${VERSION}' -X 'gitlab.com/wirepusher/cli/cmd.commit=${COMMIT}' -X 'gitlab.com/wirepusher/cli/cmd.date=${DATE}'" \
    -o wirepusher main.go

# Final minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/wirepusher /usr/local/bin/

ENTRYPOINT ["wirepusher"]
CMD ["--help"]
