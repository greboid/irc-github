FROM golang:1.27.4 as builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/github

FROM alpine:3.24.4 AS secret-env

ARG SECRET_ENV_TAG="v0.1.0"

RUN apk add --no-cache git=2.54.1-r0 rust=1.96.1-r0 cargo=1.96.1-r0 musl-dev=1.2.6-r3
RUN git clone --depth=1 -b $SECRET_ENV_TAG https://git.mouse-lake.ts.net/greboid/secret-env /src
WORKDIR /src
RUN RUSTFLAGS='-C target-feature=+crt-static' cargo build --release --locked

FROM ghcr.io/greboid/dockerbase/nonroot:1.20261105.0

COPY --from=builder /app/main /usr/libexec/irc-github.real
COPY --from=secret-env /src/target/release/secret-env /irc-github
COPY secret-env.conf /etc/secret-env/irc-github.conf
CMD ["/irc-github"]
