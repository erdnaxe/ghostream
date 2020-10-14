# Ghostream

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/mod/gitlab.crans.org/nounous/ghostream)](https://pkg.go.dev/mod/gitlab.crans.org/nounous/ghostream)
[![Go Report Card](https://goreportcard.com/badge/gitlab.crans.org/nounous/ghostream)](https://goreportcard.com/report/gitlab.crans.org/nounous/ghostream)
[![pipeline status](https://gitlab.crans.org/nounous/ghostream/badges/dev/pipeline.svg)](https://gitlab.crans.org/nounous/ghostream/commits/dev)
[![coverage report](https://gitlab.crans.org/nounous/ghostream/badges/dev/coverage.svg)](https://gitlab.crans.org/nounous/ghostream/-/commits/dev)
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/erdnaxe/ghostream)](https://hub.docker.com/r/erdnaxe/ghostream)

*Boooo!* A simple streaming server with authentication and open-source technologies.

This project was developped at [Cr@ns](https://crans.org/) to stream events.

Features:

-   WebRTC playback with a lightweight web interface.
-   SRT stream input, supported by FFMpeg, OBS and Gstreamer.
-   Low-latency streaming, sub-second with web player.
-   Authentication of incoming stream using LDAP server.
-   Possibility to forward stream to other streaming servers.

## Installation on Debian/Ubuntu


On Ubuntu 20.10+ or Debian 11+, you can install directly ghostream,

```bash
sudo apt install git golang ffmpeg libsrt1-openssl
go get gitlab.crans.org/nounous/ghostream
```

On Ubuntu 20.04 or Debian Buster, you may manually install libsrt 1.4.1: install [libsrt1-openssl 1.4.1](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt1-openssl_1.4.1-5+b1_amd64.deb) then [libsrt-openssl-dev 1.4.1](http://ftp.fr.debian.org/debian/pool/main/s/srt/libsrt-openssl-dev_1.4.1-5+b1_amd64.deb).

For development, you may clone this repository, then `go run main.go`.

## Installation with Docker

An example is given in [docs/docker-compose.yml](docs/docker-compose.yml).
It uses Traefik reverse proxy.

You can also launch the Docker image using,

```
docker build . -t ghostream
docker run -it --rm -p 2112:2112 -p 9710:9710/udp -p 8080:8080 -p 10000-10005:10000-10005/udp ghostream
```

## Configuration

Ghostream can be configured by placing [ghostream.yml](docs/ghostream.example.yml) in `/etc/ghostream/`.
You can overwrite the configuration path with `GHOSTREAM_CONFIG` environnement variable.
You can also overwride any value using environnement variables, e.g. `GHOSTREAM_AUTH_BACKEND=ldap` will change the authentification backend.

## Streaming

As stated by OBS wiki, when streaming you should adapt the latency to `2.5 * (the round-trip time with server, in μs)`.

### With OBS

As OBS uses FFMpeg, you need to have FFMpeg compiled with SRT support. To check if SR is available, run `ffmpeg -protocols | grep srt`.
On Windows and MacOS, OBS comes with his own FFMpeg that will work.

In OBS, go to "Settings" -> "Output" -> "Recording" the select "Output to URL" and change the URL to `srt://127.0.0.1:9710?streamid=demo:demo`.
For container, you may use MPEGTS for now (will change).

### With GStreamer

To stream your X11 screen,

```bash
gst-launch-1.0 ximagesrc startx=0 show-pointer=true use-damage=0 \
! videoconvert \
! x264enc bitrate=32000 tune=zerolatency speed-preset=veryfast byte-stream=true threads=1 key-int-max=15 intra-refresh=true ! video/x-h264, profile=baseline, framerate=30/1 \
! mpegtsmux \
! srtserversink uri=srt://127.0.0.1:9710/ latency=1000000 streamid=demo:demo
```

*This might not work at the moment.*

## Playing stream

### With a web browser and WebRTC

Ghostream expose a web server on `0.0.0.0:8080` by default.
By opening this in a browser, you will be able to get instructions on how to stream, and if you append `/streamname` to the URL, then you will be able to watch the stream named `streamname`.

The web player also integrates a side widget that is configurable.

#### Integrate the player in an iframe

To integrate the player without the side widget, you can append `?nowidget` to the URL.

```HTML
<iframe src="https://example.com/stream_name?nowidget" scrolling="no" allowfullscreen="true" width="1280" height="750.4" frameborder="0"></iframe>
```

The iframe size should be a 16/9 ratio, with additionnal 30.4px for the control bar.

### With ffplay

You may directly open the SRT stream with ffplay:

```bash
ffplay -fflags nobuffer srt://127.0.0.1:9710?streamid=demo
```

### With MPV

As MPV uses ffmpeg libav, support for SRT streams can be easily added.
[See current pull request.](https://github.com/mpv-player/mpv/pull/8139)

## Troubleshooting

### ld returns an error when launching ghostream

When missing `libsrt-openssl-dev` on Debian/Ubuntu,
then srtgo package is unable to build.

```bash
~/ghostream$ go run main.go
# github.com/haivision/srtgo
/usr/bin/ld: cannot find -lsrt
/usr/bin/ld: cannot find -lsrt
/usr/bin/ld: cannot find -lsrt
collect2: error: ld returned 1 exit status
```

## References

-   Phil Cluff (2019), *[Streaming video on the internet without MPEG.](https://mux.com/blog/streaming-video-on-the-internet-without-mpeg/)*
-   MDN web docs, *[Signaling and video calling.](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling)*
-   [WebRTC For The Curious](https://webrtcforthecurious.com/)
-   OBS Wiki, *[Streaming With SRT Protocol.](https://obsproject.com/wiki/Streaming-With-SRT-Protocol)*
-   Livepeer media server, *[Evaluate Go-FFmpeg Bindings](https://github.com/livepeer/lpms/issues/24)*
