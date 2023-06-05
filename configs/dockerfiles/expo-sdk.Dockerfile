FROM ubuntu:latest

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG NDK_VERSION=21.4.7075529
ARG NVM_VERSION=0.39.3
ARG NODE_VERSION=18.16.0
ARG YARN_VERSION=1.22.19

ENV DEBIAN_FRONTEND=noninteractive
ENV USERNAME=wojtek
ENV ANDROID_HOME=/home/${USERNAME}/Android/Sdk
ENV ANDROID_SDK_HOME=${ANDROID_HOME}
ENV ANDROID_SDK_ROOT=${ANDROID_HOME}
ENV ANDROID_NDK_HOME=${ANDROID_HOME}/ndk/${NDK_VERSION}
ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ENV PATH=${ANDROID_NDK_HOME}:${ANDROID_HOME}/cmdline-tools/tools/bin:${ANDROID_HOME}/tools:${ANDROID_HOME}/tools/bin:${ANDROID_HOME}/platform-tools:${PATH}

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

#
# https://github.com/inversepath/usbarmory-debian-base_image/issues/9#issuecomment-466594168
#
RUN mkdir ~/.gnupg && echo "disable-ipv6" >> ~/.gnupg/dirmngr.conf

RUN useradd -ms /bin/bash $USERNAME && adduser $USERNAME sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER $USERNAME
WORKDIR /home/$USERNAME


#
# Install Android dependencies
#
RUN curl -s https://dl.google.com/android/repository/commandlinetools-linux-7583922_latest.zip > /tmp/tools.zip && \
  mkdir -p $ANDROID_HOME/cmdline-tools && \
  unzip -q -d $ANDROID_HOME/cmdline-tools /tmp/tools.zip && \
  mv $ANDROID_HOME/cmdline-tools/cmdline-tools $ANDROID_HOME/cmdline-tools/tools && \
  rm /tmp/tools.zip && \
  (yes || true) | sdkmanager --licenses > /dev/null && \
  (yes || true) | sdkmanager "platform-tools" > /dev/null && \
  (yes || true) | sdkmanager \
  "platforms;android-30" \
  "build-tools;29.0.3" \
  "extras;android;m2repository" \
  "extras;google;m2repository" \
  "extras;google;google_play_services" \
  "ndk;$NDK_VERSION" > /dev/null

#
# Install nvm and node
#
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v$NVM_VERSION/install.sh | bash \
  && source /home/$USERNAME/.nvm/nvm.sh \
  && nvm install $NODE_VERSION \
  && npm install -g yarn@$YARN_VERSION

RUN curl -L -o mycli https://github.com/wkozyra95/dotfiles/releases/download/v0.0.0/mycli-linux && \
  chmod +x mycli && \
  ./mycli tool setup:environment:docker

RUN mkdir /home/$USERNAME/project
WORKDIR /home/$USERNAME/project

ENTRYPOINT bash
