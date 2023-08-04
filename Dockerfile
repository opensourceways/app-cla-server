FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/app-cla-server
RUN cd /go/src/github.com/opensourceways/app-cla-server && GO111MODULE=on CGO_ENABLED=0 go build -o cla-server

# copy binary config and utils
FROM golang:latest
RUN apt-get update && apt-get install -y python3 && apt-get install -y python3-pip && pip3 install PyPDF2==3.0.0 --break-system-packages
RUN useradd -ms /bin/bash cla
USER cla
WORKDIR /home/cla
COPY --chown=cla ./conf /home/cla/conf
COPY --chown=cla ./deploy/app.conf /home/cla/conf/app.conf
COPY --chown=cla ./util/merge_signature.py /home/cla/util/merge_signature.py
COPY --chown=cla --from=BUILDER /go/src/github.com/opensourceways/app-cla-server/cla-server /home/cla

ENTRYPOINT ["/home/cla/cla-server"]
