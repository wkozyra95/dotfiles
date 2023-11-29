FROM ubuntu:mantic-20231011

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV DEBIAN_FRONTEND=noninteractive
ENV USERNAME=wojtek

ARG RUST_VERSION=1.74

RUN echo ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true | debconf-set-selections
RUN apt-get update -y -qq && \
  apt-get install -y \
    build-essential \
    curl \
    wget \
    file \
    gnupg2 \
    openjdk-11-jdk \
    unzip \
    direnv \
    python3 \
    rsync \
    zsh \
    build-essential curl pkg-config libssl-dev libclang-dev git sudo \
    libnss3 libatk1.0-0 libatk-bridge2.0-0 libgdk-pixbuf2.0-0 libgtk-3-0 \
    libegl1-mesa-dev libgl1-mesa-dri libxcb-xfixes0-dev mesa-vulkan-drivers \
    ffmpeg libavcodec-dev libavformat-dev libavfilter-dev libavdevice-dev \
    ttf-mscorefonts-installer

ENV UID=1000
ENV GID=1001

RUN deluser ubuntu && groupadd -g $GID $USERNAME && useradd -u $UID -g $GID -ms /bin/bash $USERNAME && adduser $USERNAME sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN mkdir /home/$USERNAME/project
USER $USERNAME

RUN curl https://sh.rustup.rs -sSf | bash -s -- -y
RUN source ~/.cargo/env && rustup install $RUST_VERSION && rustup default $RUST_VERSION

WORKDIR /home/$USERNAME/project

ENTRYPOINT bash
