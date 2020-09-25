# Ghostream

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![pipeline status](https://gitlab.crans.org/nounous/ghostream/badges/master/pipeline.svg)](https://gitlab.crans.org/nounous/ghostream/commits/master)

*Boooo!* A simple streaming server with authentication and open-source technologies.

This project was developped at [Cr@ns](https://crans.org/) to stream events.

Features:

-   WebRTC playback with a lightweight web interface.
-   Low-latency streaming, sub-second with web player.
-   Authentification of incoming stream using LDAP server.

## Installation with Docker

An example is given in [docs/docker-compose.yml](doc/docker-compose.yml).
It uses Traefik reverse proxy.

## References

-   Phil Cluff (2019), *[Streaming video on the internet without MPEG.](https://mux.com/blog/streaming-video-on-the-internet-without-mpeg/)*
-   MDN web docs, *[Signaling and video calling.](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling)*
-   [WebRTC For The Curious](https://webrtcforthecurious.com/)
