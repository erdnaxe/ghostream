# Install dependencies then build ghostream
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache -X https://dl-cdn.alpinelinux.org/alpine/edge/community/ gcc libsrt-dev musl-dev
WORKDIR /code
COPY go.* ./
RUN go mod download && go get github.com/markbates/pkger/cmd/pkger
COPY . .
RUN go generate && go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add --no-cache -X https://dl-cdn.alpinelinux.org/alpine/edge/community/ ffmpeg libsrt
COPY --from=build_base /code/out/ghostream /app/ghostream
WORKDIR /app
# 2112 for monitoring, 8023 for Telnet, 8080 for Web, 9710 for SRT, 10000-11000 (UDP) for WebRTC
EXPOSE 2112 8023 8080 9710/udp 10000-11000/udp
CMD ["/app/ghostream"]
