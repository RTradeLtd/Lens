# Set up Golang build environment
FROM golang:alpine AS build-env
ENV BUILD_HOME=/go/src/github.com/RTradeLtd/Lens

# Mount source code
ADD . ${BUILD_HOME}
WORKDIR ${BUILD_HOME}

# install gcc libs
# RUN apk add build-base

# install git
RUN apk add git

# install RTFS dependency
RUN go get -u -v github.com/RTradeLtd/RTFS

# Build temporal binary
RUN go build -o /bin/temporal-lens \
    ./cmd/temporal-lens

# Copy binary into clean image
FROM alpine
LABEL maintainer "RTrade Technologies Ltd."
RUN mkdir -p /daemon
WORKDIR /daemon
COPY --from=build-env /bin/temporal-lens /usr/local/bin

# Set up directories
RUN mkdir -p /data/lens  

# Set default configuration
ENV CONFIG_DAG /data/lens/config.json
COPY ./test/config.json /data/lens/config.json

# Set default command
ENTRYPOINT [ "temporal-lens", "server"]
