# Set up Golang build environment
FROM golang:latest AS build-env

# Set versions
ARG LENSVERSION

# Mount source code
ENV BUILD_HOME=/go/src/github.com/RTradeLtd/Lens
ADD . ${BUILD_HOME}
WORKDIR ${BUILD_HOME}

# Install build dependencies
RUN apt-get update; \
    apt-get install -y sudo curl git
## Tensorflow
RUN bash setup/scripts/tensorflow_install.sh
ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:/usr/local/lib
## Tesseract
RUN bash setup/scripts/tesseract_install.sh
## Golang dependencies
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure -v
## Go-fitz
RUN go get -u -tags gcc7 github.com/gen2brain/go-fitz

# Build Lens binary
RUN go build \
    -tags gcc7 \
    -o /bin/temporal-lens \
    -ldflags "-X main.Version=$LENSVERSION" \
    ./cmd/temporal-lens

# Copy binary into clean image
FROM ubuntu:16.04
LABEL maintainer "RTrade Technologies Ltd."
COPY --from=build-env /bin/temporal-lens /usr/local/bin
ADD setup /setup

# Install runtime dependencies
RUN apt-get update; \
    apt-get install -y sudo curl bash
## Tensorflow
RUN bash setup/scripts/tensorflow_install.sh
ENV LD_LIBRARY_PATH=/usr/local/lib
## Tesseract
RUN bash /setup/scripts/tesseract_install.sh
RUN ls /usr/lib/x86_64-linux-gnu

# Set up directories
RUN mkdir -p /data/lens

# Set default configuration
ENV CONFIG_DAG /data/lens/config.json
COPY ./test/config.json /data/lens/config.json

# Set default command
ENTRYPOINT [ "temporal-lens" ]
CMD [ "server" ]
