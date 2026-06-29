# STAGE 1: Build the static Go binary executable
FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go transcode.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker main.go transcode.go

FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/worker .

RUN mkdir /data

CMD ["./worker"]