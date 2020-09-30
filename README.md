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
-   SRT stream input, supported by FFMpeg, OBS and Gstreamer.
-   Low-latency streaming, sub-second with web player.
-   Authentification of incoming stream using LDAP server.
-   Possibility to forward stream to other streaming servers.

## Installation on Debian/Ubuntu

You need at least libsrt 1.4.1. On Ubuntu 20.04 or Debian Buster, you may manually install [libsrt-openssl-dev](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt1-openssl_1.4.1-5+b1_amd64.deb) then [libsrt-openssl-dev](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt-openssl-dev_1.4.1-5+b1_amd64.deb).

You may clone this repository, then `go run main.go` for debugging, or `go get gitlab.crans.org/nounous/ghostream`.

## Installation with Docker

An example is given in [docs/docker-compose.yml](docs/docker-compose.yml).
It uses Traefik reverse proxy.

You can also launch the Docker image using,

```
docker build . -t ghostream
docker run -it --rm -p 2112:2112 -p 9710:9710/udp -p 8080:8080 -p 10000-10005:10000-10005/udp ghostream
```

## Streaming

As stated by OBS wiki, when streaming you should adapt the latency to `2.5 * (the round-trip time with server, in μs)`.

### With OBS

As OBS uses FFMpeg, you need to have FFMpeg compiled with SRT support. To check if SR is available, run `ffmpeg -protocols | grep srt`.
On Windows and MacOS, OBS comes with his own FFMpeg that will work.

In OBS, go to "Settings" -> "Stream" and change "Service" to "Custom..." and "Server" to `srt://127.0.0.1:9710`.

### With GStreamer

To stream X11 screen,

```bash
gst-launch-1.0 ximagesrc startx=0 show-pointer=true use-damage=0 \
! videoconvert \
! x264enc bitrate=32000 tune=zerolatency speed-preset=veryfast byte-stream=true threads=1 key-int-max=15 intra-refresh=true ! video/x-h264, profile=baseline, framerate=30/1 \
! mpegtsmux \
! srtserversink uri=srt://127.0.0.1:9710/ latency=1000000
```

## References

-   Phil Cluff (2019), *[Streaming video on the internet without MPEG.](https://mux.com/blog/streaming-video-on-the-internet-without-mpeg/)*
-   MDN web docs, *[Signaling and video calling.](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling)*
-   [WebRTC For The Curious](https://webrtcforthecurious.com/)
-   OBS Wiki, *[Streaming With SRT Protocol.](https://obsproject.com/wiki/Streaming-With-SRT-Protocol)*
