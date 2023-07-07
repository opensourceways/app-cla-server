FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/app-cla-server
RUN cd /go/src/github.com/opensourceways/app-cla-server && GO111MODULE=on CGO_ENABLED=0 go build

# copy binary config and utils
FROM golang:latest
RUN apt-get update && apt-get install -y python3 && apt-get install -y python3-pip && pip3 install PyPDF2 && mkdir -p /opt/app/
COPY ./conf /opt/app/conf
COPY ./util/merge-signature.py /opt/app/util/merge-signature.py
# overwrite config yaml
COPY ./deploy/app.conf /opt/app/conf
COPY ./deploy/app.conf.yaml /opt/app/conf
COPY  --from=BUILDER /go/src/github.com/opensourceways/app-cla-server/cla-server /opt/app

WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/cla-server"]
