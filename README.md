# Ghostream

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/mod/gitlab.crans.org/nounous/ghostream)](https://pkg.go.dev/mod/gitlab.crans.org/nounous/ghostream)
[![Go Report Card](https://goreportcard.com/badge/gitlab.crans.org/nounous/ghostream)](https://goreportcard.com/report/gitlab.crans.org/nounous/ghostream)
[![pipeline status](https://gitlab.crans.org/nounous/ghostream/badges/master/pipeline.svg)](https://gitlab.crans.org/nounous/ghostream/commits/master)
[![coverage report](https://gitlab.crans.org/nounous/ghostream/badges/master/coverage.svg)](https://gitlab.crans.org/nounous/ghostream/-/commits/master)

*Boooo!* A simple streaming server with authentication and open-source technologies.

This project was developped at [Cr@ns](https://crans.org/) to stream events.

Features:

-   WebRTC playback with a lightweight web interface.
-   Low-latency streaming, sub-second with web player.
-   Authentification of incoming stream using LDAP server.

## Installation on Debian/Ubuntu

You need at least libsrt 1.4.1. On Ubuntu 20.04 or Debian Buster, you may manually install [libsrt1-openssl](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt1-openssl_1.4.1-5+b1_amd64.deb) then [libsrt-openssl-dev](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt-openssl-dev_1.4.1-5+b1_amd64.deb).

You may clone this repository, then `go run main.go` for debugging, or `go get gitlab.crans.org/nounous/ghostream`.

## Installation with Docker

An example is given in [docs/docker-compose.yml](docs/docker-compose.yml).
It uses Traefik reverse proxy.

You can also launch the Docker image using,

```
docker build . -t ghostream
docker run -it --rm -p 8080:8080 -p 2112:2112 -p 9710:9710 ghostream
```

## References

-   Phil Cluff (2019), *[Streaming video on the internet without MPEG.](https://mux.com/blog/streaming-video-on-the-internet-without-mpeg/)*
-   MDN web docs, *[Signaling and video calling.](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling)*
-   [WebRTC For The Curious](https://webrtcforthecurious.com/)
