# Install dependencies, build libsrt then build ghostream
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache git build-base tcl pkgconfig cmake libressl-dev linux-headers
RUN git clone --depth 1 --branch v1.4.2 https://github.com/Haivision/srt && \
    cd srt && ./configure --enable-apps=OFF && make install && cd .. && rm -rf srt
WORKDIR /code
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add ca-certificates libressl libstdc++ libgcc
COPY --from=build_base /code/out/ghostream /app/ghostream
COPY --from=build_base /code/web/static /app/web/static
COPY --from=build_base /usr/local/lib64/libsrt.so.1 /lib/libsrt.so.1
WORKDIR /app
# 8080 for Web and Websocket, 2112 for prometheus monitoring and 9710 for SRT
EXPOSE 8080 2112 9710
CMD ["/app/ghostream"]
