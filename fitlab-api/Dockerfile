ARG GO_VERSION=1.21
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
  RUN go build -v -o /run-app .
  RUN go build -v -o /fitlab-scheduler ./cmd/scheduler


FROM debian:bookworm

# Install CA certificates for TLS email sending
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /run-app /usr/local/bin/
COPY --from=builder /fitlab-scheduler /usr/local/bin/
CMD ["run-app"]