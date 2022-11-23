FROM golang:1.18.3 AS base
ENV GOFLAGS -mod=vendor

RUN apt-get update && apt-get install -y git curl wait-for-it && rm -fr /var/lib/apt/lists/* && \
  mkdir -p /go/src/github.com/kevineaton

ADD ./docker/task /go/bin/
ADD ./docker/migrate /go/bin/

FROM base

ADD ./ /go/src/github.com/kevineaton/kesplora-api
WORKDIR /go/src/github.com/kevineaton/kesplora-api


RUN task build
CMD ["task", "run"]
