FROM golang:1.11-alpine AS build

# go buildに最低限のツールとsrcへの移動
RUN apk --no-cache add gcc g++ make ca-certificates git
WORKDIR /go/src/github.com/s-take/go-http-sample

# もしdepがいるなら
# RUN apk update \
#   && apk add --no-cache git \
#   && go get -u github.com/golang/dep/cmd/dep
# COPY Gopkg.lock Gopkg.toml ./
# COPY ../main.go main.go
# RUN  dep ensure

COPY main.go main.go
RUN go get golang.org/x/net/netutil

RUN go install ./...

# 上記でできたバイナリだけをalpineに入れる
FROM alpine:3.10
WORKDIR /usr/bin
COPY --from=build /go/bin .