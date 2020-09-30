# Install dependencies then build ghostream
FROM debian:bullseye-slim AS build_base
RUN apt-get update && \
        apt-get install -y --no-install-recommends ca-certificates \
        gcc golang libc6-dev libsrt1 libsrt-openssl-dev musl-dev
WORKDIR /code
COPY go.* ./
RUN go mod download && go get github.com/markbates/pkger/cmd/pkger
COPY . .
RUN PATH=/root/go/bin:$PATH go generate && go build -o ./out/ghostream .

# Production image
FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ffmpeg libsrt1 musl
COPY --from=build_base /code/out/ghostream /app/ghostream
WORKDIR /app
# 9710 (UDP) for SRT, 8080 for Web, 2112 for monitoring and 10000-10005 (UDP) for WebRTC
EXPOSE 9710/udp 8080 2112 10000-10005/udp
CMD ["/app/ghostream"]
