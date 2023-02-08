#!/bin/bash

BINFILE=autoupdater
VERSION=$(git describe --tags)
echo "Building $VERSION..."
go build -o $BINFILE -ldflags "-X github.com/containrrr/watchtower/internal/meta.Version=$VERSION" cmd/autoupdater
