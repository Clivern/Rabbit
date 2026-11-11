FROM golang:1.27.4 AS builder

ARG SLOTH_VERSION=0.1.0
ARG SLOTH_COMMIT=none
ARG SLOTH_BUILD_DATE=unknown
ARG SLOTH_BUILT_BY=docker
ARG SLOTH_EDITION=oss

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${SLOTH_VERSION} -X main.commit=${SLOTH_COMMIT} -X main.date=${SLOTH_BUILD_DATE} -X main.builtBy=${SLOTH_BUILT_BY} -X main.edition=${SLOTH_EDITION}" \
    -o /out/sloth .

RUN mkdir -p /out/configs /out/var/logs && \
    cp config.dist.yml /out/configs/config.dist.yml

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /out/sloth /app/sloth
COPY --from=builder --chown=65532:65532 /out/configs /app/configs
COPY --from=builder --chown=65532:65532 /out/var /app/var

EXPOSE 8080

VOLUME ["/app/configs", "/app/var"]

USER 65532:65532

ENTRYPOINT ["/app/sloth"]
CMD ["server", "-c", "/app/configs/config.dist.yml"]
