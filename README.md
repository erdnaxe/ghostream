# Ghostream

[![License: GPL v2](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](https://www.gnu.org/licenses/gpl-2.0.txt)
[![pipeline status](https://gitlab.crans.org/nounous/ghostream/badges/master/pipeline.svg)](https://gitlab.crans.org/nounous/ghostream/commits/master)

*Boooo!* A simple streaming server with authentication and open-source technologies.

Features:

-   RTMPS ingest (OBS compatible).
-   RTMPS playback (VLC, MPV compatible).
-   WebRTC playback with a lightweight web interface.
-   Low-latency streaming, sub-second with web player.
-   Authentification of incoming stream using a LDAP server.

## Installation on Debian/Ubuntu

This instructions were tested on Debian Buster.
You need to unable non-free repository to have AAC codec.

### NGINX

Install NGINX server.

Copy [60-ghostream.conf module](doc/nginx/modules-available/60-ghostream.conf) to `/etc/nginx/modules-available/60-ghostream.conf`.
Then symlink this file to `/etc/nginx/modules-enabled/60-ghostream.conf`.

Copy [ghostream site](doc/nginx/sites-available/ghostream) to `/etc/nginx/sites-available/ghostream`.
Then symlink this file to `/etc/nginx/sites-enabled/ghostream`.

Restart NGINX.

You may need to generate SSL certificates.
For the initial generation, you may comment the SSL rules and use Certbot NGINX module.

### OvenMediaEngine

Install some required libraries,

```bash
sudo apt update
sudo apt install --no-install-recommends build-essential ca-certificates nasm autoconf zlib1g-dev tcl cmake curl libssl-dev libsrtp2-dev libopus-dev libjemalloc-dev pkg-config libvpx-dev libswscale-dev libswresample-dev libavfilter-dev libavcodec-dev libx264-dev libfdk-aac-dev
```

Then compile libsrt,

```
curl -LOJ https://github.com/Haivision/srt/archive/v1.3.3.tar.gz
tar xvf srt-1.3.3.tar.gz
cd srt-1.3.3
./configure --prefix="/opt/ovenmediaengine" --enable-shared --enable-static=0
make -j 2
sudo make install
cd ..
```

Then compile OvenMediaEngine FFMpeg,

```
curl -LOJ https://github.com/AirenSoft/FFmpeg/archive/ome/3.4.tar.gz
tar xvf FFmpeg-ome-3.4.tar.gz
cd FFmpeg-ome-3.4
./configure --prefix="/opt/ovenmediaengine" --enable-gpl --enable-nonfree --extra-cflags="-I/opt/ovenmediaengine/include" --extra-ldflags="-L$/opt/ovenmediaengine/lib -Wl,-rpath,/opt/ovenmediaengine/lib" --extra-libs=-ldl --enable-shared --disable-static --disable-debug --disable-doc --disable-programs --disable-avdevice --disable-dct --disable-dwt --disable-error-resilience --disable-lsp --disable-lzo --disable-rdft --disable-faan --disable-pixelutils --disable-everything --enable-zlib --enable-libopus --enable-libvpx --enable-libfdk_aac --enable-libx264 --enable-encoder=libvpx_vp8,libvpx_vp9,libopus,libfdk_aac,libx264 --enable-decoder=aac,aac_latm,aac_fixed,h264 --enable-parser=aac,aac_latm,aac_fixed,h264 --enable-network --enable-protocol=tcp --enable-protocol=udp --enable-protocol=rtp --enable-demuxer=rtsp --enable-filter=asetnsamples,aresample,aformat,channelmap,channelsplit,scale,transpose,fps,settb,asettb
make
sudo make install
cd ..
```

Finally compile OvenMediaEngine,

```
curl -LOJ https://github.com/AirenSoft/OvenMediaEngine/archive/v0.10.7.tar.gz
tar xvf OvenMediaEngine-0.10.7.tar.gz
cd OvenMediaEngine-0.10.7/src
make release
sudo make install
```

Copy [Server.xml](doc/ovenmediaengine/conf/Server.xml) to `/usr/share/ovenmediaengine/conf/Server.xml`.

Now enable and start OvenMediaEngine, `sudo systemctl enable --now ovenmediaengine`.

### Ghostreamer web server

On Debian you can install [ghostream deb](https://gitlab.crans.org/nounous/ghostream/-/jobs/artifacts/master/raw/build/ghostream_0.1.0_all.deb?job=build-deb).

On other system, you might install manually the Python module and Systemd unit [ghostreamer.service](debian/ghostream.service).

You might customize `/etc/default/ghostream`.

## Installation with Docker

An example is given in [doc/docker-compose.yml](doc/docker-compose.yml).
It uses Traefik reverse proxy.
