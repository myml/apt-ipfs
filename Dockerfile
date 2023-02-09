FROM golang:1.17 as builder
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -ldflags="-s -w"

FROM debian:stable-slim
COPY --from=builder /src/apt-ipfs /
WORKDIR /
VOLUME /ipfs
CMD ["./apt-ipfs"]