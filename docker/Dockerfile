# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder

ARG VERSION=unknown
ENV CGO_ENABLED=1

WORKDIR /src

RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc make && \
    rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux go build -trimpath -ldflags "-s -w -X github.com/blob42/gosuki/pkg/build.Describe=${VERSION}" \
    -o /out/gosuki ./cmd/gosuki && \
    GOOS=linux go build -trimpath -ldflags "-s -w -X github.com/blob42/gosuki/pkg/build.Describe=${VERSION}" \
    -o /out/suki ./cmd/suki

FROM debian:bookworm-slim

ENV GOSUKI_DATA_DIR=/var/lib/gosuki
ENV GOSUKI_CONFIG_DIR=/etc/gosuki
ENV GOSUKI_DB=${GOSUKI_DATA_DIR}/gosuki.db
ENV GOSUKI_CONFIG=${GOSUKI_CONFIG_DIR}/config.toml

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd -r gosuki && useradd -r -g gosuki -d /var/lib/gosuki -s /sbin/nologin gosuki && \
    mkdir -p ${GOSUKI_DATA_DIR} ${GOSUKI_CONFIG_DIR} && \
    chown -R gosuki:gosuki /var/lib/gosuki /etc/gosuki

COPY --from=builder /out/gosuki /usr/local/bin/gosuki
COPY --from=builder /out/suki /usr/local/bin/suki
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

VOLUME ["/var/lib/gosuki", "/etc/gosuki"]

EXPOSE 2025

USER gosuki

ENTRYPOINT ["/entrypoint.sh"]
CMD ["start"]
