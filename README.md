<div align="center">

  # Portainer-AutoUpdater

  A process for automating Docker container updates for containers managed by portainer.
  <br/><br/>

[![GoDoc](https://godoc.org/github.com/sjafferali/portainer-autoupdater?status.svg)](https://pkg.go.dev/github.com/sjafferali/portainer-autoupdater)
[![Go Report Card](https://goreportcard.com/badge/github.com/sjafferali/portainer-autoupdater)](https://goreportcard.com/report/github.com/sjafferali/portainer-autoupdater)
[![Release](https://github.com/sjafferali/portainer-autoupdater/actions/workflows/release.yaml/badge.svg)](https://github.com/sjafferali/portainer-autoupdater/actions?query=branch%3Amain)
[![Pulls from DockerHub](https://img.shields.io/docker/pulls/sjafferali/portainer-autoupdater.svg)](https://hub.docker.com/r/sjafferali/portainer-autoupdater)
[![latest version](https://img.shields.io/github/tag/sjafferali/portainer-autoupdater.svg)](https://github.com/sjafferali/portainer-autoupdater/releases)

</div>

## Usage

With portainer-autoupdater you can update the running version of your containerized app simply by pushing a new image to the Docker Hub or your own image registry. Unlike watchtower, and other similar tools, this makes use of the portainer API for all checks as well as upgrades which ensures portainer will always be up to date with all updates performed.

The autoupdater app will query the portainer API, then trigger a  update on the stack, service or container to perform the update.

### Current Features
- Auto updating of stacks
- Auto updating of services (swarm only)

### Planned Features
- Update Notifications
- User/Password Authentication
- Auto updating of containers (non-swarm only)

### Why not Watchtower?
Watchtower and other docker auto-update tools use docker directly for checking for updates as well as performing updates. This app is specifically for systems being managed by portainer as the update process makes use of the portainer API for all checks and updates. 

## Examples

### Minimal Compose
```
version: '3.8'
services:
  autoupdater:
    image: sjafferali/portainer-autoupdater:latest
    restart: unless-stopped
    environment:
      - AUTOUPDATER_DRY_RUN=0
      - AUTOUPDATER_ENDPOINT=http://portainer.url:9000
      - AUTOUPDATER_TOKEN=${AUTOUPDATER_TOKEN}
```

### Full Compose
```
version: '3.8'
services:
  autoupdater:
    image: sjafferali/portainer-autoupdater:latest
    container_name: autoupdater
    restart: unless-stopped
    environment:
      - AUTOUPDATER_INTERVAL=300s
      - AUTOUPDATER_DRY_RUN=0
      - AUTOUPDATER_ENDPOINT=http://portainer.url:9000
      - AUTOUPDATER_TOKEN=${AUTOUPDATER_TOKEN}
      - AUTOUPDATER_LOGLEVEL=INFO

      - AUTOUPDATER_ENABLE_STACKS=1
      - AUTOUPDATER_INCLUDE_STACK_IDS=122
      - AUTOUPDATER_EXCLUDE_STACK_IDS=255,54
      - AUTOUPDATER_INCLUDE_STACK_NAMES=dozzle
      - AUTOUPDATER_EXCLUDE_STACK_NAMES=cupsd

      - AUTOUPDATER_ENABLE_SERVICES=0
      - AUTOUPDATER_INCLUDE_SERVICE_IDS=122
      - AUTOUPDATER_EXCLUDE_SERVICE_IDS=255,54
      - AUTOUPDATER_INCLUDE_SERVICE_NAMES=dozzle
      - AUTOUPDATER_EXCLUDE_SERVICE_NAMES=cupsd

      # not implemented
      - AUTOUPDATER_ENABLE_CONTAINERS=0
```

### Environment Variables

| Name | Default | Required | Description |
|:--|:--|:--|:--|
| AUTOUPDATER_INTERVAL | 300s | no | interval at which the updater checks for image updates to be performed |
| AUTOUPDATER_DRY_RUN | 1 | no | only log, but don't perform updates |
| AUTOUPDATER_ENDPOINT |  | yes | portainer api endpoint |
| AUTOUPDATER_TOKEN |  | yes | portioner api token to use for authentication |
| AUTOUPDATER_LOGLEVEL | INFO | no | loglevel to use for runs |
| AUTOUPDATER_ENABLE_STACKS | 1 | no | enable checking for stack updates |
| AUTOUPDATER_EXCLUDE_STACK_IDS |  | no | stack IDs of stacks that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_STACK_IDS |  | no | stack IDs of stacks that should be included from checks; if not set, all stacks are included |
| AUTOUPDATER_EXCLUDE_STACK_NAMES |  | no | stack names of stacks that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_STACK_NAMES |  | no | stack names of stacks that should be included from checks; if not set, all stacks are included |
| AUTOUPDATER_ENABLE_CONTAINERS | 1 | no | enable checking for container updates |
| AUTOUPDATER_EXCLUDE_CONTAINERS |  | no | container IDs of containers that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_CONTAINERS |  | no | containers IDs of containers that should be included from checks; if not set, all containers are included |
| AUTOUPDATER_ENABLE_SERVICES | 1 | no | enable checking for service updates (swarm only) |
| AUTOUPDATER_EXCLUDE_SERVICE_IDS |  | no | service IDs of services that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_SERVICE_IDS |  | no | services IDs of services that should be included from checks; if not set, all services are included |
| AUTOUPDATER_EXCLUDE_SERVICE_NAMES |  | no | service names of services that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_SERVICE_NAMES |  | no | services names of services that should be included from checks; if not set, all services are included |
