# Install dependencies then build ghostream
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache build-base gcc
# libsrt is not yet packaged in community repository
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing libsrt-dev
WORKDIR /code
COPY go.* ./
RUN go mod download && go get github.com/markbates/pkger/cmd/pkger
COPY . .
RUN go generate && go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add ca-certificates
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing libsrt
COPY --from=build_base /code/out/ghostream /app/ghostream
WORKDIR /app
# 8080 for Web and Websocket, 2112 for prometheus monitoring and 9710 for SRT
EXPOSE 8080 2112 9710
CMD ["/app/ghostream"]
