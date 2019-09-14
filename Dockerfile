FROM golang:1.13-alpine AS builder
RUN apk add --no-cache git
ADD . /src
RUN chown -R 1000:users /src
USER 1000
WORKDIR /src
ENV GOCACHE=/tmp/.go-cache
RUN go build

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder --chown=0:0 /src/ec2-metadata-exporter /usr/local/bin/ec2-metadata-exporter
USER 1000
ENTRYPOINT ["/usr/local/bin/ec2-metadata-exporter"]
