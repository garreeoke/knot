#FROM mhart/alpine-node:6.4.0
FROM alpine:3.5
MAINTAINER Chris Dornsife <chris@applariat.com>

ARG KOPS_VERSION
ARG KUBECTL_VERSION

RUN apk update \
    && apk add --no-cache ca-certificates wget make openssh bash py-pip gcc libffi-dev libtool musl-dev openssl-dev python-dev \
    && wget https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl \
    && mv kubectl /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && wget https://github.com/kubernetes/kops/releases/download/$KOPS_VERSION/kops-linux-amd64 \
    && chmod +x kops-linux-amd64 \
    && mv kops-linux-amd64 /usr/local/bin/kops \
    && pip install azure-cli

ADD target/cluster-manager /cluster-manager
ADD kube-addons /kube-addons
ADD kops-editor.sh /kops-editor.sh
ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

