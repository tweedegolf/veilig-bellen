FROM docker.tgrep.nl/docker/debian-dev:buster

ARG GO_VERSION
ENV GO_VERSION ${GO_VERSION}

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        libssl-dev \
        inotify-tools \ 
        go-dep \
        procps \
    ; \
    rm -rf /var/lib/apt/lists/*;

# install go
RUN wget "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz"
RUN tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
RUN rm go$GO_VERSION.linux-amd64.tar.gz

ENV PATH=${PATH}:/usr/local/go/bin
ENV GOPATH=/go

WORKDIR $GOPATH/src
