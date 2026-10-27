# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS builder
WORKDIR /app
COPY . /app/
# Build the application
RUN make build

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
USER app
WORKDIR /app
COPY --from=builder /app/bin /app/bin

ENTRYPOINT ["/app/bin/app"]
