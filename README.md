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

## Application

With portainer-autoupdater you can update the running version of your containerized app simply by pushing a new image to the Docker Hub or your own image registry. 

The autoupdater app will query the portainer API, then trigger a  update on the stack, service or container to perform the update.   

Current Features
- Auto updating of stacks. 

Planned Features
- Update Notifications
- User/Password Authentication
- Auto updating of services
- Auto updating of containers

## Usage

### Minimal Compose

```
version: '3.8'
services:
  autoupdater:
    image: sjafferali/portainer-autoupdater:latest
    restart: unless-stopped
    environment:
      - AUTOUPDATER_INTERVAL=300s
      - AUTOUPDATER_DRY_RUN=0
      - AUTOUPDATER_ENDPOINT=http://portainer.url:9000
      - AUTOUPDATER_TOKEN=${AUTOUPDATER_TOKEN}
      - AUTOUPDATER_LOGLEVEL=info
      - AUTOUPDATER_ENABLE_STACKS=1
      - AUTOUPDATER_EXCLUDE_STACKS=3,65
      - AUTOUPDATER_INCLUDE_STACKS=""
```

### Full Compose
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

### Environment Variables

| Name | Default | Required | Description |
|:--|:--|:--|:--|
| AUTOUPDATER_INTERVAL | 300s | no | interval at which the updater checks for image updates to be performed |
| AUTOUPDATER_DRY_RUN | 1 | no | only log, but don't perform updates |
| AUTOUPDATER_ENDPOINT |  | yes | portainer api endpoint |
| AUTOUPDATER_TOKEN |  | yes | portioner api token to use for authentication |
| AUTOUPDATER_LOGLEVEL | INFO | no | loglevel to use for runs |
| AUTOUPDATER_ENABLE_STACKS | 1 | no | enable checking for stack updates |
| AUTOUPDATER_EXCLUDE_STACKS |  | no | stack IDs of stacks that should be excluded from auto update |
| AUTOUPDATER_INCLUDE_STACKS |  | no | stack IDs of stacks that should be included from checks; if not set, all stacks are included |
