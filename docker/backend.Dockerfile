FROM golang:1.14-buster

RUN apt-get update -y
RUN apt-get upgrade -y
RUN apt-get install inotify-tools -y

WORKDIR /app

CMD bin/watch.sh main.go