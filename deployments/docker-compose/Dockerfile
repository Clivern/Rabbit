FROM golang:1.12.5 as builder

ENV GO111MODULE=on

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN mkdir -p /app/var/build
RUN mkdir -p /app/var/releases

WORKDIR /app

RUN curl -sL https://github.com/Clivern/Rabbit/releases/download/0.1.1/Rabbit_0.1.1_Linux_x86_64.tar.gz | tar xz
RUN curl -sL https://github.com/goreleaser/goreleaser/releases/download/v0.109.0/goreleaser_Linux_x86_64.tar.gz | tar xz

RUN mv Rabbit rabbit

FROM golang:1.12.5

ENV GO111MODULE=on

COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/var/logs /app/var/logs
COPY --from=builder /app/var/build /app/var/build
COPY --from=builder /app/var/releases /app/var/releases
COPY --from=builder /app/rabbit /app/rabbit
COPY --from=builder /app/goreleaser /bin/goreleaser

WORKDIR /app

EXPOSE 8080

CMD ["./rabbit", "--config", "/app/configs/config.docker.yml"]