# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS builder
WORKDIR /app
COPY . /app/
# Install gl dependencies for fyne
RUN apt-get update && \
    apt-get install -y \
    libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev && \
    rm -rf /var/lib/apt/lists/*
# Build the application
RUN make build

RUN ls -la /app/bin

ENTRYPOINT ["sleep", "infinity"]
#FROM alpine:3.22
#RUN addgroup -S app && adduser -S -G app app
#USER app
#WORKDIR /app
#COPY --from=builder /app/bin /app/bin
#
#ENTRYPOINT ["/app/bin/app"]
