FROM ubuntu

RUN apt update && apt install -y libpcap-dev
RUN mkdir /tools
COPY . /tools

WORKDIR /tools
