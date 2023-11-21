FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER TommyLike<tommylikehu@gmail.com>

# build binary
COPY . /go/src/github.com/opensourceways/app-cla-server
RUN cd /go/src/github.com/opensourceways/app-cla-server && GO111MODULE=on CGO_ENABLED=0 go build -o cla-server -buildmode=pie --ldflags "-s -linkmode 'external' -extldflags '-Wl,-z,now'"

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow git python3 python3-pip && \
    pip3 install git+https://github.com/py-pdf/pypdf.git@3.12.0 && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 cla && \
    useradd -u 1000 -g cla -s /sbin/nologin -m cla

RUN echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd
RUN mkdir /home/cla -p
RUN chmod 700 /home/cla
RUN chown cla:cla /home/cla

RUN echo 'set +o history' >> /root/.bashrc
RUN sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs
RUN rm -rf /tmp/*

USER cla
WORKDIR /home/cla

COPY --chown=cla ./conf /home/cla/conf
COPY --chown=cla ./deploy/app.conf /home/cla/conf/app.conf
COPY --chown=cla ./util/merge_signature.py /home/cla/util/merge_signature.py
COPY --chown=cla --from=BUILDER /go/src/github.com/opensourceways/app-cla-server/cla-server /home/cla

RUN chmod 750 /home/cla/conf
RUN chmod 640 /home/cla/conf/app.conf
RUN chmod 550 /home/cla/util/merge_signature.py
RUN chmod 550 /home/cla/cla-server

RUN echo "umask 027" >> /home/cla/.bashrc
RUN echo 'set +o history' >> /home/cla/.bashrc

ENTRYPOINT ["/home/cla/cla-server"]
