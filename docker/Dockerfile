# docker build -t houseabsolute/catalauncher-player -f Dockerfile .

FROM ubuntu:18.04

RUN rm /etc/machine-id && \
    apt-get -y update && \
    apt-get -y install \
        libfreetype6 \
        liblua5.3 \
        libsdl2-2.0-0 \
        libsdl2-image-2.0-0 \
        libsdl2-mixer-2.0-0 \
        libsdl2-ttf-2.0-0
