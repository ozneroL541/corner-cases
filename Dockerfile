# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS builder
WORKDIR /app
# Install build dependencies for Fyne.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        pkg-config \
        libgl1-mesa-dev \
        xorg-dev \
        libwayland-dev \
        libxkbcommon-dev && \
    rm -rf /var/lib/apt/lists/*
COPY . .
RUN make build

FROM debian:trixie-slim AS runtime
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libgl1 \
        libglx0 \
        libgl1-mesa-dri \
        libx11-6 \
        libxcursor1 \
        libxrandr2 \
        libxinerama1 \
        libxi6 \
        libxxf86vm1 \
        libxkbcommon0 \
        libwayland-client0 \
        libwayland-cursor0 \
        libwayland-egl1 && \
    rm -rf /var/lib/apt/lists/*
RUN groupadd --system app && \
    useradd --system --gid app --create-home app
WORKDIR /app
COPY --from=builder --chown=app:app /app/bin/app /app/bin/app
USER app
ENTRYPOINT ["/app/bin/app"]
