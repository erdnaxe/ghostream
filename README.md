# Ghostream

*Boooo!* A simple streaming server with authentication and open-source technologies.

![logo](doc/ghostream.svg)

## Installation

### NGINX

Copy [60-ghostream.conf module](doc/nginx/modules-available/60-ghostream.conf) to `/etc/nginx/modules-available/60-ghostream.conf`.

Copy [ghostream site](doc/nginx/sites-available/ghostream) to `/etc/nginx/sites-available/ghostream`.

### OvenMediaEngine

Copy [Server.xml](doc/ovenmediaengine/conf/Server.xml) to `/usr/share/ovenmediaengine/conf/Server.xml`.

Now enable and start OvenMediaEngine, `sudo systemctl enable --now ovenmediaengine`.
