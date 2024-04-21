# syntax=docker/dockerfile:1

FROM golang:1.22.1-alpine as container-runtime
RUN apk add --no-cache git ca-certificates gcc linux-headers musl-dev
COPY . /src
WORKDIR /src/container-runtime
RUN --mount=type=cache,sharing=locked,id=gomod,target=/go/pkg/mod/cache \
    --mount=type=cache,sharing=locked,id=goroot,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o container-runtime

FROM alpine
RUN apk add  --no-cache curl
COPY --from=container-runtime /src/container-runtime/container-runtime /usr/bin/container-runtime
ENTRYPOINT ["/usr/bin/container-runtime"]