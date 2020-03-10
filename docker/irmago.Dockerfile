FROM golang:1.14-buster

RUN apt-get update -y
RUN apt-get upgrade -y
RUN apt-get install go-dep -y

RUN go get -d github.com/privacybydesign/irmago
WORKDIR $GOPATH/src/github.com/privacybydesign/irmago

RUN dep init
RUN dep ensure
RUN go install ./irma

CMD irma server -v