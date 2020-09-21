# Install dependencies and build
FROM golang:1.15-alpine AS build_base
RUN apk add --no-cache git
WORKDIR /code
COPY . .
RUN go mod download
RUN go build -o ./out/ghostream .

# Production image
FROM alpine:3.12
RUN apk add ca-certificates
COPY --from=build_base /code/out/ghostream /app/ghostream
EXPOSE 8080
CMD ["/app/ghostream"]