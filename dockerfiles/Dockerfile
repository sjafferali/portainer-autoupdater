#
# Builder
#

FROM golang:alpine as builder

RUN apk add --no-cache \
    alpine-sdk \
    ca-certificates \
    git \
    tzdata

COPY . /autoupdater

RUN \
  cd /autoupdater && \
  \
  GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-extldflags '-static' -X github.com/sjafferali/portainer-autoupdater/internal/meta.Version=$(git describe --tags)" ./cmd/autoupdater


#
# autoupdater
#

FROM scratch


# copy files from other container
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /autoupdater/autoupdater /autoupdater

ENTRYPOINT ["/autoupdater"]
