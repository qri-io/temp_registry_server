# FROM golang:1.12.4
# LABEL maintainer="sparkle_pony_2000@qri.io"

# ADD . /go/src/github.com/qri-io/qri

# ENV GO111MODULE=on

# # run build
# RUN cd /go/src/github.com/qri-io/qri && make build

# # set default port to 8080, default log level, QRI_PATH env, IPFS_PATH env
# ENV PORT=8080 IPFS_LOGGING="" QRI_PATH=/data/qri IPFS_PATH=/data/ipfs

# # Ports for Swarm TCP, Swarm uTP, API, Gateway, Swarm Websockets
# EXPOSE 4001 4002/udp 5001 8080 8081

# # create directories for IPFS & QRI, setting proper owners
# RUN mkdir -p $IPFS_PATH && mkdir -p $QRI_PATH \
#   && adduser --disabled-password --home $IPFS_PATH --uid 1000 --gid 100 ipfs \
#   && chown 1000:100 $IPFS_PATH \
#   && chown 1000:100 $QRI_PATH

# # Expose the fs-repo & qri-repos as volumes.
# # Important this happens after the USER directive so permission are correct.
# # VOLUME $IPFS_PATH
# # VOLUME $QRI_PATH

# # Set binary as entrypoint, initalizing ipfs & qri repos if none is mounted
# CMD ["qri", "connect", "--setup"]

# build image to actually build the
FROM golang:latest AS builder
LABEL maintainer="sparkle_pony_2000@qri.io"

# add local files to cloud backend
ADD . /temp_registry_server
WORKDIR /temp_registry_server

# build environment variables:
#   * enable go modules
#   * use goproxy
#   * disable cgo for our builds
#   * ensure target os is linux
ENV GO111MODULE=on \
  GOPROXY=https://proxy.golang.org \
  CGO_ENABLED=0 \
  GOOS=linux

# install to produce a binary called "main" in the pwd
#   -a flag rebuild all the packages we’re using,
#      which means all the imports will be rebuilt with cgo disabled.
#   -installsuffix cgo keeps output separate in build caches
#   -o sets the output name to main
#   . says "build this package"
RUN go build -a -installsuffix cgo -o main .

# *** production image ***
# use alpine as base for smaller images
FROM alpine:latest
LABEL maintainer="sparkle_pony_2000@qri.io" 

# configure environment variables,
#   * provide a default server port
#   * set qri & IPFS paths to the data directory
ENV PORT=3000

# Copy our static executable, qri and IPFS directories, which are empty
COPY --from=builder /temp_registry_server/main /bin/temp_registry_server

# need to update to latest ca-certificates, otherwise TLS won't work properly.
# Informative:
# https://hackernoon.com/alpine-docker-image-with-secured-communication-ssl-tls-go-restful-api-128eb6b54f1f
# RUN apk update \
#     && apk upgrade \
#     && apk add --no-cache \
#     ca-certificates \
#     && update-ca-certificates 2>/dev/null || true

# command to run is the temp_registry_server binary
CMD ["/bin/temp_registry_server"]