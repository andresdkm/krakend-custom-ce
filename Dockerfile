# Multi-stage build for KrakenD CE
FROM golang:1.25-alpine AS builder

RUN go install github.com/cespare/reflex@latest

# Install build dependencies
RUN apk add --no-cache git make bash

# Set working directory
WORKDIR /build

# Clone KrakenD CE repository
RUN git clone --branch v2.11.2 https://github.com/krakend/krakend-ce.git .

# Build KrakenD
RUN make build

# Final stage - lightweight runtime image
FROM alpine:latest

# Install runtime dependencies including cgo support
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    binutils-gold \
    git \
    make \
    bash

# Copy compiled binary from builder
COPY --from=builder /build/krakend /usr/bin/krakend
COPY --from=builder /go/bin/reflex /usr/bin/reflex

# Copy Go toolchain for plugin builds
COPY --from=builder /usr/local/go /usr/local/go

# Set Go environment variables
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:/usr/local/go/bin
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Expose KrakenD port
EXPOSE 8080

ADD entrypoint.sh /

WORKDIR /etc/krakend

ENTRYPOINT [ "/entrypoint.sh" ]