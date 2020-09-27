# Install dependencies, build libsrt then build ghostream
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache git build-base tcl pkgconfig cmake libressl-dev linux-headers
RUN git clone --depth 1 --branch v1.4.2 https://github.com/Haivision/srt && \
    cd srt && ./configure --enable-apps=OFF && make install && cd .. && rm -rf srt
WORKDIR /code
COPY . .
RUN go mod download
RUN go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add ca-certificates
COPY --from=build_base /code/out/ghostream /app/ghostream
COPY --from=build_base /code/web/static /app/web/static
COPY --from=build_base /code/web/template /app/web/template
EXPOSE 8080
CMD ["/app/ghostream"]
