# MIT License
#
# Copyright (c) 2020 CADCloud
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
# FROM golang:latest
FROM ubuntu:18.04

# Add Maintainer Info
LABEL maintainer="Jean-Marie Verdun <jmverdun3@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

ENV GOPATH=$GOPATH:/go/src/base:/app
ENV TLS_KEY_PATH=/app/src/tools/certificat.key.unlock
ENV TLS_CERT_PATH=/app/src/tools/certificat.crt
ENV STATIC_ASSETS_DIR=/app/src/static/
ENV TEMPLATES_DIR=/app/src/templates/
ENV CREDENTIALS_TCPPORT=:9100
ENV CREDENTIALS_URI=127.0.0.1
ENV STORAGE_TCPPORT=:9200
ENV MINIO_TCPPORT=:9300
ENV FREECAD_URI=127.0.0.1
ENV FREECAD_TCPPORT=:9210
ENV FREECAD_BINARY=/snap/bin/freecad
ENV FREECAD_TEMPLATE=/app/freecad/
ENV FREECAD_TEMP=/app/root/freecad/
ENV PROJECT_URI=127.0.0.1
ENV PROJECT_TCPPORT=:9211
ENV PROJECT_MINIO_TCPPORT=:9212
ENV PROJECT_TEMP=/app/root/projects/
ENV container docker
ENV PATH /snap/bin:$PATH
# let's build
RUN apt-get --allow-unauthenticated update --allow-insecure-repositories
RUN apt install -y apt-utils
RUN apt install -y golang-1.10 golang-1.10-go golang-1.10-race-detector-runtime golang-1.10-src golang-go golang-race-detector-runtime golang-src
RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates && rm -rf /var/lib/apt/lists/*
RUN cat /etc/apt/sources.list
RUN apt-get update && apt install -qq -y build-essential libssl-dev libcurl4-gnutls-dev libexpat1-dev gettext unzip wget xvfb snapd squashfuse fuse snap-confine sudo fontconfig vim rand
RUN apt-get -y install git
RUN go get golang.org/x/crypto/bcrypt
RUN go get golang.org/x/net/idna
RUN mkdir bin
RUN go build -o bin/storage src/backend/storage.go
RUN go build -o bin/freecad  src/backend/freecad.go
RUN go build -o bin/cacheServer src/backend/cacheServer.go
RUN go build -o bin/minioServer src/backend/minioServer.go
RUN go build -o bin/users src/credential/users.go
RUN go build -o bin/projects src/credential/projects.go
RUN go build -o bin/master src/frontend/master.go
RUN /bin/bash scripts/generate_certificates

RUN chmod 777 /app/start_container
RUN systemctl enable snapd
STOPSIGNAL SIGRTMIN+3

# Expose port 8080 to the outside world
EXPOSE 443

# Command to run the executable
# CMD /app/start_container
CMD [ "/sbin/init" ]

