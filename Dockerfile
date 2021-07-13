FROM golang:1.16.6

ARG GORELEASER_VERSION=0.116.0
ARG RABBIT_VERSION=2.0.1

ENV GO111MODULE=on

RUN mkdir -p /app/configs \
  && mkdir -p /app/var/logs \
  && mkdir -p /app/var/build \
  && mkdir -p /app/var/releases \
  && curl -L -o /tmp/goreleaser.deb https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_amd64.deb \
  && dpkg -i /tmp/goreleaser.deb \
  && rm -f /tmp/goreleaser.deb \
  && apt-get update \
  && apt-get install -y git \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN curl -sL https://github.com/Clivern/Rabbit/releases/download/${RABBIT_VERSION}/Rabbit_${RABBIT_VERSION}_Linux_x86_64.tar.gz | tar xz
RUN rm LICENSE
RUN rm README.md
RUN mv Rabbit rabbit

COPY ./config.docker.yml /app/configs/

EXPOSE 8080

VOLUME /app/configs
VOLUME /app/var

HEALTHCHECK --interval=5s --timeout=2s --retries=5 --start-period=2s \
  CMD ./rabbit --config /app/configs/config.docker.yml --exec health

CMD ["./rabbit", "--config", "/app/configs/config.docker.yml"]