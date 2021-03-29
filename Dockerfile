FROM golang:latest as BUILDER

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
ARG GOPROXY
WORKDIR /go/src/github.com/opensourceways/app-cla-server
COPY ./ .
RUN CGO_ENABLED=1 go build -v -o ./cla-server main.go

# copy binary config and utils
FROM python:latest
ARG GOPROXY
RUN if [ "$GOPROXY" != "" ]; then pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple; fi
RUN pip install PyPDF2
WORKDIR /opt/app/
COPY ./conf ./conf
COPY ./util/merge-signature.py ./util
# overwrite config yaml
COPY ./deploy/app.conf ./conf
COPY  --from=BUILDER /go/src/github.com/opensourceways/app-cla-server/cla-server .

ENTRYPOINT ["/opt/app/cla-server"]
