FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build
WORKDIR /src
ENV GOPROXY=https://proxy.golang.org|direct \
    GOTOOLCHAIN=local
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
COPY go.mod go.sum ./
RUN for attempt in 1 2 3; do \
      go mod download && exit 0; \
      test "$attempt" -eq 3 && exit 1; \
      sleep $((attempt * 3)); \
    done
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags="-s -w -X github.com/yuanweize/atomic-sync/internal/buildinfo.Version=$VERSION -X github.com/yuanweize/atomic-sync/internal/buildinfo.Commit=$COMMIT -X github.com/yuanweize/atomic-sync/internal/buildinfo.Date=$BUILD_DATE" \
    -o /out/atomic-sync ./cmd/atomic-sync

# Build the current rclone release from source with the two transfer backends
# and commands Atomic Sync actually uses. Explicit dependency floors retain
# fixes for CVE-2026-56852 and GHSA-hrxh-6v49-42gf.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS rclone-build
WORKDIR /src
ENV GOPROXY=https://proxy.golang.org|direct \
    GOTOOLCHAIN=local
ARG TARGETOS
ARG TARGETARCH
ARG RCLONE_VERSION=v1.75.0
COPY build/rclone-main.go.in ./main.go
RUN go mod init atomic-sync-rclone \
    && go mod edit -require=github.com/rclone/rclone@$RCLONE_VERSION \
    && go mod edit -require=golang.org/x/text@v0.39.0 \
    && go mod edit -require=google.golang.org/grpc@v1.82.1
RUN for attempt in 1 2 3; do \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -mod=mod -trimpath \
        -ldflags="-s -w -X github.com/rclone/rclone/fs.Version=${RCLONE_VERSION#v}" \
        -o /out/rclone ./main.go && exit 0; \
      test "$attempt" -eq 3 && exit 1; \
      sleep $((attempt * 3)); \
    done

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
      org.opencontainers.image.description="Auditable directory-unit file transfer control plane" \
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
