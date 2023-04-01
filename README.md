# dotfiles

### Setup `mycli` inside a Docker container

`v0.0.0` is a fake release that is used to publicly host `mycli` binaries. Files are updated every time `mycli tool release` is run. 

```
curl -L -o mycli https://github.com/wkozyra95/dotfiles/releases/download/v0.0.0/mycli-linux && \
chmod +x mycli && \
./mycli tool setup:environment:docker
```
