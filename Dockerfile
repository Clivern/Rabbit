FROM golang:1.12.5

ENV GO111MODULE=on

ARG VERSION=1.0.0

RUN mkdir -p /app

RUN cd /app

RUN go get -u github.com/goreleaser/goreleaser

RUN curl -sL https://github.com/Clivern/Rabbit/releases/download/$VERSION/Rabbit_$VERSION_Linux_x86_64.tar.gz | tar xz

RUN mv Rabbit rabbit

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN mkdir -p /app/var/build
RUN mkdir -p /app/var/releases

VOLUME /app/configs
VOLUME /app/var/logs
VOLUME /app/var/build
VOLUME /app/var/releases

WORKDIR /app

EXPOSE 8080

CMD ["./rabbit", "--config", "/app/configs/config.prod.yml"]
