FROM debian:buster

RUN mkdir /home/echoserver
COPY ./echoserver /usr/bin
COPY ./create_directories.sh /home/echoserver
COPY ./entrypoint.sh /home/echoserver

ENV SERVER_PORT=

ENTRYPOINT /bin/bash /home/echoserver/entrypoint.sh $SERVER_PORT

