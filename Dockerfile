# Install dependencies then build ghostream
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache gcc musl-dev
# libsrt is not yet packaged in community repository
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing libsrt-dev
WORKDIR /code
COPY go.* ./
RUN go mod download && go get github.com/markbates/pkger/cmd/pkger
COPY . .
RUN go generate && go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing libsrt
COPY --from=build_base /code/out/ghostream /app/ghostream
WORKDIR /app
# 9710 for SRT, 8080 for Web, 2112 for monitoring and 10000-10005 (UDP) for WebRTC
EXPOSE 9710 8080 2112 10000-10005/udp
CMD ["/app/ghostream"]
