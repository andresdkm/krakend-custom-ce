# ARG GOLANG_VERSION
# ARG ALPINE_VERSION
# FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

# RUN apk --no-cache --virtual .build-deps add make gcc musl-dev binutils-gold

# COPY . /app
# WORKDIR /app

# RUN make build


# FROM alpine:${ALPINE_VERSION} as runtime

# LABEL maintainer="community@krakend.io"

# RUN apk upgrade --no-cache --no-interactive && apk add --no-cache ca-certificates tzdata && \
#     adduser -u 1000 -S -D -H krakend && \
#     mkdir /etc/krakend && \
#     echo '{ "version": 3 }' > /etc/krakend/krakend.json

# COPY --from=builder /app/krakend /usr/bin/krakend

# USER 1000

# WORKDIR /etc/krakend

# ENTRYPOINT [ "/usr/bin/krakend" ]
# CMD [ "run", "-c", "/etc/krakend/krakend.json" ]

# EXPOSE 8000 8090



# Multi-stage build for KrakenD CE
# FROM golang:1.25-alpine AS builder

# RUN go install github.com/cespare/reflex@latest

# # Install build dependencies
# RUN apk add --no-cache git make bash

# # Set working directory
# WORKDIR /build

# # Clone KrakenD CE repository
# RUN git clone --branch v2.11.2 https://github.com/krakend/krakend-ce.git .

# # Build KrakenD
# RUN make build

# # Final stage - lightweight runtime image
# FROM alpine:latest

# # Install runtime dependencies including cgo support
# RUN apk add --no-cache \
#     ca-certificates \
#     tzdata \
#     gcc \
#     musl-dev \
#     binutils-gold \
#     git \
#     make \
#     bash

# # Copy compiled binary from builder
# COPY --from=builder /build/krakend /usr/bin/krakend
# COPY --from=builder /go/bin/reflex /usr/bin/reflex

# # Copy Go toolchain for plugin builds
# COPY --from=builder /usr/local/go /usr/local/go

# # Set Go environment variables
# ENV GOROOT=/usr/local/go
# ENV PATH=$PATH:/usr/local/go/bin
# ENV CGO_ENABLED=1
# ENV GO111MODULE=on

# # Expose KrakenD port
# EXPOSE 8080

# ADD entrypoint.sh /

# WORKDIR /etc/krakend

# ENTRYPOINT [ "/entrypoint.sh" ]