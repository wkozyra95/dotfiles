FROM debian
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && \
        apt-get -y install sudo

RUN useradd -ms /bin/bash test
RUN adduser test sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER test
WORKDIR /home/test
ADD . .dotfiles

RUN ~/.dotfiles/bin/mycli install --noninteractive


