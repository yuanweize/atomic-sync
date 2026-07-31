# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X github.com/yuanweize/atomic-sync/internal/buildinfo.Version=$VERSION -X github.com/yuanweize/atomic-sync/internal/buildinfo.Commit=$COMMIT -X github.com/yuanweize/atomic-sync/internal/buildinfo.Date=$BUILD_DATE" \
    -o /out/atomic-sync ./cmd/atomic-sync

# rclone 1.74.4 is the current upstream release, but its published binary was
# built before fixes for CVE-2026-56852 and GHSA-hrxh-6v49-42gf landed. Build
# the same release from source with only those security-fixed dependencies
# lifted, so the runtime image does not have to waive known HIGH findings.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS rclone-build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
ARG RCLONE_VERSION=v1.74.4
COPY build/rclone-main.go.in ./main.go
RUN go mod init atomic-sync-rclone \
    && go mod edit -require=github.com/rclone/rclone@$RCLONE_VERSION \
    && go mod edit -require=golang.org/x/text@v0.39.0 \
    && go mod edit -require=google.golang.org/grpc@v1.82.1 \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -mod=mod -trimpath \
      -ldflags="-s -w -X github.com/rclone/rclone/fs.Version=${RCLONE_VERSION#v}" \
      -o /out/rclone ./main.go

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 1000 atomic \
    && adduser -S -D -u 1000 -G atomic -h /home/atomic atomic \
    && mkdir -p /data /config/rclone /home/atomic \
    && chown -R atomic:atomic /data /config /home/atomic \
    && chmod 0750 /data /config /config/rclone /home/atomic
COPY --from=rclone-build /out/rclone /usr/local/bin/rclone
COPY --from=build /out/atomic-sync /usr/local/bin/atomic-sync
LABEL org.opencontainers.image.title="Atomic Sync" \
      org.opencontainers.image.description="Atomic directory migration and media archive orchestrator" \
      org.opencontainers.image.source="https://github.com/yuanweize/atomic-sync" \
      org.opencontainers.image.version="$VERSION" \
      org.opencontainers.image.revision="$COMMIT" \
      org.opencontainers.image.created="$BUILD_DATE" \
      org.opencontainers.image.licenses="MIT"
USER 1000:1000
EXPOSE 8080
ENV HOME=/home/atomic \
    XDG_CONFIG_HOME=/config \
    RCLONE_CONFIG=/config/rclone/rclone.conf \
    ATOMIC_DATA_DIR=/data \
    ATOMIC_LISTEN=:8080 \
    ATOMIC_RCLONE_BIN=/usr/local/bin/rclone
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8080/api/ready || exit 1
ENTRYPOINT ["/usr/local/bin/atomic-sync"]
