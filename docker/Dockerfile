FROM golang:1.12.5 as builder

WORKDIR /app
COPY ./ /app

RUN GO111MODULE=on go build -o rabbit rabbit.go

FROM golang:1.12.5

ARG GORELEASER_VERSION=0.110.0

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

COPY docker/config.docker.yml /app/configs/config.yml
COPY --from=builder /app/rabbit /app/rabbit
COPY --from=builder /app/LICENSE /app/LICENSE

WORKDIR /app

EXPOSE 8080

CMD ["./rabbit", "--config", "/app/configs/config.yml"]

VOLUME /app/var/releases

HEALTHCHECK --interval=5s --timeout=2s --retries=5 --start-period=2s \
  CMD ./rabbit --config /app/configs/config.yml --exec health
