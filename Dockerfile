FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/app-cla-server
RUN cd /go/src/github.com/opensourceways/app-cla-server && CGO_ENABLED=1 go build -v -o ./cla-server main.go

# copy binary and config
FROM golang:latest
RUN apt-get update && apt-get install -y python-pip && mkdir -p /opt/app/
COPY ./conf /opt/app/
COPY  --from=BUILDER /go/src/github.com/opensourceways/app-cla-server/cla-server /opt/app

WORKDIR /opt/app/cla-server
ENTRYPOINT ["/opt/app/cla-server"]
