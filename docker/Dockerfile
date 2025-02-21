FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3

RUN microdnf install -y golang

WORKDIR /app

ARG VERSION=dev
ENV VERSION=${VERSION}
COPY bin/findcert-${VERSION}-linux-amd64 /app/findcert

EXPOSE 8080
CMD ["/app/findcert"]