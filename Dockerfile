FROM golang:1.25 as builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/github

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0

COPY --from=builder /app/main /irc-github
CMD ["/irc-github"]
