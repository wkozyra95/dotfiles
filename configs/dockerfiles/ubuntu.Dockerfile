FROM ubuntu:latest

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV DEBIAN_FRONTEND=noninteractive
ENV USERNAME=wojtek
ENV CURRENT_ENV=docker

RUN dpkg --add-architecture i386 \
  && apt-get update \
  && apt-get install -yq \
  build-essential \
  curl \
  wget \
  file \
  git \
  gnupg2 \
  openjdk-11-jdk \
  unzip \
  direnv \
  netcat \
  python3 \
  rsync \
  sudo

RUN useradd -ms /bin/bash $USERNAME && adduser $USERNAME sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN mkdir /home/$USERNAME/project
USER $USERNAME

WORKDIR /home/$USERNAME

RUN curl -L -o mycli https://github.com/wkozyra95/dotfiles/releases/download/v0.0.0/mycli-linux && \
  chmod +x mycli && \
  ./mycli tool setup:environment:docker

WORKDIR /home/$USERNAME/project

ENTRYPOINT bash
