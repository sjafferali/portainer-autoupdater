#!/bin/bash

BINFILE=autoupdater
VERSION=$(git describe --tags)
echo "Building $VERSION..."
go build -o $BINFILE -ldflags "-X github.com/sjafferali/portainer-autoupdater/internal/meta.Version=$VERSION" github.com/sjafferali/portainer-autoupdater/cmd/autoupdater
