# syntax = docker/dockerfile:1.3
FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS builder

ARG REPOSITORY=Lirt
ARG PLUGIN=k8s-secret-replicator
ARG PKG=github.com/Lirt/k8s-secret-replicator
ARG VERSION=0.0.0
ARG GIT_SHA=nil

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ARG GOOS=linux
ARG GOARCH=amd64

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GOARM=${TARGETVARIANT}

ENV GOPROXY=https://proxy.golang.org


WORKDIR /build
COPY . .

RUN \
  export GOARM=$( echo "${GOARM}" | cut -c2-) && \
  CGO_ENABLED=0 \
    go build \
	-ldflags "-s -w" \
	-o bin/k8s-secret-replicator \
	.

# Use distroless as minimal base image to package the k8s-secret-replicator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /build/bin/k8s-secret-replicator ./k8s-secret-replicator

USER 65532
ENTRYPOINT [ "/k8s-secret-replicator" ]
