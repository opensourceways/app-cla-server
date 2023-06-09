FROM golang:latest as BUILDER

# build binary
COPY . /go/src/github.com/opensourceways/software-package-server
RUN cd /go/src/github.com/opensourceways/software-package-server/message-server && GO111MODULE=on CGO_ENABLED=0 go build

WORKDIR /go/src/github.com/opensourceways/software-package-server/download
RUN cd /go/src/github.com/opensourceways/software-package-server/download && GO111MODULE=on CGO_ENABLED=0 go build -a -o download .

# copy binary config and utils
FROM alpine:latest
WORKDIR /opt/app/

COPY  --from=BUILDER /go/src/github.com/opensourceways/software-package-server/message-server/message-server /opt/app/message-server
COPY  --from=BUILDER /go/src/github.com/opensourceways/software-package-server/download/download /opt/app/download

COPY softwarepkg/infrastructure/pkgciimpl/clone_repo.sh /opt/app/clone_repo.sh
COPY softwarepkg/infrastructure/pkgciimpl/pull_request.sh /opt/app/pull_request.sh

RUN chmod +x /opt/app/*.sh  \
    && apk update  \
    && apk add --no-cache git libc6-compat curl

ENTRYPOINT ["/opt/app/message-server"]
